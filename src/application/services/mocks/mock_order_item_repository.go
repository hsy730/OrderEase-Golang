package mocks

import (
	"orderease/domain/order"
	"orderease/domain/shared"

	"github.com/stretchr/testify/mock"
)

// MockOrderItemRepository is a mock implementation of order.OrderItemRepository
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
