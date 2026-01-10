package order

import (
	"errors"
	"testing"
	"time"

	"orderease/domain/product"
	"orderease/domain/shared"

	"github.com/stretchr/testify/assert"
)

func TestNewOrder(t *testing.T) {
	tests := []struct {
		name    string
		userID  shared.ID
		shopID  uint64
		items   []OrderItem
		remark  string
		wantErr bool
		errMsg  string
		validate func(*testing.T, *Order)
	}{
		{
			name:   "valid order",
			userID: shared.ID(123),
			shopID: 456,
			items: []OrderItem{
				{
					ProductID:  shared.ID(789),
					Quantity:   2,
					Price:      shared.Price(100),
					TotalPrice: shared.Price(200),
				},
			},
			remark:  "测试订单",
			wantErr: false,
			validate: func(t *testing.T, o *Order) {
				assert.Equal(t, shared.ID(0), o.ID)
				assert.Equal(t, shared.ID(123), o.UserID)
				assert.Equal(t, uint64(456), o.ShopID)
				assert.Equal(t, OrderStatusPending, o.Status)
				assert.Equal(t, "测试订单", o.Remark)
				assert.Equal(t, shared.Price(200), o.TotalPrice)
				assert.Len(t, o.Items, 1)
				assert.False(t, o.CreatedAt.IsZero())
				assert.False(t, o.UpdatedAt.IsZero())
			},
		},
		{
			name:    "empty userID",
			userID:  shared.ID(0),
			shopID:  456,
			items:   validOrderItems(),
			remark:  "",
			wantErr: true,
			errMsg:  "用户ID不能为空",
		},
		{
			name:    "empty shopID",
			userID:  shared.ID(123),
			shopID:  0,
			items:   validOrderItems(),
			remark:  "",
			wantErr: true,
			errMsg:  "店铺ID不能为空",
		},
		{
			name:    "empty items",
			userID:  shared.ID(123),
			shopID:  456,
			items:   []OrderItem{},
			remark:  "",
			wantErr: true,
			errMsg:  "订单项不能为空",
		},
		{
			name:    "item with zero productID",
			userID:  shared.ID(123),
			shopID:  456,
			items: []OrderItem{
				{ProductID: shared.ID(0), Quantity: 1, Price: shared.Price(100)},
			},
			remark:  "",
			wantErr: true,
			errMsg:  "商品ID不能为空",
		},
		{
			name:    "item with zero quantity",
			userID:  shared.ID(123),
			shopID:  456,
			items: []OrderItem{
				{ProductID: shared.ID(789), Quantity: 0, Price: shared.Price(100)},
			},
			remark:  "",
			wantErr: true,
			errMsg:  "商品数量必须大于0",
		},
		{
			name:    "item with negative quantity",
			userID:  shared.ID(123),
			shopID:  456,
			items: []OrderItem{
				{ProductID: shared.ID(789), Quantity: -1, Price: shared.Price(100)},
			},
			remark:  "",
			wantErr: true,
			errMsg:  "商品数量必须大于0",
		},
		{
			name:   "multiple items",
			userID: shared.ID(123),
			shopID: 456,
			items: []OrderItem{
				{ProductID: shared.ID(1), Quantity: 2, Price: shared.Price(100), TotalPrice: shared.Price(200)},
				{ProductID: shared.ID(2), Quantity: 1, Price: shared.Price(50), TotalPrice: shared.Price(50)},
			},
			remark:  "",
			wantErr: false,
			validate: func(t *testing.T, o *Order) {
				assert.Equal(t, shared.Price(250), o.TotalPrice)
				assert.Len(t, o.Items, 2)
			},
		},
		{
			name:   "valid with empty remark",
			userID: shared.ID(123),
			shopID: 456,
			items:  validOrderItems(),
			remark: "",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewOrder(tt.userID, tt.shopID, tt.items, tt.remark)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, got)
				if tt.validate != nil {
					tt.validate(t, got)
				}
			}
		})
	}
}

func TestOrder_TransitionTo(t *testing.T) {
	tests := []struct {
		name        string
		initialStatus OrderStatus
		newStatus   OrderStatus
		flow        OrderStatusFlow
		wantErr     bool
		errMsg      string
		validate    func(*testing.T, *Order)
	}{
		{
			name:         "valid transition pending to accepted",
			initialStatus: OrderStatusPending,
			newStatus:    OrderStatusAccepted,
			flow:         createTestFlow(),
			wantErr:      false,
			validate: func(t *testing.T, o *Order) {
				assert.Equal(t, OrderStatusAccepted, o.Status)
				assert.True(t, o.UpdatedAt.After(time.Time{}) || o.UpdatedAt.Equal(time.Time{}))
			},
		},
		{
			name:         "valid transition accepted to shipped",
			initialStatus: OrderStatusAccepted,
			newStatus:    OrderStatusShipped,
			flow:         createTestFlow(),
			wantErr:      false,
		},
		{
			name:         "valid transition shipped to complete",
			initialStatus: OrderStatusShipped,
			newStatus:    OrderStatusComplete,
			flow:         createTestFlow(),
			wantErr:      false,
		},
		{
			name:         "invalid transition pending to complete",
			initialStatus: OrderStatusPending,
			newStatus:    OrderStatusComplete,
			flow:         createTestFlow(),
			wantErr:      true,
			errMsg:       "不允许转换到状态",
		},
		{
			name:         "transition from final status",
			initialStatus: OrderStatusComplete,
			newStatus:    OrderStatusPending,
			flow:         createTestFlow(),
			wantErr:      true,
			errMsg:       "不允许转换到状态",
		},
		{
			name:         "transition to same status",
			initialStatus: OrderStatusPending,
			newStatus:    OrderStatusPending,
			flow:         createTestFlow(),
			wantErr:      true,
			errMsg:       "不允许转换到状态",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			before := time.Now()
			o := &Order{
				Status:    tt.initialStatus,
				UpdatedAt: before,
			}

			err := o.TransitionTo(tt.newStatus, tt.flow)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				assert.Equal(t, tt.initialStatus, o.Status, "status should not change on error")
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.newStatus, o.Status)
				assert.True(t, o.UpdatedAt.After(before) || o.UpdatedAt.Equal(before))
				if tt.validate != nil {
					tt.validate(t, o)
				}
			}
		})
	}
}

func TestOrder_CanTransitionTo(t *testing.T) {
	tests := []struct {
		name         string
		currentStatus OrderStatus
		newStatus    OrderStatus
		flow         OrderStatusFlow
		wantErr      bool
	}{
		{
			name:         "allowed transition",
			currentStatus: OrderStatusPending,
			newStatus:    OrderStatusAccepted,
			flow:         createTestFlow(),
			wantErr:      false,
		},
		{
			name:         "not allowed transition",
			currentStatus: OrderStatusPending,
			newStatus:    OrderStatusComplete,
			flow:         createTestFlow(),
			wantErr:      true,
		},
		{
			name:         "from final status",
			currentStatus: OrderStatusComplete,
			newStatus:    OrderStatusPending,
			flow:         createTestFlow(),
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := &Order{Status: tt.currentStatus}
			err := o.CanTransitionTo(tt.newStatus, tt.flow)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestOrder_IsFinal(t *testing.T) {
	tests := []struct {
		name     string
		status   OrderStatus
		expected bool
	}{
		{"pending is not final", OrderStatusPending, false},
		{"accepted is not final", OrderStatusAccepted, false},
		{"shipped is not final", OrderStatusShipped, false},
		{"complete is final", OrderStatusComplete, true},
		{"rejected is final", OrderStatusRejected, true},
		{"canceled is final", OrderStatusCanceled, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := &Order{Status: tt.status}
			got := o.IsFinal()
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestOrder_IsUnfinished(t *testing.T) {
	flow := createTestFlow()

	tests := []struct {
		name     string
		status   OrderStatus
		expected bool
	}{
		{"pending is unfinished", OrderStatusPending, true},
		{"accepted is unfinished", OrderStatusAccepted, true},
		{"shipped is unfinished", OrderStatusShipped, true},
		{"complete is finished", OrderStatusComplete, false},
		{"rejected is finished", OrderStatusRejected, false},
		{"canceled is finished", OrderStatusCanceled, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := &Order{Status: tt.status}
			got := o.IsUnfinished(flow)
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestOrder_ValidateItems(t *testing.T) {
	tests := []struct {
		name        string
		order       func() *Order
		setupMock   func(*MockProductFinder)
		wantErr     bool
		errContains string
		validate    func(*testing.T, *Order)
	}{
		{
			name: "all items valid",
			order: func() *Order {
				return &Order{
					ShopID: 456,
					Items: []OrderItem{
						{ProductID: shared.ID(1), Quantity: 2},
						{ProductID: shared.ID(2), Quantity: 1},
					},
				}
			},
			setupMock: func(m *MockProductFinder) {
				m.On("FindProduct", shared.ID(1)).Return(&product.Product{
					ID:     shared.ID(1),
					ShopID: 456,
					Name:   "商品1",
					Stock:  10,
				}, nil)
				m.On("FindProduct", shared.ID(2)).Return(&product.Product{
					ID:     shared.ID(2),
					ShopID: 456,
					Name:   "商品2",
					Stock:  5,
				}, nil)
			},
			wantErr: false,
			validate: func(t *testing.T, o *Order) {
				assert.Equal(t, "商品1", o.Items[0].ProductName)
				assert.Equal(t, "商品2", o.Items[1].ProductName)
			},
		},
		{
			name: "product not found",
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
			name: "shop mismatch",
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
			name: "insufficient stock",
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
					Name:   "商品1",
					Stock:  5,
				}, nil)
			},
			wantErr:     true,
			errContains: "库存不足",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockFinder := new(MockProductFinder)
			tt.setupMock(mockFinder)

			ord := tt.order()
			err := ord.ValidateItems(mockFinder)

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

func TestOrder_CalculateTotal(t *testing.T) {
	tests := []struct {
		name        string
		order       func() *Order
		setupMock   func(*MockProductFinder)
		wantTotal   shared.Price
		wantErr     bool
		errContains string
	}{
		{
			name: "calculate total without options",
			order: func() *Order {
				return &Order{
					Items: []OrderItem{
						{Price: shared.Price(100), Quantity: 2, Options: []OrderItemOption{}},
						{Price: shared.Price(50), Quantity: 1, Options: []OrderItemOption{}},
					},
				}
			},
			setupMock: func(m *MockProductFinder) {},
			wantTotal: shared.Price(250),
			wantErr:   false,
		},
		{
			name: "calculate with item options",
			order: func() *Order {
				return &Order{
					Items: []OrderItem{
						{
							Price:    shared.Price(100),
							Quantity: 2,
							Options: []OrderItemOption{
								{OptionID: shared.ID(1)},
							},
						},
					},
				}
			},
			setupMock: func(m *MockProductFinder) {
				m.On("FindOption", shared.ID(1)).Return(&product.ProductOption{
					ID:              shared.ID(1),
					CategoryID:      shared.ID(10),
					Name:            "大",
					PriceAdjustment: 10,
				}, nil)
				m.On("FindOptionCategory", shared.ID(10)).Return(&product.ProductOptionCategory{
					ID:   shared.ID(10),
					Name: "尺寸",
				}, nil)
			},
			wantTotal: shared.Price(220), // (100 + 10) * 2
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockFinder := new(MockProductFinder)
			tt.setupMock(mockFinder)

			ord := tt.order()
			err := ord.CalculateTotal(mockFinder)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantTotal, ord.TotalPrice)
			}

			mockFinder.AssertExpectations(t)
		})
	}
}

func TestOrder_Timestamps(t *testing.T) {
	before := time.Now()
	ord, err := NewOrder(shared.ID(123), 456, validOrderItems(), "测试")
	after := time.Now()

	assert.NoError(t, err)
	assert.True(t, ord.CreatedAt.After(before) || ord.CreatedAt.Equal(before))
	assert.True(t, ord.CreatedAt.Before(after) || ord.CreatedAt.Equal(after))
	assert.True(t, ord.UpdatedAt.After(before) || ord.UpdatedAt.Equal(before))
	assert.True(t, ord.UpdatedAt.Before(after) || ord.UpdatedAt.Equal(after))
}

// Helper functions

func validOrderItems() []OrderItem {
	return []OrderItem{
		{
			ProductID:  shared.ID(789),
			Quantity:   2,
			Price:      shared.Price(100),
			TotalPrice: shared.Price(200),
		},
	}
}

func createTestFlow() OrderStatusFlow {
	return OrderStatusFlow{
		Statuses: []OrderStatusConfig{
			{
				Value:   OrderStatusPending,
				IsFinal: false,
				Actions: []OrderStatusTransition{
					{NextStatus: OrderStatusAccepted},
					{NextStatus: OrderStatusRejected},
					{NextStatus: OrderStatusCanceled},
				},
			},
			{
				Value:   OrderStatusAccepted,
				IsFinal: false,
				Actions: []OrderStatusTransition{
					{NextStatus: OrderStatusShipped},
					{NextStatus: OrderStatusCanceled},
				},
			},
			{
				Value:   OrderStatusShipped,
				IsFinal: false,
				Actions: []OrderStatusTransition{
					{NextStatus: OrderStatusComplete},
				},
			},
			{
				Value:   OrderStatusComplete,
				IsFinal: true,
				Actions: []OrderStatusTransition{},
			},
			{
				Value:   OrderStatusRejected,
				IsFinal: true,
				Actions: []OrderStatusTransition{},
			},
			{
				Value:   OrderStatusCanceled,
				IsFinal: true,
				Actions: []OrderStatusTransition{},
			},
		},
	}
}
