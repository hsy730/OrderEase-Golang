package mocks

import (
	"orderease/domain/product"
	"orderease/domain/shared"

	"github.com/stretchr/testify/mock"
)

// MockProductRepository is a mock implementation of product.ProductRepository
type MockProductRepository struct {
	mock.Mock
}

func (m *MockProductRepository) Save(prod *product.Product) error {
	args := m.Called(prod)
	return args.Error(0)
}

func (m *MockProductRepository) FindByID(id shared.ID) (*product.Product, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*product.Product), args.Error(1)
}

func (m *MockProductRepository) FindByIDAndShopID(id shared.ID, shopID uint64) (*product.Product, error) {
	args := m.Called(id, shopID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*product.Product), args.Error(1)
}

func (m *MockProductRepository) FindByShopID(shopID uint64, page, pageSize int, search string, excludeOffline bool) ([]product.Product, int64, error) {
	args := m.Called(shopID, page, pageSize, search, excludeOffline)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]product.Product), args.Get(1).(int64), args.Error(2)
}

func (m *MockProductRepository) FindByIDs(ids []shared.ID) ([]product.Product, error) {
	args := m.Called(ids)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]product.Product), args.Error(1)
}

func (m *MockProductRepository) Delete(id shared.ID) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockProductRepository) Update(prod *product.Product) error {
	args := m.Called(prod)
	return args.Error(0)
}

func (m *MockProductRepository) CountByProductID(productID shared.ID) (int64, error) {
	args := m.Called(productID)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockProductRepository) FindOptionByID(id shared.ID) (*product.ProductOption, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*product.ProductOption), args.Error(1)
}

func (m *MockProductRepository) FindOptionCategoryByID(id shared.ID) (*product.ProductOptionCategory, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*product.ProductOptionCategory), args.Error(1)
}
