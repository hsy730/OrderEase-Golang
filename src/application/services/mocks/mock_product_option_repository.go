package mocks

import (
	"orderease/domain/product"
	"orderease/domain/shared"

	"github.com/stretchr/testify/mock"
)

// MockProductOptionRepository is a mock implementation of product.ProductOptionRepository
type MockProductOptionRepository struct {
	mock.Mock
}

func (m *MockProductOptionRepository) Save(option *product.ProductOption) error {
	args := m.Called(option)
	return args.Error(0)
}

func (m *MockProductOptionRepository) FindByID(id shared.ID) (*product.ProductOption, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*product.ProductOption), args.Error(1)
}

func (m *MockProductOptionRepository) FindByCategoryID(categoryID shared.ID) ([]product.ProductOption, error) {
	args := m.Called(categoryID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]product.ProductOption), args.Error(1)
}

func (m *MockProductOptionRepository) DeleteByCategoryID(categoryID shared.ID) error {
	args := m.Called(categoryID)
	return args.Error(0)
}
