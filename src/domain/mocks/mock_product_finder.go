package mocks

import (
	"orderease/domain/order"
	"orderease/domain/product"
	"orderease/domain/shared"

	"github.com/stretchr/testify/mock"
)

// MockProductFinder is a mock implementation of order.ProductFinder
type MockProductFinder struct {
	mock.Mock
}

// FindProduct provides a mock function with given fields: id
func (m *MockProductFinder) FindProduct(id shared.ID) (*product.Product, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*product.Product), args.Error(1)
}

// FindOption provides a mock function with given fields: id
func (m *MockProductFinder) FindOption(id shared.ID) (*product.ProductOption, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*product.ProductOption), args.Error(1)
}

// FindOptionCategory provides a mock function with given fields: id
func (m *MockProductFinder) FindOptionCategory(id shared.ID) (*product.ProductOptionCategory, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*product.ProductOptionCategory), args.Error(1)
}

// Helper function to create a mock product finder with predefined products
func NewMockProductFinderWithProducts(products map[shared.ID]*product.Product) *MockProductFinder {
	mockFinder := &MockProductFinder{}
	for id, prod := range products {
		mockFinder.On("FindProduct", id).Return(prod, nil)
	}
	return mockFinder
}

// Helper function to create a mock product finder with predefined options
func NewMockProductFinderWithOptions(options map[shared.ID]*product.ProductOption) *MockProductFinder {
	mockFinder := &MockProductFinder{}
	for id, opt := range options {
		mockFinder.On("FindOption", id).Return(opt, nil)
	}
	return mockFinder
}

// Helper function to create a mock product finder with predefined categories
func NewMockProductFinderWithCategories(categories map[shared.ID]*product.ProductOptionCategory) *MockProductFinder {
	mockFinder := &MockProductFinder{}
	for id, cat := range categories {
		mockFinder.On("FindOptionCategory", id).Return(cat, nil)
	}
	return mockFinder
}

// Ensure MockProductFinder implements ProductFinder interface
var _ order.ProductFinder = (*MockProductFinder)(nil)
