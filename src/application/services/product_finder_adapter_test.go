package services

import (
	"errors"
	"testing"

	"orderease/domain/product"
	"orderease/domain/shared"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockProductRepository is a mock for product.ProductRepository
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

// MockProductOptionRepository is a mock for product.ProductOptionRepository
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

// MockProductOptionCategoryRepository is a mock for product.ProductOptionCategoryRepository
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

func TestNewProductFinderAdapter(t *testing.T) {
	mockProductRepo := new(MockProductRepository)
	mockOptionRepo := new(MockProductOptionRepository)
	mockCategoryRepo := new(MockProductOptionCategoryRepository)

	adapter := NewProductFinderAdapter(mockProductRepo, mockOptionRepo, mockCategoryRepo)

	assert.NotNil(t, adapter)
	assert.IsType(t, &ProductFinderAdapter{}, adapter)
}

func TestProductFinderAdapter_FindProduct(t *testing.T) {
	tests := []struct {
		name        string
		productID   shared.ID
		setupMock   func(*MockProductRepository)
		wantProduct *product.Product
		wantErr     bool
		errContains string
	}{
		{
			name:      "find existing product",
			productID: shared.ID(123),
			setupMock: func(m *MockProductRepository) {
				prod := &product.Product{
					ID:     shared.ID(123),
					ShopID: 456,
					Name:   "测试商品",
					Stock:  10,
				}
				m.On("FindByID", shared.ID(123)).Return(prod, nil)
			},
			wantProduct: &product.Product{
				ID:     shared.ID(123),
				ShopID: 456,
				Name:   "测试商品",
				Stock:  10,
			},
			wantErr: false,
		},
		{
			name:      "product not found",
			productID: shared.ID(999),
			setupMock: func(m *MockProductRepository) {
				m.On("FindByID", shared.ID(999)).Return(nil, errors.New("not found"))
			},
			wantProduct: nil,
			wantErr:     true,
			errContains: "not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockProductRepo := new(MockProductRepository)
			mockOptionRepo := new(MockProductOptionRepository)
			mockCategoryRepo := new(MockProductOptionCategoryRepository)

			tt.setupMock(mockProductRepo)

			adapter := NewProductFinderAdapter(mockProductRepo, mockOptionRepo, mockCategoryRepo)
			got, err := adapter.FindProduct(tt.productID)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantProduct.ID, got.ID)
				assert.Equal(t, tt.wantProduct.ShopID, got.ShopID)
				assert.Equal(t, tt.wantProduct.Name, got.Name)
			}

			mockProductRepo.AssertExpectations(t)
		})
	}
}

func TestProductFinderAdapter_FindOption(t *testing.T) {
	tests := []struct {
		name        string
		optionID    shared.ID
		setupMock   func(*MockProductOptionRepository)
		wantOption  *product.ProductOption
		wantErr     bool
		errContains string
	}{
		{
			name:     "find existing option",
			optionID: shared.ID(1),
			setupMock: func(m *MockProductOptionRepository) {
				opt := &product.ProductOption{
					ID:              shared.ID(1),
					CategoryID:      shared.ID(10),
					Name:            "大",
					PriceAdjustment: 10,
				}
				m.On("FindByID", shared.ID(1)).Return(opt, nil)
			},
			wantOption: &product.ProductOption{
				ID:              shared.ID(1),
				CategoryID:      shared.ID(10),
				Name:            "大",
				PriceAdjustment: 10,
			},
			wantErr: false,
		},
		{
			name:      "option not found",
			optionID:  shared.ID(999),
			setupMock: func(m *MockProductOptionRepository) {
				m.On("FindByID", shared.ID(999)).Return(nil, errors.New("option not found"))
			},
			wantOption:  nil,
			wantErr:     true,
			errContains: "option not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockProductRepo := new(MockProductRepository)
			mockOptionRepo := new(MockProductOptionRepository)
			mockCategoryRepo := new(MockProductOptionCategoryRepository)

			tt.setupMock(mockOptionRepo)

			adapter := NewProductFinderAdapter(mockProductRepo, mockOptionRepo, mockCategoryRepo)
			got, err := adapter.FindOption(tt.optionID)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantOption.ID, got.ID)
				assert.Equal(t, tt.wantOption.CategoryID, got.CategoryID)
				assert.Equal(t, tt.wantOption.Name, got.Name)
			}

			mockOptionRepo.AssertExpectations(t)
		})
	}
}

func TestProductFinderAdapter_FindOptionCategory(t *testing.T) {
	tests := []struct {
		name            string
		categoryID      shared.ID
		setupMock       func(*MockProductOptionCategoryRepository)
		wantCategory    *product.ProductOptionCategory
		wantErr         bool
		errContains     string
	}{
		{
			name:     "find existing category",
			categoryID: shared.ID(10),
			setupMock: func(m *MockProductOptionCategoryRepository) {
				cat := &product.ProductOptionCategory{
					ID:     shared.ID(10),
					Name:   "尺寸",
				}
				m.On("FindByID", shared.ID(10)).Return(cat, nil)
			},
			wantCategory: &product.ProductOptionCategory{
				ID:     shared.ID(10),
				Name:   "尺寸",
			},
			wantErr: false,
		},
		{
			name:       "category not found",
			categoryID: shared.ID(999),
			setupMock: func(m *MockProductOptionCategoryRepository) {
				m.On("FindByID", shared.ID(999)).Return(nil, errors.New("category not found"))
			},
			wantCategory: nil,
			wantErr:      true,
			errContains:  "category not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockProductRepo := new(MockProductRepository)
			mockOptionRepo := new(MockProductOptionRepository)
			mockCategoryRepo := new(MockProductOptionCategoryRepository)

			tt.setupMock(mockCategoryRepo)

			adapter := NewProductFinderAdapter(mockProductRepo, mockOptionRepo, mockCategoryRepo)
			got, err := adapter.FindOptionCategory(tt.categoryID)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantCategory.ID, got.ID)
				assert.Equal(t, tt.wantCategory.Name, got.Name)
			}

			mockCategoryRepo.AssertExpectations(t)
		})
	}
}

func TestProductFinderAdapter_AllOperations(t *testing.T) {
	// Integration-style test: use finder to find product, option, and category
	mockProductRepo := new(MockProductRepository)
	mockOptionRepo := new(MockProductOptionRepository)
	mockCategoryRepo := new(MockProductOptionCategoryRepository)

	// Setup product mock
	mockProductRepo.On("FindByID", shared.ID(1)).Return(&product.Product{
		ID:     shared.ID(1),
		ShopID: 456,
		Name:   "商品1",
		Stock:  10,
	}, nil)

	// Setup option mock
	mockOptionRepo.On("FindByID", shared.ID(10)).Return(&product.ProductOption{
		ID:              shared.ID(10),
		CategoryID:      shared.ID(100),
		Name:            "大",
		PriceAdjustment: 10,
	}, nil)

	// Setup category mock
	mockCategoryRepo.On("FindByID", shared.ID(100)).Return(&product.ProductOptionCategory{
		ID:   shared.ID(100),
		Name: "尺寸",
	}, nil)

	adapter := NewProductFinderAdapter(mockProductRepo, mockOptionRepo, mockCategoryRepo)

	// Find product
	prod, err := adapter.FindProduct(shared.ID(1))
	assert.NoError(t, err)
	assert.Equal(t, shared.ID(1), prod.ID)
	assert.Equal(t, "商品1", prod.Name)

	// Find option
	opt, err := adapter.FindOption(shared.ID(10))
	assert.NoError(t, err)
	assert.Equal(t, shared.ID(10), opt.ID)
	assert.Equal(t, "大", opt.Name)

	// Find category
	cat, err := adapter.FindOptionCategory(shared.ID(100))
	assert.NoError(t, err)
	assert.Equal(t, shared.ID(100), cat.ID)
	assert.Equal(t, "尺寸", cat.Name)

	mockProductRepo.AssertExpectations(t)
	mockOptionRepo.AssertExpectations(t)
	mockCategoryRepo.AssertExpectations(t)
}
