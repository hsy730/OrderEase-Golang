package order

import (
	"testing"
	"time"

	"orderease/domain/product"
	"orderease/domain/shared"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockProductFinder is a mock implementation of ProductFinder
type MockProductFinder struct {
	mock.Mock
}

func (m *MockProductFinder) FindProduct(id shared.ID) (*product.Product, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*product.Product), args.Error(1)
}

func (m *MockProductFinder) FindOption(id shared.ID) (*product.ProductOption, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*product.ProductOption), args.Error(1)
}

func (m *MockProductFinder) FindOptionCategory(id shared.ID) (*product.ProductOptionCategory, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*product.ProductOptionCategory), args.Error(1)
}

func TestNewOrderItem(t *testing.T) {
	tests := []struct {
		name        string
		productID   shared.ID
		quantity    int
		price       shared.Price
		options     []OrderItemOption
		wantTotal   shared.Price
		validate    func(*testing.T, OrderItem)
	}{
		{
			name:      "single item without options",
			productID: shared.ID(123),
			quantity:  2,
			price:     shared.Price(100),
			options:   []OrderItemOption{},
			wantTotal: shared.Price(200),
			validate: func(t *testing.T, item OrderItem) {
				assert.Equal(t, shared.ID(0), item.ID)
				assert.Equal(t, shared.ID(123), item.ProductID)
				assert.Equal(t, 2, item.Quantity)
				assert.Equal(t, shared.Price(100), item.Price)
				assert.Equal(t, shared.Price(200), item.TotalPrice)
				assert.Empty(t, item.Options)
			},
		},
		{
			name:      "single item with one option",
			productID: shared.ID(456),
			quantity:  1,
			price:     shared.Price(50),
			options: []OrderItemOption{
				{PriceAdjustment: 10},
			},
			wantTotal: shared.Price(60),
		},
		{
			name:      "single item with multiple options",
			productID: shared.ID(789),
			quantity:  3,
			price:     shared.Price(100),
			options: []OrderItemOption{
				{PriceAdjustment: 10},
				{PriceAdjustment: 5},
			},
			wantTotal: shared.Price(345), // (100 + 10 + 5) * 3
		},
		{
			name:      "item with negative price adjustment",
			productID: shared.ID(111),
			quantity:  2,
			price:     shared.Price(100),
			options: []OrderItemOption{
				{PriceAdjustment: -10},
			},
			wantTotal: shared.Price(180), // (100 - 10) * 2
		},
		{
			name:      "zero quantity",
			productID: shared.ID(222),
			quantity:  0,
			price:     shared.Price(100),
			options:   []OrderItemOption{},
			wantTotal: shared.Price(0),
		},
		{
			name:      "item with zero price adjustment options",
			productID: shared.ID(333),
			quantity:  5,
			price:     shared.Price(20),
			options: []OrderItemOption{
				{PriceAdjustment: 0},
				{PriceAdjustment: 0},
			},
			wantTotal: shared.Price(100), // (20 + 0 + 0) * 5
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewOrderItem(tt.productID, tt.quantity, tt.price, tt.options)

			assert.Equal(t, tt.wantTotal, got.TotalPrice)
			if tt.validate != nil {
				tt.validate(t, got)
			}
		})
	}
}

func TestNewOrderItemOption(t *testing.T) {
	tests := []struct {
		name            string
		categoryID      shared.ID
		optionID        shared.ID
		optionName      string
		categoryName    string
		priceAdjustment float64
		validate        func(*testing.T, OrderItemOption)
	}{
		{
			name:            "valid option",
			categoryID:      shared.ID(1),
			optionID:        shared.ID(2),
			optionName:      "大",
			categoryName:    "尺寸",
			priceAdjustment: 5.0,
			validate: func(t *testing.T, o OrderItemOption) {
				assert.Equal(t, shared.ID(0), o.ID)
				assert.Equal(t, shared.ID(1), o.CategoryID)
				assert.Equal(t, shared.ID(2), o.OptionID)
				assert.Equal(t, "大", o.OptionName)
				assert.Equal(t, "尺寸", o.CategoryName)
				assert.Equal(t, 5.0, o.PriceAdjustment)
				assert.False(t, o.CreatedAt.IsZero())
				assert.False(t, o.UpdatedAt.IsZero())
			},
		},
		{
			name:            "negative price adjustment",
			categoryID:      shared.ID(3),
			optionID:        shared.ID(4),
			optionName:      "小",
			categoryName:    "尺寸",
			priceAdjustment: -2.5,
		},
		{
			name:            "zero price adjustment",
			categoryID:      shared.ID(5),
			optionID:        shared.ID(6),
			optionName:      "中",
			categoryName:    "尺寸",
			priceAdjustment: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			before := time.Now()
			got := NewOrderItemOption(tt.categoryID, tt.optionID, tt.optionName, tt.categoryName, tt.priceAdjustment)
			after := time.Now()

			assert.Equal(t, tt.categoryID, got.CategoryID)
			assert.Equal(t, tt.optionID, got.OptionID)
			assert.Equal(t, tt.optionName, got.OptionName)
			assert.Equal(t, tt.categoryName, got.CategoryName)
			assert.Equal(t, tt.priceAdjustment, got.PriceAdjustment)
			assert.True(t, got.CreatedAt.After(before) || got.CreatedAt.Equal(before))
			assert.True(t, got.CreatedAt.Before(after) || got.CreatedAt.Equal(after))
			assert.True(t, got.UpdatedAt.After(before) || got.UpdatedAt.Equal(before))
			assert.True(t, got.UpdatedAt.Before(after) || got.UpdatedAt.Equal(after))

			if tt.validate != nil {
				tt.validate(t, got)
			}
		})
	}
}

func TestOrderItem_CalculatePrice(t *testing.T) {
	tests := []struct {
		name         string
		orderItem    OrderItem
		setupMock    func(*MockProductFinder)
		wantPrice    shared.Price
		wantErr      bool
		errContains  string
		validateItem func(*testing.T, *OrderItem)
	}{
		{
			name: "item without options",
			orderItem: OrderItem{
				ProductID: shared.ID(123),
				Quantity:  2,
				Price:     shared.Price(100),
				Options:   []OrderItemOption{},
			},
			setupMock: func(m *MockProductFinder) {
				// No options, no mock calls needed
			},
			wantPrice: shared.Price(200),
			wantErr:   false,
		},
		{
			name: "item with options",
			orderItem: OrderItem{
				ProductID: shared.ID(123),
				Quantity:  2,
				Price:     shared.Price(100),
				Options: []OrderItemOption{
					{OptionID: shared.ID(1)},
					{OptionID: shared.ID(2)},
				},
			},
			setupMock: func(m *MockProductFinder) {
				m.On("FindOption", shared.ID(1)).Return(&product.ProductOption{
					ID:           shared.ID(1),
					CategoryID:   shared.ID(10),
					Name:         "大",
					PriceAdjustment: 10,
				}, nil)
				m.On("FindOptionCategory", shared.ID(10)).Return(&product.ProductOptionCategory{
					ID:   shared.ID(10),
					Name: "尺寸",
				}, nil)
				m.On("FindOption", shared.ID(2)).Return(&product.ProductOption{
					ID:           shared.ID(2),
					CategoryID:   shared.ID(20),
					Name:         "加冰",
					PriceAdjustment: 5,
				}, nil)
				m.On("FindOptionCategory", shared.ID(20)).Return(&product.ProductOptionCategory{
					ID:   shared.ID(20),
					Name: "温度",
				}, nil)
			},
			wantPrice: shared.Price(230), // (100 + 10 + 5) * 2
			wantErr:   false,
			validateItem: func(t *testing.T, item *OrderItem) {
				assert.Equal(t, "大", item.Options[0].OptionName)
				assert.Equal(t, "尺寸", item.Options[0].CategoryName)
				assert.Equal(t, 10.0, item.Options[0].PriceAdjustment)
				assert.Equal(t, "加冰", item.Options[1].OptionName)
				assert.Equal(t, "温度", item.Options[1].CategoryName)
				assert.Equal(t, 5.0, item.Options[1].PriceAdjustment)
			},
		},
		{
			name: "option not found",
			orderItem: OrderItem{
				ProductID: shared.ID(123),
				Quantity:  2,
				Price:     shared.Price(100),
				Options: []OrderItemOption{
					{OptionID: shared.ID(999)},
				},
			},
			setupMock: func(m *MockProductFinder) {
				m.On("FindOption", shared.ID(999)).Return(nil, assert.AnError)
			},
			wantErr:     true,
			errContains: "商品参数选项不存在",
		},
		{
			name: "category not found",
			orderItem: OrderItem{
				ProductID: shared.ID(123),
				Quantity:  2,
				Price:     shared.Price(100),
				Options: []OrderItemOption{
					{OptionID: shared.ID(1)},
				},
			},
			setupMock: func(m *MockProductFinder) {
				m.On("FindOption", shared.ID(1)).Return(&product.ProductOption{
					ID:           shared.ID(1),
					CategoryID:   shared.ID(999),
					Name:         "大",
					PriceAdjustment: 10,
				}, nil)
				m.On("FindOptionCategory", shared.ID(999)).Return(nil, assert.AnError)
			},
			wantErr:     true,
			errContains: "商品参数类别不存在",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockFinder := new(MockProductFinder)
			tt.setupMock(mockFinder)

			item := tt.orderItem
			gotPrice, err := item.CalculatePrice(mockFinder)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantPrice, gotPrice)
				assert.Equal(t, tt.wantPrice, item.TotalPrice)
				if tt.validateItem != nil {
					tt.validateItem(t, &item)
				}
			}

			mockFinder.AssertExpectations(t)
		})
	}
}

func TestOrderItem_ValidateProduct(t *testing.T) {
	tests := []struct {
		name        string
		orderItem   OrderItem
		product     *product.Product
		shopID      uint64
		wantErr     bool
		errContains string
	}{
		{
			name: "valid product",
			orderItem: OrderItem{
				ProductID: shared.ID(123),
				Quantity:  2,
			},
			product: &product.Product{
				ID:     shared.ID(123),
				ShopID: 456,
				Name:   "测试商品",
				Stock:  10,
			},
			shopID:  456,
			wantErr: false,
		},
		{
			name: "shop mismatch",
			orderItem: OrderItem{
				ProductID: shared.ID(123),
				Quantity:  2,
			},
			product: &product.Product{
				ID:     shared.ID(123),
				ShopID: 789,
				Name:   "测试商品",
				Stock:  10,
			},
			shopID:      456,
			wantErr:     true,
			errContains: "商品不属于该店铺",
		},
		{
			name: "insufficient stock",
			orderItem: OrderItem{
				ProductID: shared.ID(123),
				Quantity:  10,
			},
			product: &product.Product{
				ID:     shared.ID(123),
				ShopID: 456,
				Name:   "测试商品",
				Stock:  5,
			},
			shopID:      456,
			wantErr:     true,
			errContains: "库存不足",
		},
		{
			name: "exact stock match",
			orderItem: OrderItem{
				ProductID: shared.ID(123),
				Quantity:  5,
			},
			product: &product.Product{
				ID:     shared.ID(123),
				ShopID: 456,
				Name:   "测试商品",
				Stock:  5,
			},
			shopID:  456,
			wantErr: false,
		},
		{
			name: "zero quantity",
			orderItem: OrderItem{
				ProductID: shared.ID(123),
				Quantity:  0,
			},
			product: &product.Product{
				ID:     shared.ID(123),
				ShopID: 456,
				Name:   "测试商品",
				Stock:  0,
			},
			shopID:  456,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.orderItem.ValidateProduct(tt.product, tt.shopID)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestOrderItem_SetProductSnapshot(t *testing.T) {
	tests := []struct {
		name     string
		product  *product.Product
		validate func(*testing.T, *OrderItem)
	}{
		{
			name: "set full snapshot",
			product: &product.Product{
				Name:        "测试商品",
				Description: "这是一个测试商品",
				ImageURL:    "http://example.com/image.jpg",
				Price:       shared.Price(99.99),
			},
			validate: func(t *testing.T, item *OrderItem) {
				assert.Equal(t, "测试商品", item.ProductName)
				assert.Equal(t, "这是一个测试商品", item.ProductDescription)
				assert.Equal(t, "http://example.com/image.jpg", item.ProductImageURL)
				assert.Equal(t, shared.Price(99.99), item.Price)
			},
		},
		{
			name: "set snapshot with empty fields",
			product: &product.Product{
				Name:        "",
				Description: "",
				ImageURL:    "",
				Price:       shared.Price(0),
			},
			validate: func(t *testing.T, item *OrderItem) {
				assert.Equal(t, "", item.ProductName)
				assert.Equal(t, "", item.ProductDescription)
				assert.Equal(t, "", item.ProductImageURL)
				assert.Equal(t, shared.Price(0), item.Price)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			item := &OrderItem{}
			item.SetProductSnapshot(tt.product)

			if tt.validate != nil {
				tt.validate(t, item)
			}
		})
	}
}

func TestOrderItem_OptionsSnapshot(t *testing.T) {
	t.Run("options preserve timestamps", func(t *testing.T) {
		before := time.Now()
		option := NewOrderItemOption(
			shared.ID(1),
			shared.ID(2),
			"大",
			"尺寸",
			10.0,
		)
		after := time.Now()

		assert.True(t, option.CreatedAt.After(before) || option.CreatedAt.Equal(before))
		assert.True(t, option.CreatedAt.Before(after) || option.CreatedAt.Equal(after))
		assert.True(t, option.UpdatedAt.After(before) || option.UpdatedAt.Equal(before))
		assert.True(t, option.UpdatedAt.Before(after) || option.UpdatedAt.Equal(after))
	})
}
