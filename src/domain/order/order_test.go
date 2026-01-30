package order

import (
	"testing"
	"time"

	"github.com/bwmarrin/snowflake"
	"github.com/stretchr/testify/assert"
	"orderease/domain/shared/value_objects"
	"orderease/models"
)

func TestNewOrder(t *testing.T) {
	userID := snowflake.ID(123)
	shopID := snowflake.ID(456)

	order := NewOrder(userID, shopID)

	assert.NotNil(t, order)
	assert.Equal(t, userID, order.UserID())
	assert.Equal(t, shopID, order.ShopID())
	assert.Equal(t, value_objects.OrderStatusPending, order.Status())
	assert.Equal(t, models.Price(0), order.TotalPrice())
	assert.Equal(t, "", order.Remark())
	assert.Empty(t, order.Items())
	assert.False(t, order.createdAt.IsZero())
}

func TestOrder_Getters(t *testing.T) {
	userID := snowflake.ID(123)
	shopID := snowflake.ID(456)
	order := NewOrder(userID, shopID)
	testID := snowflake.ID(999)
	order.SetID(testID)

	assert.Equal(t, testID, order.ID())
	assert.Equal(t, userID, order.UserID())
	assert.Equal(t, shopID, order.ShopID())
	assert.Equal(t, models.Price(0), order.TotalPrice())
	assert.Equal(t, value_objects.OrderStatusPending, order.Status())
	assert.Equal(t, "", order.Remark())
	assert.Empty(t, order.Items())
}

func TestOrder_Setters(t *testing.T) {
	order := NewOrder(123, 456)

	// Test SetTotalPrice
	newPrice := models.Price(10000)
	order.SetTotalPrice(newPrice)
	assert.Equal(t, newPrice, order.TotalPrice())

	// Test SetStatus
	newStatus := value_objects.OrderStatusAccepted
	order.SetStatus(newStatus)
	assert.Equal(t, newStatus, order.Status())

	// Test SetRemark
	remark := "测试备注"
	order.SetRemark(remark)
	assert.Equal(t, remark, order.Remark())

	// Test SetItems
	items := []OrderItem{*NewOrderItem(111, 2, 5000)}
	order.SetItems(items)
	assert.Len(t, order.Items(), 1)

	// Test SetCreatedAt
	now := time.Now()
	order.SetCreatedAt(now)
	assert.Equal(t, now, order.CreatedAt())

	// Test SetUpdatedAt
	updated := time.Now()
	order.SetUpdatedAt(updated)
	assert.Equal(t, updated, order.UpdatedAt())
}

func TestOrder_AddItem(t *testing.T) {
	order := NewOrder(123, 456)
	item := NewOrderItem(111, 2, 5000)

	order.AddItem(*item)

	assert.Len(t, order.Items(), 1)
	assert.Equal(t, item.productID, order.Items()[0].productID)
}

func TestOrder_AddMultipleItems(t *testing.T) {
	order := NewOrder(123, 456)
	item1 := NewOrderItem(111, 2, 5000)
	item2 := NewOrderItem(222, 1, 10000)

	order.AddItem(*item1)
	order.AddItem(*item2)

	assert.Len(t, order.Items(), 2)
}

func TestOrder_ValidateItems_EmptyItems(t *testing.T) {
	order := NewOrder(123, 456)

	err := order.ValidateItems()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "订单项不能为空")
}

func TestOrder_ValidateItems_InvalidProductID(t *testing.T) {
	order := NewOrder(123, 456)
	item := NewOrderItem(0, 2, 5000) // productID = 0
	order.AddItem(*item)

	err := order.ValidateItems()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "商品ID不能为空")
}

func TestOrder_ValidateItems_InvalidQuantity(t *testing.T) {
	order := NewOrder(123, 456)
	item := NewOrderItem(111, 0, 5000) // quantity = 0
	order.AddItem(*item)

	err := order.ValidateItems()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "商品数量必须大于0")
}

func TestOrder_ValidateItems_NegativeQuantity(t *testing.T) {
	order := NewOrder(123, 456)
	item := NewOrderItem(111, -1, 5000) // quantity = -1
	order.AddItem(*item)

	err := order.ValidateItems()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "商品数量必须大于0")
}

func TestOrder_ValidateItems_ValidItems(t *testing.T) {
	order := NewOrder(123, 456)
	item := NewOrderItem(111, 2, 5000)
	order.AddItem(*item)

	err := order.ValidateItems()

	assert.NoError(t, err)
}

func TestOrder_CalculateTotal_EmptyItems(t *testing.T) {
	order := NewOrder(123, 456)

	total := order.CalculateTotal()

	assert.Equal(t, models.Price(0), total)
}

func TestOrder_CalculateTotal_SingleItem(t *testing.T) {
	order := NewOrder(123, 456)
	item := NewOrderItem(111, 2, 5000) // 2 * 5000 = 10000
	order.AddItem(*item)

	total := order.CalculateTotal()

	assert.Equal(t, models.Price(10000), total)
}

func TestOrder_CalculateTotal_MultipleItems(t *testing.T) {
	order := NewOrder(123, 456)
	item1 := NewOrderItem(111, 2, 5000)  // 2 * 5000 = 10000
	item2 := NewOrderItem(222, 1, 3000)  // 1 * 3000 = 3000
	order.AddItem(*item1)
	order.AddItem(*item2)

	total := order.CalculateTotal()

	assert.Equal(t, models.Price(13000), total)
}

func TestOrder_CalculateTotal_WithOptions(t *testing.T) {
	order := NewOrder(123, 456)
	item := NewOrderItem(111, 2, 5000) // base: 2 * 5000 = 10000

	// Add option with price adjustment
	option := OrderItemOption{
		PriceAdjustment: 500, // 2 * 500 = 1000
	}
	item.AddOption(option)

	order.AddItem(*item)

	total := order.CalculateTotal()

	assert.Equal(t, models.Price(11000), total)
}

func TestOrder_CalculateTotal_Complex(t *testing.T) {
	order := NewOrder(123, 456)

	// Item 1: 2 * 5000 = 10000, option: 2 * 200 = 400 => 10400
	item1 := NewOrderItem(111, 2, 5000)
	item1.AddOption(OrderItemOption{PriceAdjustment: 200})

	// Item 2: 1 * 8000 = 8000, options: 1 * 300 + 1 * 150 = 450 => 8450
	item2 := NewOrderItem(222, 1, 8000)
	item2.AddOption(OrderItemOption{PriceAdjustment: 300})
	item2.AddOption(OrderItemOption{PriceAdjustment: 150})

	order.AddItem(*item1)
	order.AddItem(*item2)

	total := order.CalculateTotal()

	assert.Equal(t, models.Price(18850), total)
}

func TestOrder_CanTransitionTo(t *testing.T) {
	order := NewOrder(123, 456)
	order.SetStatus(value_objects.OrderStatusPending)

	tests := []struct {
		name     string
		to       value_objects.OrderStatus
		allowed  bool
	}{
		{
			name:    "Pending to Accepted - allowed",
			to:      value_objects.OrderStatusAccepted,
			allowed: true,
		},
		{
			name:    "Pending to Canceled - allowed",
			to:      value_objects.OrderStatusCanceled,
			allowed: true,
		},
		{
			name:    "Pending to Rejected - allowed",
			to:      value_objects.OrderStatusRejected,
			allowed: true,
		},
		{
			name:    "Pending to Complete - not allowed",
			to:      value_objects.OrderStatusComplete,
			allowed: false,
		},
		{
			name:    "Pending to Pending - not allowed",
			to:      value_objects.OrderStatusPending,
			allowed: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := order.CanTransitionTo(tt.to)
			assert.Equal(t, tt.allowed, got)
		})
	}
}

func TestOrder_IsFinal(t *testing.T) {
	tests := []struct {
		name      string
		status    value_objects.OrderStatus
		wantFinal bool
	}{
		{
			name:      "Pending - not final",
			status:    value_objects.OrderStatusPending,
			wantFinal: false,
		},
		{
			name:      "Accepted - not final",
			status:    value_objects.OrderStatusAccepted,
			wantFinal: false,
		},
		{
			name:      "Complete - final",
			status:    value_objects.OrderStatusComplete,
			wantFinal: true,
		},
		{
			name:      "Canceled - final",
			status:    value_objects.OrderStatusCanceled,
			wantFinal: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			order := NewOrder(123, 456)
			order.SetStatus(tt.status)

			got := order.IsFinal()
			assert.Equal(t, tt.wantFinal, got)
		})
	}
}

func TestOrder_GetItemCount(t *testing.T) {
	order := NewOrder(123, 456)

	assert.Equal(t, 0, order.GetItemCount())

	order.AddItem(*NewOrderItem(111, 1, 1000))
	assert.Equal(t, 1, order.GetItemCount())

	order.AddItem(*NewOrderItem(222, 1, 2000))
	assert.Equal(t, 2, order.GetItemCount())
}

func TestOrder_GetTotalQuantity(t *testing.T) {
	order := NewOrder(123, 456)

	assert.Equal(t, 0, order.GetTotalQuantity())

	item1 := NewOrderItem(111, 2, 5000)
	item2 := NewOrderItem(222, 3, 3000)
	order.AddItem(*item1)
	order.AddItem(*item2)

	assert.Equal(t, 5, order.GetTotalQuantity())
}

func TestOrder_IsPending(t *testing.T) {
	order := NewOrder(123, 456)

	assert.True(t, order.IsPending())

	order.SetStatus(value_objects.OrderStatusAccepted)
	assert.False(t, order.IsPending())
}

func TestOrder_CanBeDeleted(t *testing.T) {
	tests := []struct {
		name         string
		status       value_objects.OrderStatus
		canBeDeleted bool
	}{
		{
			name:         "Pending - can delete",
			status:       value_objects.OrderStatusPending,
			canBeDeleted: true,
		},
		{
			name:         "Accepted - can delete",
			status:       value_objects.OrderStatusAccepted,
			canBeDeleted: true,
		},
		{
			name:         "Shipped - can delete",
			status:       value_objects.OrderStatusShipped,
			canBeDeleted: true,
		},
		{
			name:         "Complete - cannot delete",
			status:       value_objects.OrderStatusComplete,
			canBeDeleted: false,
		},
		{
			name:         "Canceled - cannot delete",
			status:       value_objects.OrderStatusCanceled,
			canBeDeleted: false,
		},
		{
			name:         "Rejected - can delete",
			status:       value_objects.OrderStatusRejected,
			canBeDeleted: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			order := NewOrder(123, 456)
			order.SetStatus(tt.status)

			got := order.CanBeDeleted()
			assert.Equal(t, tt.canBeDeleted, got)
		})
	}
}

func TestOrder_HasItems(t *testing.T) {
	order := NewOrder(123, 456)

	assert.False(t, order.HasItems())

	order.AddItem(*NewOrderItem(111, 1, 1000))
	assert.True(t, order.HasItems())
}

func TestOrder_IsEmpty(t *testing.T) {
	order := NewOrder(123, 456)

	assert.True(t, order.IsEmpty())

	order.AddItem(*NewOrderItem(111, 1, 1000))
	assert.False(t, order.IsEmpty())
}

func TestOrder_ToModel(t *testing.T) {
	userID := snowflake.ID(123)
	shopID := snowflake.ID(456)
	orderID := snowflake.ID(789)

	order := NewOrder(userID, shopID)
	order.SetID(orderID)
	order.SetTotalPrice(10000)
	order.SetRemark("测试备注")
	order.SetStatus(value_objects.OrderStatusAccepted)

	item := NewOrderItem(111, 2, 5000)
	order.AddItem(*item)

	model := order.ToModel()

	assert.Equal(t, orderID, model.ID)
	assert.Equal(t, userID, model.UserID)
	assert.Equal(t, shopID, model.ShopID)
	assert.Equal(t, models.Price(10000), model.TotalPrice)
	assert.Equal(t, int(value_objects.OrderStatusAccepted), model.Status)
	assert.Equal(t, "测试备注", model.Remark)
	assert.Len(t, model.Items, 1)
}

func TestOrder_OrderFromModel(t *testing.T) {
	userID := snowflake.ID(123)
	shopID := snowflake.ID(456)
	orderID := snowflake.ID(789)
	productID := snowflake.ID(999)

	model := &models.Order{
		ID:         orderID,
		UserID:     userID,
		ShopID:     shopID,
		TotalPrice: 10000,
		Status:     int(value_objects.OrderStatusAccepted),
		Remark:     "测试备注",
		Items: []models.OrderItem{
			{
				OrderID:  orderID,
				ProductID: productID,
				Quantity:  2,
				Price:     5000,
				Options:   []models.OrderItemOption{},
			},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	order := OrderFromModel(model)

	assert.Equal(t, orderID, order.ID())
	assert.Equal(t, userID, order.UserID())
	assert.Equal(t, shopID, order.ShopID())
	assert.Equal(t, models.Price(10000), order.TotalPrice())
	assert.Equal(t, value_objects.OrderStatusAccepted, order.Status())
	assert.Equal(t, "测试备注", order.Remark())
	assert.Len(t, order.Items(), 1)
}

func TestOrder_ToCreateOrderRequest(t *testing.T) {
	order := NewOrder(123, 456)
	order.SetID(789)
	order.SetRemark("测试备注")

	item := NewOrderItem(111, 2, 5000)
	item.AddOption(OrderItemOption{
		CategoryID: 1,
		OptionID:   2,
	})
	order.AddItem(*item)

	request := order.ToCreateOrderRequest()

	assert.Equal(t, snowflake.ID(789), request.ID)
	assert.Equal(t, snowflake.ID(123), request.UserID)
	assert.Equal(t, snowflake.ID(456), request.ShopID)
	assert.Equal(t, "测试备注", request.Remark)
	assert.Equal(t, int(value_objects.OrderStatusPending), request.Status)
	assert.Len(t, request.Items, 1)
	assert.Equal(t, snowflake.ID(111), request.Items[0].ProductID)
	assert.Equal(t, 2, request.Items[0].Quantity)
	assert.Len(t, request.Items[0].Options, 1)
}

func TestToOrderElements(t *testing.T) {
	userID := snowflake.ID(123)
	shopID := snowflake.ID(456)
	orderID := snowflake.ID(789)

	orders := []models.Order{
		{
			ID:         orderID,
			UserID:     userID,
			ShopID:     shopID,
			TotalPrice: 10000,
			Status:     int(value_objects.OrderStatusAccepted),
			Remark:     "测试备注",
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		},
	}

	elements := ToOrderElements(orders)

	assert.Len(t, elements, 1)
	assert.Equal(t, orderID, elements[0].ID)
	assert.Equal(t, userID, elements[0].UserID)
	assert.Equal(t, shopID, elements[0].ShopID)
	assert.Equal(t, models.Price(10000), elements[0].TotalPrice)
	assert.Equal(t, int(value_objects.OrderStatusAccepted), elements[0].Status)
}

// ==================== OrderItem Tests ====================

func TestNewOrderItem(t *testing.T) {
	productID := snowflake.ID(111)
	quantity := 2
	price := models.Price(5000)

	item := NewOrderItem(productID, quantity, price)

	assert.NotNil(t, item)
	assert.Equal(t, productID, item.ProductID())
	assert.Equal(t, quantity, item.Quantity())
	assert.Equal(t, price, item.Price())
	assert.Equal(t, models.Price(10000), item.TotalPrice()) // 2 * 5000
}

func TestOrderItem_Getters(t *testing.T) {
	item := NewOrderItem(111, 2, 5000)

	assert.Equal(t, snowflake.ID(0), item.ID())
	assert.Equal(t, snowflake.ID(111), item.ProductID())
	assert.Equal(t, 2, item.Quantity())
	assert.Equal(t, models.Price(5000), item.Price())
	assert.Equal(t, models.Price(10000), item.TotalPrice())
	assert.Equal(t, "", item.ProductName())
	assert.Equal(t, "", item.ProductDescription())
	assert.Equal(t, "", item.ProductImageURL())
	assert.Empty(t, item.Options())
}

func TestOrderItem_Setters(t *testing.T) {
	item := NewOrderItem(111, 2, 5000)

	// Test SetProductName
	item.SetProductName("测试商品")
	assert.Equal(t, "测试商品", item.ProductName())

	// Test SetProductDescription
	item.SetProductDescription("这是测试商品描述")
	assert.Equal(t, "这是测试商品描述", item.ProductDescription())

	// Test SetProductImageURL
	item.SetProductImageURL("http://example.com/image.jpg")
	assert.Equal(t, "http://example.com/image.jpg", item.ProductImageURL())

	// Test SetTotalPrice
	newTotal := models.Price(15000)
	item.SetTotalPrice(newTotal)
	assert.Equal(t, newTotal, item.TotalPrice())

	// Test SetPrice
	newPrice := models.Price(8000)
	item.SetPrice(newPrice)
	assert.Equal(t, newPrice, item.Price())
}

func TestOrderItem_AddOption(t *testing.T) {
	item := NewOrderItem(111, 2, 5000)

	assert.Empty(t, item.Options())

	option := OrderItemOption{
		CategoryID:      1,
		OptionID:        2,
		OptionName:      "大杯",
		CategoryName:    "杯型",
		PriceAdjustment: 500,
	}

	item.AddOption(option)

	assert.Len(t, item.Options(), 1)
	assert.Equal(t, option, item.Options()[0])
}

func TestOrderItem_AddMultipleOptions(t *testing.T) {
	item := NewOrderItem(111, 2, 5000)

	item.AddOption(OrderItemOption{CategoryID: 1, OptionID: 2})
	item.AddOption(OrderItemOption{CategoryID: 3, OptionID: 4})

	assert.Len(t, item.Options(), 2)
}

func TestOrderItem_ToModel(t *testing.T) {
	orderID := snowflake.ID(123)
	productID := snowflake.ID(456)
	item := NewOrderItem(productID, 2, 5000)

	item.SetProductName("测试商品")
	item.SetProductDescription("测试描述")
	item.SetProductImageURL("http://example.com/image.jpg")

	option := OrderItemOption{
		ID:              snowflake.ID(789),
		OrderItemID:     snowflake.ID(0),
		CategoryID:      snowflake.ID(1),
		OptionID:        snowflake.ID(2),
		OptionName:      "大杯",
		CategoryName:    "杯型",
		PriceAdjustment: 500,
	}
	item.AddOption(option)

	model := item.ToModel(orderID)

	assert.Equal(t, productID, model.ProductID)
	assert.Equal(t, 2, model.Quantity)
	assert.Equal(t, models.Price(5000), model.Price)
	assert.Equal(t, models.Price(10000), model.TotalPrice)
	assert.Equal(t, "测试商品", model.ProductName)
	assert.Equal(t, "测试描述", model.ProductDescription)
	assert.Equal(t, "http://example.com/image.jpg", model.ProductImageURL)
	assert.Len(t, model.Options, 1)
}

func TestOrderItemFromModel(t *testing.T) {
	orderID := snowflake.ID(123)
	itemID := snowflake.ID(456)
	productID := snowflake.ID(789)
	categoryID := snowflake.ID(1)
	optionID := snowflake.ID(2)

	model := &models.OrderItem{
		ID:                 itemID,
		OrderID:            orderID,
		ProductID:          productID,
		Quantity:           3,
		Price:              6000,
		TotalPrice:         18000,
		ProductName:        "测试商品",
		ProductDescription: "测试描述",
		ProductImageURL:    "http://example.com/image.jpg",
		Options: []models.OrderItemOption{
			{
				ID:              snowflake.ID(999),
				OrderItemID:     itemID,
				CategoryID:      categoryID,
				OptionID:        optionID,
				OptionName:      "大杯",
				CategoryName:    "杯型",
				PriceAdjustment: 500,
			},
		},
	}

	item := OrderItemFromModel(model)

	assert.Equal(t, itemID, item.ID())
	assert.Equal(t, productID, item.ProductID())
	assert.Equal(t, 3, item.Quantity())
	assert.Equal(t, models.Price(6000), item.Price())
	assert.Equal(t, models.Price(18000), item.TotalPrice())
	assert.Equal(t, "测试商品", item.ProductName())
	assert.Equal(t, "测试描述", item.ProductDescription())
	assert.Equal(t, "http://example.com/image.jpg", item.ProductImageURL())
	assert.Len(t, item.Options(), 1)
}
