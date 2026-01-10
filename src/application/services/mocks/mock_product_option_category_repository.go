package mocks

import (
	"orderease/domain/product"
	"orderease/domain/shared"

	"github.com/stretchr/testify/mock"
)

// MockProductOptionCategoryRepository is a mock implementation of product.ProductOptionCategoryRepository
type MockProductOptionCategoryRepository struct {
	mock.Mock
}

func (m *MockProductOptionCategoryRepository) Save(category *product.ProductOptionCategory) error {
	args := m.Called(category)
	return args.Error(0)
}

func (m *MockProductOptionCategoryRepository) FindByID(id shared.ID) (*product.ProductOptionCategory, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*product.ProductOptionCategory), args.Error(1)
}

func (m *MockProductOptionCategoryRepository) FindByProductID(productID shared.ID) ([]product.ProductOptionCategory, error) {
	args := m.Called(productID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]product.ProductOptionCategory), args.Error(1)
}

func (m *MockProductOptionCategoryRepository) DeleteByProductID(productID shared.ID) error {
	args := m.Called(productID)
	return args.Error(0)
}
