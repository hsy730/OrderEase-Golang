package order

import (
	"errors"
	"testing"

	"orderease/domain/product"
	"orderease/domain/shared"

	"github.com/stretchr/testify/assert"
)

func TestOrderDomainService_ValidateOrderCreation(t *testing.T) {
	tests := []struct {
		name        string
		order       func() *Order
		setupMock   func(*MockProductFinder)
		wantErr     bool
		errContains string
		validate    func(*testing.T, *Order)
	}{
		{
			name: "valid order creation",
			order: func() *Order {
				return &Order{
					ShopID: 456,
					Items: []OrderItem{
						{ProductID: shared.ID(1), Quantity: 2, Price: shared.Price(100)},
					},
					TotalPrice: shared.Price(200),
				}
			},
			setupMock: func(m *MockProductFinder) {
				m.On("FindProduct", shared.ID(1)).Return(&product.Product{
					ID:     shared.ID(1),
					ShopID: 456,
					Name:   "商品1",
					Stock:  10,
					Price:  shared.Price(100),
				}, nil)
			},
			wantErr: false,
			validate: func(t *testing.T, o *Order) {
				assert.Equal(t, "商品1", o.Items[0].ProductName)
				assert.Equal(t, shared.Price(200), o.TotalPrice)
			},
		},
		{
			name: "validation fails - product not found",
			order: func() *Order {
				return &Order{
					ShopID: 456,
					Items: []OrderItem{
						{ProductID: shared.ID(999), Quantity: 1},
					},
				}
			},
			setupMock: func(m *MockProductFinder) {
				m.On("FindProduct", shared.ID(999)).Return(nil, errors.New("not found"))
			},
			wantErr:     true,
			errContains: "商品不存在",
		},
		{
			name: "validation fails - insufficient stock",
			order: func() *Order {
				return &Order{
					ShopID: 456,
					Items: []OrderItem{
						{ProductID: shared.ID(1), Quantity: 10},
					},
				}
			},
			setupMock: func(m *MockProductFinder) {
				m.On("FindProduct", shared.ID(1)).Return(&product.Product{
					ID:     shared.ID(1),
					ShopID: 456,
					Stock:  5,
					Name:   "商品1",
				}, nil)
			},
			wantErr:     true,
			errContains: "库存不足",
		},
		{
			name: "validation fails - shop mismatch",
			order: func() *Order {
				return &Order{
					ShopID: 456,
					Items: []OrderItem{
						{ProductID: shared.ID(1), Quantity: 1},
					},
				}
			},
			setupMock: func(m *MockProductFinder) {
				m.On("FindProduct", shared.ID(1)).Return(&product.Product{
					ID:     shared.ID(1),
					ShopID: 789,
					Name:   "商品1",
					Stock:  10,
				}, nil)
			},
			wantErr:     true,
			errContains: "商品不属于该店铺",
		},
		{
			name: "calculate total fails - option not found",
			order: func() *Order {
				return &Order{
					ShopID: 456,
					Items: []OrderItem{
						{
							ProductID: shared.ID(1),
							Quantity:  2,
							Price:     shared.Price(100),
							Options: []OrderItemOption{
								{OptionID: shared.ID(999)},
							},
						},
					},
				}
			},
			setupMock: func(m *MockProductFinder) {
				m.On("FindProduct", shared.ID(1)).Return(&product.Product{
					ID:     shared.ID(1),
					ShopID: 456,
					Stock:  10,
					Name:   "商品1",
				}, nil)
				m.On("FindOption", shared.ID(999)).Return(nil, errors.New("option not found"))
			},
			wantErr:     true,
			errContains: "商品参数选项不存在",
		},
		{
			name: "valid order with multiple items and options",
			order: func() *Order {
				return &Order{
					ShopID: 456,
					Items: []OrderItem{
						{
							ProductID: shared.ID(1),
							Quantity:  2,
							Price:     shared.Price(100),
							Options: []OrderItemOption{
								{OptionID: shared.ID(10)},
							},
						},
						{
							ProductID: shared.ID(2),
							Quantity:  1,
							Price:     shared.Price(50),
							Options: []OrderItemOption{
								{OptionID: shared.ID(20)},
							},
						},
					},
				}
			},
			setupMock: func(m *MockProductFinder) {
				// First product and its option
				m.On("FindProduct", shared.ID(1)).Return(&product.Product{
					ID:     shared.ID(1),
					ShopID: 456,
					Stock:  10,
					Name:   "商品1",
					Price:  shared.Price(100),
				}, nil)
				m.On("FindOption", shared.ID(10)).Return(&product.ProductOption{
					ID:              shared.ID(10),
					CategoryID:      shared.ID(100),
					Name:            "大",
					PriceAdjustment: 10,
				}, nil)
				m.On("FindOptionCategory", shared.ID(100)).Return(&product.ProductOptionCategory{
					ID:   shared.ID(100),
					Name: "尺寸",
				}, nil)

				// Second product and its option
				m.On("FindProduct", shared.ID(2)).Return(&product.Product{
					ID:     shared.ID(2),
					ShopID: 456,
					Stock:  5,
					Name:   "商品2",
					Price:  shared.Price(50),
				}, nil)
				m.On("FindOption", shared.ID(20)).Return(&product.ProductOption{
					ID:              shared.ID(20),
					CategoryID:      shared.ID(200),
					Name:            "加冰",
					PriceAdjustment: 5,
				}, nil)
				m.On("FindOptionCategory", shared.ID(200)).Return(&product.ProductOptionCategory{
					ID:   shared.ID(200),
					Name: "温度",
				}, nil)
			},
			wantErr: false,
			validate: func(t *testing.T, o *Order) {
				// (100 + 10) * 2 + (50 + 5) * 1 = 220 + 55 = 275
				assert.Equal(t, shared.Price(275), o.TotalPrice)
				assert.Equal(t, "商品1", o.Items[0].ProductName)
				assert.Equal(t, "商品2", o.Items[1].ProductName)
				assert.Equal(t, "大", o.Items[0].Options[0].OptionName)
				assert.Equal(t, "加冰", o.Items[1].Options[0].OptionName)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockFinder := new(MockProductFinder)
			tt.setupMock(mockFinder)

			service := NewOrderDomainService()
			ord := tt.order()

			err := service.ValidateOrderCreation(ord, mockFinder)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
			} else {
				assert.NoError(t, err)
				if tt.validate != nil {
					tt.validate(t, ord)
				}
			}

			mockFinder.AssertExpectations(t)
		})
	}
}

func TestNewOrderDomainService(t *testing.T) {
	service := NewOrderDomainService()
	assert.NotNil(t, service)
}

func TestOrderDomainService_NilService(t *testing.T) {
	// Test that service methods handle nil receiver gracefully
	var service *OrderDomainService

	// This should panic or handle gracefully
	// In this case, we expect a panic when calling methods on nil service
	assert.Panics(t, func() {
		service.ValidateOrderCreation(nil, nil)
	})
}

func TestOrderDomainService_ValidationOrder(t *testing.T) {
	t.Run("validation order: items first, then calculate", func(t *testing.T) {
		mockFinder := new(MockProductFinder)

		// Setup calls in expected order
		mockFinder.On("FindProduct", shared.ID(1)).Return(&product.Product{
			ID:     shared.ID(1),
			ShopID: 456,
			Stock:  10,
			Name:   "商品1",
			Price:  shared.Price(100),
		}, nil)

		ord := &Order{
			ShopID: 456,
			Items: []OrderItem{
				{ProductID: shared.ID(1), Quantity: 2, Price: shared.Price(100)},
			},
		}

		service := NewOrderDomainService()
		err := service.ValidateOrderCreation(ord, mockFinder)

		assert.NoError(t, err)
		mockFinder.AssertExpectations(t)
	})
}
