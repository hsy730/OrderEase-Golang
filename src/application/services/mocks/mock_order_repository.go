package mocks

import (
	"orderease/domain/order"
	"orderease/domain/shared"
	"time"

	"github.com/stretchr/testify/mock"
)

// MockOrderRepository is a mock implementation of order.OrderRepository
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

func (m *MockOrderRepository) FindByShopID(shopID uint64, page, pageSize int, search string, excludeOffline bool) ([]order.Order, int64, error) {
	args := m.Called(shopID, page, pageSize, search, excludeOffline)
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
