package services

import (
	"errors"
	"testing"
	"time"

	"orderease/application/dto"
	"orderease/domain/order"
	"orderease/domain/shared"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockOrderRepository is a mock for order.OrderRepository
type MockOrderRepository struct {
	mock.Mock
}

func (m *MockOrderRepository) Save(ord *order.Order) error {
	args := m.Called(ord)
	return args.Error(0)
}

func (m *MockOrderRepository) FindByID(id shared.ID) (*order.Order, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*order.Order), args.Error(1)
}

func (m *MockOrderRepository) FindByIDAndShopID(id shared.ID, shopID uint64) (*order.Order, error) {
	args := m.Called(id, shopID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*order.Order), args.Error(1)
}

func (m *MockOrderRepository) FindByShopID(shopID uint64, page, pageSize int) ([]order.Order, int64, error) {
	args := m.Called(shopID, page, pageSize)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]order.Order), args.Get(1).(int64), args.Error(2)
}

func (m *MockOrderRepository) FindByUserID(userID shared.ID, shopID uint64, page, pageSize int) ([]order.Order, int64, error) {
	args := m.Called(userID, shopID, page, pageSize)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]order.Order), args.Get(1).(int64), args.Error(2)
}

func (m *MockOrderRepository) FindUnfinishedByShopID(shopID uint64, flow order.OrderStatusFlow, page, pageSize int) ([]order.Order, int64, error) {
	args := m.Called(shopID, flow, page, pageSize)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]order.Order), args.Get(1).(int64), args.Error(2)
}

func (m *MockOrderRepository) Search(shopID uint64, userID string, statuses []order.OrderStatus, startTime, endTime time.Time, page, pageSize int) ([]order.Order, int64, error) {
	args := m.Called(shopID, userID, statuses, startTime, endTime, page, pageSize)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]order.Order), args.Get(1).(int64), args.Error(2)
}

func (m *MockOrderRepository) Delete(id shared.ID) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockOrderRepository) Update(ord *order.Order) error {
	args := m.Called(ord)
	return args.Error(0)
}

// MockOrderItemRepository is a mock for order.OrderItemRepository
type MockOrderItemRepository struct {
	mock.Mock
}

func (m *MockOrderItemRepository) Save(item *order.OrderItem) error {
	args := m.Called(item)
	return args.Error(0)
}

func (m *MockOrderItemRepository) FindByOrderID(orderID shared.ID) ([]order.OrderItem, error) {
	args := m.Called(orderID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]order.OrderItem), args.Error(1)
}

func (m *MockOrderItemRepository) DeleteByOrderID(orderID shared.ID) error {
	args := m.Called(orderID)
	return args.Error(0)
}

// MockOrderItemOptionRepository is a mock for order.OrderItemOptionRepository
type MockOrderItemOptionRepository struct {
	mock.Mock
}

func (m *MockOrderItemOptionRepository) Save(option *order.OrderItemOption) error {
	args := m.Called(option)
	return args.Error(0)
}

func (m *MockOrderItemOptionRepository) FindByOrderItemID(orderItemID shared.ID) ([]order.OrderItemOption, error) {
	args := m.Called(orderItemID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]order.OrderItemOption), args.Error(1)
}

func (m *MockOrderItemOptionRepository) DeleteByOrderItemID(orderItemID shared.ID) error {
	args := m.Called(orderItemID)
	return args.Error(0)
}

// MockOrderStatusLogRepository is a mock for order.OrderStatusLogRepository
type MockOrderStatusLogRepository struct {
	mock.Mock
}

func (m *MockOrderStatusLogRepository) Save(log *order.OrderStatusLog) error {
	args := m.Called(log)
	return args.Error(0)
}

func (m *MockOrderStatusLogRepository) FindByOrderID(orderID shared.ID) ([]order.OrderStatusLog, error) {
	args := m.Called(orderID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]order.OrderStatusLog), args.Error(1)
}

func (m *MockOrderStatusLogRepository) DeleteByOrderID(orderID shared.ID) error {
	args := m.Called(orderID)
	return args.Error(0)
}

func TestNewOrderService(t *testing.T) {
	mockProductRepo := new(MockProductRepository)
	mockProductOptionRepo := new(MockProductOptionRepository)
	mockProductOptionCategoryRepo := new(MockProductOptionCategoryRepository)
	mockOrderRepo := new(MockOrderRepository)
	mockOrderItemRepo := new(MockOrderItemRepository)
	mockOrderItemOptionRepo := new(MockOrderItemOptionRepository)
	mockOrderStatusLogRepo := new(MockOrderStatusLogRepository)

	// Use nil for DB since we're testing methods that don't require DB operations
	service := NewOrderService(
		nil,
		mockProductRepo,
		mockProductOptionRepo,
		mockProductOptionCategoryRepo,
		mockOrderRepo,
		mockOrderItemRepo,
		mockOrderItemOptionRepo,
		mockOrderStatusLogRepo,
	)

	assert.NotNil(t, service)
}

func TestOrderService_buildOrderItems(t *testing.T) {
	mockProductRepo := new(MockProductRepository)
	mockProductOptionRepo := new(MockProductOptionRepository)
	mockProductOptionCategoryRepo := new(MockProductOptionCategoryRepository)
	mockOrderRepo := new(MockOrderRepository)
	mockOrderItemRepo := new(MockOrderItemRepository)
	mockOrderItemOptionRepo := new(MockOrderItemOptionRepository)
	mockOrderStatusLogRepo := new(MockOrderStatusLogRepository)

	service := NewOrderService(
		nil,
		mockProductRepo,
		mockProductOptionRepo,
		mockProductOptionCategoryRepo,
		mockOrderRepo,
		mockOrderItemRepo,
		mockOrderItemOptionRepo,
		mockOrderStatusLogRepo,
	)

	tests := []struct {
		name     string
		reqItems []dto.CreateOrderItemRequest
		wantLen  int
		validate func(*testing.T, []order.OrderItem)
	}{
		{
			name: "single item without options",
			reqItems: []dto.CreateOrderItemRequest{
				{
					ProductID: shared.ID(1),
					Quantity:  2,
					Price:     100,
					Options:   []dto.CreateOrderItemOption{},
				},
			},
			wantLen: 1,
			validate: func(t *testing.T, items []order.OrderItem) {
				assert.Equal(t, shared.ID(1), items[0].ProductID)
				assert.Equal(t, 2, items[0].Quantity)
				assert.Equal(t, shared.Price(100), items[0].Price)
				assert.Empty(t, items[0].Options)
			},
		},
		{
			name: "item with options",
			reqItems: []dto.CreateOrderItemRequest{
				{
					ProductID: shared.ID(2),
					Quantity:  1,
					Price:     50,
					Options: []dto.CreateOrderItemOption{
						{CategoryID: shared.ID(10), OptionID: shared.ID(20)},
						{CategoryID: shared.ID(11), OptionID: shared.ID(21)},
					},
				},
			},
			wantLen: 1,
			validate: func(t *testing.T, items []order.OrderItem) {
				assert.Equal(t, shared.ID(2), items[0].ProductID)
				assert.Len(t, items[0].Options, 2)
				assert.Equal(t, shared.ID(10), items[0].Options[0].CategoryID)
				assert.Equal(t, shared.ID(20), items[0].Options[0].OptionID)
			},
		},
		{
			name: "multiple items",
			reqItems: []dto.CreateOrderItemRequest{
				{ProductID: shared.ID(1), Quantity: 2, Price: 100, Options: []dto.CreateOrderItemOption{}},
				{ProductID: shared.ID(2), Quantity: 1, Price: 50, Options: []dto.CreateOrderItemOption{}},
			},
			wantLen: 2,
			validate: func(t *testing.T, items []order.OrderItem) {
				assert.Equal(t, shared.ID(1), items[0].ProductID)
				assert.Equal(t, shared.ID(2), items[1].ProductID)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := service.buildOrderItems(tt.reqItems)
			assert.Len(t, got, tt.wantLen)
			if tt.validate != nil {
				tt.validate(t, got)
			}
		})
	}
}

func TestOrderService_GetOrder(t *testing.T) {
	mockProductRepo := new(MockProductRepository)
	mockProductOptionRepo := new(MockProductOptionRepository)
	mockProductOptionCategoryRepo := new(MockProductOptionCategoryRepository)
	mockOrderRepo := new(MockOrderRepository)
	mockOrderItemRepo := new(MockOrderItemRepository)
	mockOrderItemOptionRepo := new(MockOrderItemOptionRepository)
	mockOrderStatusLogRepo := new(MockOrderStatusLogRepository)

	service := NewOrderService(
		nil,
		mockProductRepo,
		mockProductOptionRepo,
		mockProductOptionCategoryRepo,
		mockOrderRepo,
		mockOrderItemRepo,
		mockOrderItemOptionRepo,
		mockOrderStatusLogRepo,
	)

	tests := []struct {
		name        string
		orderID     shared.ID
		shopID      uint64
		setupMock   func()
		wantErr     bool
		errContains string
		validate    func(*testing.T, *dto.OrderDetailResponse)
	}{
		{
			name:    "get existing order",
			orderID: shared.ID(123),
			shopID:  456,
			setupMock: func() {
				ord := &order.Order{
					ID:         shared.ID(123),
					UserID:     shared.ID(789),
					ShopID:     456,
					TotalPrice: shared.Price(200),
					Status:     order.OrderStatusPending,
					Remark:     "测试订单",
					Items: []order.OrderItem{
						{
							ID:          shared.ID(1),
							ProductID:   shared.ID(10),
							Quantity:    2,
							Price:       shared.Price(100),
							TotalPrice:  shared.Price(200),
							ProductName: "商品1",
							Options: []order.OrderItemOption{
								{
									ID:           shared.ID(100),
									CategoryID:   shared.ID(10),
									OptionID:     shared.ID(20),
									OptionName:   "大",
									CategoryName: "尺寸",
								},
							},
						},
					},
				}
				mockOrderRepo.On("FindByIDAndShopID", shared.ID(123), uint64(456)).Return(ord, nil)
			},
			wantErr: false,
			validate: func(t *testing.T, resp *dto.OrderDetailResponse) {
				assert.Equal(t, shared.ID(123), resp.ID)
				assert.Equal(t, shared.ID(789), resp.UserID)
				assert.Equal(t, uint64(456), resp.ShopID)
				assert.Equal(t, shared.Price(200), resp.TotalPrice)
				assert.Equal(t, order.OrderStatusPending, resp.Status)
				assert.Len(t, resp.Items, 1)
				assert.Equal(t, "商品1", resp.Items[0].ProductName)
				assert.Len(t, resp.Items[0].Options, 1)
				assert.Equal(t, "大", resp.Items[0].Options[0].OptionName)
			},
		},
		{
			name:    "order not found",
			orderID: shared.ID(999),
			shopID:  456,
			setupMock: func() {
				mockOrderRepo.On("FindByIDAndShopID", shared.ID(999), uint64(456)).Return(nil, errors.New("not found"))
			},
			wantErr:     true,
			errContains: "not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			got, err := service.GetOrder(tt.orderID, tt.shopID)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, got)
				if tt.validate != nil {
					tt.validate(t, got)
				}
			}

			mockOrderRepo.AssertExpectations(t)
		})
	}
}

func TestOrderService_GetOrders(t *testing.T) {
	mockProductRepo := new(MockProductRepository)
	mockProductOptionRepo := new(MockProductOptionRepository)
	mockProductOptionCategoryRepo := new(MockProductOptionCategoryRepository)
	mockOrderRepo := new(MockOrderRepository)
	mockOrderItemRepo := new(MockOrderItemRepository)
	mockOrderItemOptionRepo := new(MockOrderItemOptionRepository)
	mockOrderStatusLogRepo := new(MockOrderStatusLogRepository)

	service := NewOrderService(
		nil,
		mockProductRepo,
		mockProductOptionRepo,
		mockProductOptionCategoryRepo,
		mockOrderRepo,
		mockOrderItemRepo,
		mockOrderItemOptionRepo,
		mockOrderStatusLogRepo,
	)

	tests := []struct {
		name        string
		shopID      uint64
		page        int
		pageSize    int
		setupMock   func()
		wantErr     bool
		validate    func(*testing.T, *dto.OrderListResponse)
	}{
		{
			name:     "get orders successfully",
			shopID:   456,
			page:     1,
			pageSize: 10,
			setupMock: func() {
				orders := []order.Order{
					{ID: shared.ID(1), UserID: shared.ID(10), ShopID: 456, TotalPrice: shared.Price(100), Status: order.OrderStatusPending},
					{ID: shared.ID(2), UserID: shared.ID(11), ShopID: 456, TotalPrice: shared.Price(200), Status: order.OrderStatusAccepted},
				}
				mockOrderRepo.On("FindByShopID", uint64(456), 1, 10).Return(orders, int64(2), nil)
			},
			wantErr: false,
			validate: func(t *testing.T, resp *dto.OrderListResponse) {
				assert.Equal(t, int64(2), resp.Total)
				assert.Equal(t, 1, resp.Page)
				assert.Equal(t, 10, resp.PageSize)
				assert.Len(t, resp.Data, 2)
				assert.Equal(t, shared.ID(1), resp.Data[0].ID)
				assert.Equal(t, shared.ID(2), resp.Data[1].ID)
			},
		},
		{
			name:     "empty orders",
			shopID:   456,
			page:     2,
			pageSize: 10,
			setupMock: func() {
				mockOrderRepo.On("FindByShopID", uint64(456), 2, 10).Return([]order.Order{}, int64(0), nil)
			},
			wantErr: false,
			validate: func(t *testing.T, resp *dto.OrderListResponse) {
				assert.Equal(t, int64(0), resp.Total)
				assert.Len(t, resp.Data, 0)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			got, err := service.GetOrders(tt.shopID, tt.page, tt.pageSize)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, got)
				if tt.validate != nil {
					tt.validate(t, got)
				}
			}

			mockOrderRepo.AssertExpectations(t)
		})
	}
}

func TestOrderService_GetOrdersByUser(t *testing.T) {
	mockProductRepo := new(MockProductRepository)
	mockProductOptionRepo := new(MockProductOptionRepository)
	mockProductOptionCategoryRepo := new(MockProductOptionCategoryRepository)
	mockOrderRepo := new(MockOrderRepository)
	mockOrderItemRepo := new(MockOrderItemRepository)
	mockOrderItemOptionRepo := new(MockOrderItemOptionRepository)
	mockOrderStatusLogRepo := new(MockOrderStatusLogRepository)

	service := NewOrderService(
		nil,
		mockProductRepo,
		mockProductOptionRepo,
		mockProductOptionCategoryRepo,
		mockOrderRepo,
		mockOrderItemRepo,
		mockOrderItemOptionRepo,
		mockOrderStatusLogRepo,
	)

	tests := []struct {
		name        string
		userID      shared.ID
		shopID      uint64
		page        int
		pageSize    int
		setupMock   func()
		wantErr     bool
		validate    func(*testing.T, *dto.OrderListResponse)
	}{
		{
			name:     "get user orders successfully",
			userID:   shared.ID(100),
			shopID:   456,
			page:     1,
			pageSize: 10,
			setupMock: func() {
				orders := []order.Order{
					{ID: shared.ID(1), UserID: shared.ID(100), ShopID: 456, TotalPrice: shared.Price(100), Status: order.OrderStatusPending},
				}
				mockOrderRepo.On("FindByUserID", shared.ID(100), uint64(456), 1, 10).Return(orders, int64(1), nil)
			},
			wantErr: false,
			validate: func(t *testing.T, resp *dto.OrderListResponse) {
				assert.Equal(t, int64(1), resp.Total)
				assert.Len(t, resp.Data, 1)
				assert.Equal(t, shared.ID(100), resp.Data[0].UserID)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			got, err := service.GetOrdersByUser(tt.userID, tt.shopID, tt.page, tt.pageSize)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, got)
				if tt.validate != nil {
					tt.validate(t, got)
				}
			}

			mockOrderRepo.AssertExpectations(t)
		})
	}
}

func TestOrderService_GetUnfinishedOrders(t *testing.T) {
	mockProductRepo := new(MockProductRepository)
	mockProductOptionRepo := new(MockProductOptionRepository)
	mockProductOptionCategoryRepo := new(MockProductOptionCategoryRepository)
	mockOrderRepo := new(MockOrderRepository)
	mockOrderItemRepo := new(MockOrderItemRepository)
	mockOrderItemOptionRepo := new(MockOrderItemOptionRepository)
	mockOrderStatusLogRepo := new(MockOrderStatusLogRepository)

	service := NewOrderService(
		nil,
		mockProductRepo,
		mockProductOptionRepo,
		mockProductOptionCategoryRepo,
		mockOrderRepo,
		mockOrderItemRepo,
		mockOrderItemOptionRepo,
		mockOrderStatusLogRepo,
	)

	flow := order.OrderStatusFlow{
		Statuses: []order.OrderStatusConfig{
			{Value: order.OrderStatusPending, IsFinal: false},
			{Value: order.OrderStatusAccepted, IsFinal: false},
		},
	}

	tests := []struct {
		name        string
		shopID      uint64
		page        int
		pageSize    int
		setupMock   func()
		wantErr     bool
		validate    func(*testing.T, *dto.OrderListResponse)
	}{
		{
			name:     "get unfinished orders",
			shopID:   456,
			page:     1,
			pageSize: 10,
			setupMock: func() {
				orders := []order.Order{
					{ID: shared.ID(1), ShopID: 456, Status: order.OrderStatusPending},
					{ID: shared.ID(2), ShopID: 456, Status: order.OrderStatusAccepted},
				}
				mockOrderRepo.On("FindUnfinishedByShopID", uint64(456), flow, 1, 10).Return(orders, int64(2), nil)
			},
			wantErr: false,
			validate: func(t *testing.T, resp *dto.OrderListResponse) {
				assert.Equal(t, int64(2), resp.Total)
				assert.Len(t, resp.Data, 2)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			got, err := service.GetUnfinishedOrders(tt.shopID, flow, tt.page, tt.pageSize)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, got)
				if tt.validate != nil {
					tt.validate(t, got)
				}
			}

			mockOrderRepo.AssertExpectations(t)
		})
	}
}
