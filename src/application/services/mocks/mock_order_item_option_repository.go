package mocks

import (
	"orderease/domain/order"
	"orderease/domain/shared"

	"github.com/stretchr/testify/mock"
)

// MockOrderItemOptionRepository is a mock implementation of order.OrderItemOptionRepository
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
