package models

import (
	"testing"
	"time"

	"github.com/bwmarrin/snowflake"
	"github.com/stretchr/testify/assert"
)

func TestOrderStatusConstants(t *testing.T) {
	t.Run("pending status", func(t *testing.T) {
		assert.Equal(t, 1, OrderStatusPending)
	})

	t.Run("accepted status", func(t *testing.T) {
		assert.Equal(t, 2, OrderStatusAccepted)
	})

	t.Run("rejected status", func(t *testing.T) {
		assert.Equal(t, 3, OrderStatusRejected)
	})

	t.Run("shipped status", func(t *testing.T) {
		assert.Equal(t, 4, OrderStatusShipped)
	})

	t.Run("complete status", func(t *testing.T) {
		assert.Equal(t, 10, OrderStatusComplete)
	})

	t.Run("canceled status", func(t *testing.T) {
		assert.Equal(t, -1, OrderStatusCanceled)
	})
}

func TestOrderStatusTransitions(t *testing.T) {
	tests := []struct {
		name         string
		fromStatus   int
		expectedNext int
	}{
		{"pending -> accepted", OrderStatusPending, OrderStatusAccepted},
		{"accepted -> shipped", OrderStatusAccepted, OrderStatusShipped},
		{"shipped -> complete", OrderStatusShipped, OrderStatusComplete},
		{"rejected -> rejected", OrderStatusRejected, OrderStatusRejected},
		{"complete -> complete", OrderStatusComplete, OrderStatusComplete},
		{"canceled -> canceled", OrderStatusCanceled, OrderStatusCanceled},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nextStatus, exists := OrderStatusTransitions[tt.fromStatus]
			assert.True(t, exists)
			assert.Equal(t, tt.expectedNext, nextStatus)
		})
	}
}

func TestDefaultOrderStatusFlow(t *testing.T) {
	assert.NotEmpty(t, DefaultOrderStatusFlow)
	assert.Contains(t, DefaultOrderStatusFlow, "statuses")
	assert.Contains(t, DefaultOrderStatusFlow, "待处理")
	assert.Contains(t, DefaultOrderStatusFlow, "已接单")
	assert.Contains(t, DefaultOrderStatusFlow, "已完成")
	assert.Contains(t, DefaultOrderStatusFlow, "已取消")
}

func TestOrderStruct(t *testing.T) {
	now := time.Now()
	order := Order{
		ID:         snowflake.ID(123),
		UserID:     snowflake.ID(456),
		ShopID:     snowflake.ID(789),
		TotalPrice: 99.99,
		Status:     OrderStatusPending,
		Remark:     "Test order",
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	assert.Equal(t, snowflake.ID(123), order.ID)
	assert.Equal(t, snowflake.ID(456), order.UserID)
	assert.Equal(t, snowflake.ID(789), order.ShopID)
	assert.Equal(t, Price(99.99), order.TotalPrice)
	assert.Equal(t, OrderStatusPending, order.Status)
	assert.Equal(t, "Test order", order.Remark)
}

func TestOrderStructWithItems(t *testing.T) {
	item := OrderItem{
		ID:         snowflake.ID(200),
		OrderID:    snowflake.ID(123),
		ProductID:  snowflake.ID(456),
		Quantity:   2,
		Price:      49.99,
		TotalPrice: 99.98,
	}

	order := Order{
		ID:    snowflake.ID(123),
		Items: []OrderItem{item},
	}

	assert.Len(t, order.Items, 1)
	assert.Equal(t, snowflake.ID(200), order.Items[0].ID)
	assert.Equal(t, 2, order.Items[0].Quantity)
}

func TestOrderItemStruct(t *testing.T) {
	item := OrderItem{
		ID:                 snowflake.ID(123),
		OrderID:            snowflake.ID(456),
		ProductID:          snowflake.ID(789),
		Quantity:           3,
		Price:              29.99,
		TotalPrice:         89.97,
		ProductName:        "Test Product",
		ProductDescription: "Description",
		ProductImageURL:    "https://example.com/image.jpg",
	}

	assert.Equal(t, snowflake.ID(123), item.ID)
	assert.Equal(t, snowflake.ID(456), item.OrderID)
	assert.Equal(t, snowflake.ID(789), item.ProductID)
	assert.Equal(t, 3, item.Quantity)
	assert.Equal(t, Price(29.99), item.Price)
	assert.Equal(t, Price(89.97), item.TotalPrice)
	assert.Equal(t, "Test Product", item.ProductName)
	assert.Equal(t, "Description", item.ProductDescription)
	assert.Equal(t, "https://example.com/image.jpg", item.ProductImageURL)
}

func TestOrderItemStructWithOptions(t *testing.T) {
	option := OrderItemOption{
		ID:          snowflake.ID(100),
		OrderItemID: snowflake.ID(123),
		CategoryID:  snowflake.ID(456),
		OptionID:    snowflake.ID(789),
		OptionName:  "Red",
		CategoryName: "Color",
		PriceAdjustment: 10.0,
	}

	item := OrderItem{
		ID:      snowflake.ID(123),
		Options: []OrderItemOption{option},
	}

	assert.Len(t, item.Options, 1)
	assert.Equal(t, "Red", item.Options[0].OptionName)
	assert.Equal(t, "Color", item.Options[0].CategoryName)
	assert.Equal(t, 10.0, item.Options[0].PriceAdjustment)
}

func TestOrderItemOptionStruct(t *testing.T) {
	now := time.Now()
	option := OrderItemOption{
		ID:              snowflake.ID(123),
		OrderItemID:     snowflake.ID(456),
		CategoryID:      snowflake.ID(789),
		OptionID:        snowflake.ID(999),
		OptionName:      "Large",
		CategoryName:    "Size",
		PriceAdjustment: 5.0,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	assert.Equal(t, snowflake.ID(123), option.ID)
	assert.Equal(t, snowflake.ID(456), option.OrderItemID)
	assert.Equal(t, snowflake.ID(789), option.CategoryID)
	assert.Equal(t, snowflake.ID(999), option.OptionID)
	assert.Equal(t, "Large", option.OptionName)
	assert.Equal(t, "Size", option.CategoryName)
	assert.Equal(t, 5.0, option.PriceAdjustment)
}

func TestOrderStatusLogStruct(t *testing.T) {
	now := time.Now()
	log := OrderStatusLog{
		ID:          snowflake.ID(123),
		OrderID:     snowflake.ID(456),
		OldStatus:   OrderStatusPending,
		NewStatus:   OrderStatusAccepted,
		ChangedTime: now,
	}

	assert.Equal(t, snowflake.ID(123), log.ID)
	assert.Equal(t, snowflake.ID(456), log.OrderID)
	assert.Equal(t, OrderStatusPending, log.OldStatus)
	assert.Equal(t, OrderStatusAccepted, log.NewStatus)
	assert.NotNil(t, log.ChangedTime)
}

func TestOrderElementStruct(t *testing.T) {
	now := time.Now()
	element := OrderElement{
		ID:         snowflake.ID(123),
		UserID:     snowflake.ID(456),
		ShopID:     snowflake.ID(789),
		TotalPrice: 99.99,
		Status:     OrderStatusPending,
		Remark:     "Test element",
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	assert.Equal(t, snowflake.ID(123), element.ID)
	assert.Equal(t, snowflake.ID(456), element.UserID)
	assert.Equal(t, snowflake.ID(789), element.ShopID)
	assert.Equal(t, Price(99.99), element.TotalPrice)
	assert.Equal(t, OrderStatusPending, element.Status)
	assert.Equal(t, "Test element", element.Remark)
}
