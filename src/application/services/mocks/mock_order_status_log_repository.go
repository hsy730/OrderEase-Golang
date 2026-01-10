package mocks

import (
	"orderease/domain/order"
	"orderease/domain/shared"

	"github.com/stretchr/testify/mock"
)

// MockOrderStatusLogRepository is a mock implementation of order.OrderStatusLogRepository
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
