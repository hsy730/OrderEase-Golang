package mocks

import (
	"orderease/models"

	"github.com/bwmarrin/snowflake"
	"github.com/stretchr/testify/mock"
)

// MockProductRepository Mock 商品仓储
type MockProductRepository struct {
	mock.Mock
}

func (m *MockProductRepository) GetByID(id snowflake.ID) (*models.Product, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Product), args.Error(1)
}

func (m *MockProductRepository) GetProductsByShop(shopID snowflake.ID, onlyOnline bool) ([]*models.Product, error) {
	args := m.Called(shopID, onlyOnline)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Product), args.Error(1)
}

func (m *MockProductRepository) Create(product *models.Product) error {
	args := m.Called(product)
	return args.Error(0)
}

func (m *MockProductRepository) Update(product *models.Product) error {
	args := m.Called(product)
	return args.Error(0)
}

func (m *MockProductRepository) Delete(product *models.Product) error {
	args := m.Called(product)
	return args.Error(0)
}

// MockOrderRepository Mock 订单仓储
type MockOrderRepository struct {
	mock.Mock
}

func (m *MockOrderRepository) GetByID(id snowflake.ID) (*models.Order, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Order), args.Error(1)
}

func (m *MockOrderRepository) Create(order *models.Order) error {
	args := m.Called(order)
	return args.Error(0)
}

func (m *MockOrderRepository) Update(order *models.Order) error {
	args := m.Called(order)
	return args.Error(0)
}

// MockShopRepository Mock 店铺仓储
type MockShopRepository struct {
	mock.Mock
}

func (m *MockShopRepository) GetByID(id snowflake.ID) (*models.Shop, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Shop), args.Error(1)
}

func (m *MockShopRepository) GetByName(name string) (*models.Shop, error) {
	args := m.Called(name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Shop), args.Error(1)
}

func (m *MockShopRepository) Create(shop *models.Shop) error {
	args := m.Called(shop)
	return args.Error(0)
}

func (m *MockShopRepository) Update(shop *models.Shop) error {
	args := m.Called(shop)
	return args.Error(0)
}

// MockUserRepository Mock 用户仓储
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) GetUserByID(id string) (*models.User, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) GetByUsername(username string) (*models.User, error) {
	args := m.Called(username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) Create(user *models.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockUserRepository) Update(user *models.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockUserRepository) CheckPhoneExists(phone string) (bool, error) {
	args := m.Called(phone)
	return args.Bool(0), args.Error(1)
}

func (m *MockUserRepository) CheckUsernameExists(username string) (bool, error) {
	args := m.Called(username)
	return args.Bool(0), args.Error(1)
}
