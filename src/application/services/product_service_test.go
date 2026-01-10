package services

import (
	"errors"
	"testing"

	"orderease/application/dto"
	"orderease/domain/product"
	"orderease/domain/shared"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockProductTagRepository is a mock for product.ProductTagRepository
type MockProductTagRepository struct {
	mock.Mock
}

func (m *MockProductTagRepository) Save(productID shared.ID, tagID int) error {
	args := m.Called(productID, tagID)
	return args.Error(0)
}

func (m *MockProductTagRepository) FindByProductID(productID shared.ID) ([]int, error) {
	args := m.Called(productID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]int), args.Error(1)
}

func (m *MockProductTagRepository) FindByTagID(tagID int) ([]shared.ID, error) {
	args := m.Called(tagID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]shared.ID), args.Error(1)
}

func (m *MockProductTagRepository) DeleteByProductID(productID shared.ID) error {
	args := m.Called(productID)
	return args.Error(0)
}

func TestNewProductService(t *testing.T) {
	mockProductRepo := new(MockProductRepository)
	mockProductCategoryRepo := new(MockProductOptionCategoryRepository)
	mockProductOptionRepo := new(MockProductOptionRepository)
	mockProductTagRepo := new(MockProductTagRepository)

	// Use nil for DB since we're testing methods that don't require DB operations
	service := NewProductService(
		mockProductRepo,
		mockProductCategoryRepo,
		mockProductOptionRepo,
		mockProductTagRepo,
		nil,
	)

	assert.NotNil(t, service)
}

func TestProductService_GetProduct(t *testing.T) {
	mockProductRepo := new(MockProductRepository)
	mockProductCategoryRepo := new(MockProductOptionCategoryRepository)
	mockProductOptionRepo := new(MockProductOptionRepository)
	mockProductTagRepo := new(MockProductTagRepository)

	service := NewProductService(
		mockProductRepo,
		mockProductCategoryRepo,
		mockProductOptionRepo,
		mockProductTagRepo,
		nil,
	)

	tests := []struct {
		name        string
		productID   shared.ID
		shopID      uint64
		setupMock   func()
		wantErr     bool
		errContains string
		validate    func(*testing.T, *dto.ProductResponse)
	}{
		{
			name:      "get existing product",
			productID: shared.ID(123),
			shopID:    456,
			setupMock: func() {
				prod := &product.Product{
					ID:          shared.ID(123),
					ShopID:      456,
					Name:        "测试商品",
					Description: "这是一个测试商品",
					Price:       shared.Price(100),
					Stock:       10,
					ImageURL:    "test.jpg",
					Status:      product.ProductStatusOnline,
					OptionCategories: []product.ProductOptionCategory{
						{
							ID:     shared.ID(1),
							Name:   "尺寸",
							Options: []product.ProductOption{
								{ID: shared.ID(10), Name: "大", PriceAdjustment: 10},
							},
						},
					},
				}
				mockProductRepo.On("FindByIDAndShopID", shared.ID(123), uint64(456)).Return(prod, nil)
			},
			wantErr: false,
			validate: func(t *testing.T, resp *dto.ProductResponse) {
				assert.Equal(t, shared.ID(123), resp.ID)
				assert.Equal(t, uint64(456), resp.ShopID)
				assert.Equal(t, "测试商品", resp.Name)
				assert.Equal(t, "这是一个测试商品", resp.Description)
				assert.Equal(t, shared.Price(100), resp.Price)
				assert.Equal(t, 10, resp.Stock)
				assert.Equal(t, "test.jpg", resp.ImageURL)
				assert.Equal(t, product.ProductStatusOnline, resp.Status)
				assert.Len(t, resp.OptionCategories, 1)
				assert.Equal(t, "尺寸", resp.OptionCategories[0].Name)
				assert.Len(t, resp.OptionCategories[0].Options, 1)
				assert.Equal(t, "大", resp.OptionCategories[0].Options[0].Name)
			},
		},
		{
			name:      "product not found",
			productID: shared.ID(999),
			shopID:    456,
			setupMock: func() {
				mockProductRepo.On("FindByIDAndShopID", shared.ID(999), uint64(456)).Return(nil, errors.New("not found"))
			},
			wantErr:     true,
			errContains: "not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			got, err := service.GetProduct(tt.productID, tt.shopID)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, got)
				if tt.validate != nil {
					tt.validate(t, got)
				}
			}

			mockProductRepo.AssertExpectations(t)
		})
	}
}

func TestProductService_GetProducts(t *testing.T) {
	mockProductRepo := new(MockProductRepository)
	mockProductCategoryRepo := new(MockProductOptionCategoryRepository)
	mockProductOptionRepo := new(MockProductOptionRepository)
	mockProductTagRepo := new(MockProductTagRepository)

	service := NewProductService(
		mockProductRepo,
		mockProductCategoryRepo,
		mockProductOptionRepo,
		mockProductTagRepo,
		nil,
	)

	tests := []struct {
		name        string
		shopID      uint64
		page        int
		pageSize    int
		search      string
		setupMock   func()
		wantErr     bool
		validate    func(*testing.T, *dto.ProductListResponse)
	}{
		{
			name:     "get products successfully",
			shopID:   456,
			page:     1,
			pageSize: 10,
			search:   "测试",
			setupMock: func() {
				products := []product.Product{
					{ID: shared.ID(1), ShopID: 456, Name: "测试商品1", Price: shared.Price(100)},
					{ID: shared.ID(2), ShopID: 456, Name: "测试商品2", Price: shared.Price(200)},
				}
				mockProductRepo.On("FindByShopID", uint64(456), 1, 10, "测试", true).Return(products, int64(2), nil)
			},
			wantErr: false,
			validate: func(t *testing.T, resp *dto.ProductListResponse) {
				assert.Equal(t, int64(2), resp.Total)
				assert.Equal(t, 1, resp.Page)
				assert.Equal(t, 10, resp.PageSize)
				assert.Len(t, resp.Data, 2)
				assert.Equal(t, shared.ID(1), resp.Data[0].ID)
				assert.Equal(t, shared.ID(2), resp.Data[1].ID)
			},
		},
		{
			name:     "empty products",
			shopID:   456,
			page:     1,
			pageSize: 10,
			search:   "",
			setupMock: func() {
				mockProductRepo.On("FindByShopID", uint64(456), 1, 10, "", true).Return([]product.Product{}, int64(0), nil)
			},
			wantErr: false,
			validate: func(t *testing.T, resp *dto.ProductListResponse) {
				assert.Equal(t, int64(0), resp.Total)
				assert.Len(t, resp.Data, 0)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			got, err := service.GetProducts(tt.shopID, tt.page, tt.pageSize, tt.search)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, got)
				if tt.validate != nil {
					tt.validate(t, got)
				}
			}

			mockProductRepo.AssertExpectations(t)
		})
	}
}

func TestProductService_UpdateProductStatus(t *testing.T) {
	mockProductRepo := new(MockProductRepository)
	mockProductCategoryRepo := new(MockProductOptionCategoryRepository)
	mockProductOptionRepo := new(MockProductOptionRepository)
	mockProductTagRepo := new(MockProductTagRepository)

	service := NewProductService(
		mockProductRepo,
		mockProductCategoryRepo,
		mockProductOptionRepo,
		mockProductTagRepo,
		nil,
	)

	tests := []struct {
		name        string
		req         *dto.UpdateProductStatusRequest
		shopID      uint64
		setupMock   func()
		wantErr     bool
		errContains string
	}{
		{
			name: "valid status change",
			req: &dto.UpdateProductStatusRequest{
				ID:     shared.ID(123),
				Status: product.ProductStatusOnline,
			},
			shopID: 456,
			setupMock: func() {
				prod := &product.Product{
					ID:     shared.ID(123),
					ShopID: 456,
					Status: product.ProductStatusPending,
				}
				mockProductRepo.On("FindByIDAndShopID", shared.ID(123), uint64(456)).Return(prod, nil).Once()
				mockProductRepo.On("Update", prod).Return(nil).Once()
			},
			wantErr: false,
		},
		{
			name: "product not found",
			req: &dto.UpdateProductStatusRequest{
				ID:     shared.ID(999),
				Status: product.ProductStatusOnline,
			},
			shopID: 456,
			setupMock: func() {
				mockProductRepo.On("FindByIDAndShopID", shared.ID(999), uint64(456)).Return(nil, errors.New("not found"))
			},
			wantErr:     true,
			errContains: "not found",
		},
		{
			name: "invalid status transition",
			req: &dto.UpdateProductStatusRequest{
				ID:     shared.ID(123),
				Status: product.ProductStatusPending,
			},
			shopID: 456,
			setupMock: func() {
				prod := &product.Product{
					ID:     shared.ID(123),
					ShopID: 456,
					Status: product.ProductStatusOnline,
				}
				mockProductRepo.On("FindByIDAndShopID", shared.ID(123), uint64(456)).Return(prod, nil).Once()
			},
			wantErr:     true,
			errContains: "不允许的状态转换",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			err := service.UpdateProductStatus(tt.req, tt.shopID)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
			} else {
				assert.NoError(t, err)
			}

			mockProductRepo.AssertExpectations(t)
		})
	}
}

func TestProductService_toProductResponse(t *testing.T) {
	mockProductRepo := new(MockProductRepository)
	mockProductCategoryRepo := new(MockProductOptionCategoryRepository)
	mockProductOptionRepo := new(MockProductOptionRepository)
	mockProductTagRepo := new(MockProductTagRepository)

	service := NewProductService(
		mockProductRepo,
		mockProductCategoryRepo,
		mockProductOptionRepo,
		mockProductTagRepo,
		nil,
	)

	tests := []struct {
		name     string
		product  *product.Product
		validate func(*testing.T, *dto.ProductResponse)
	}{
		{
			name: "product with option categories",
			product: &product.Product{
				ID:          shared.ID(123),
				ShopID:      456,
				Name:        "测试商品",
				Description: "测试描述",
				Price:       shared.Price(100),
				Stock:       10,
				ImageURL:    "test.jpg",
				Status:      product.ProductStatusOnline,
				OptionCategories: []product.ProductOptionCategory{
					{
						ID:         shared.ID(1),
						ProductID:  shared.ID(123),
						Name:       "尺寸",
						IsRequired: true,
						IsMultiple: false,
						Options: []product.ProductOption{
							{
								ID:              shared.ID(10),
								CategoryID:      shared.ID(1),
								Name:            "大",
								PriceAdjustment: 10,
								IsDefault:       true,
							},
							{
								ID:              shared.ID(11),
								CategoryID:      shared.ID(1),
								Name:            "小",
								PriceAdjustment: 0,
								IsDefault:       false,
							},
						},
					},
					{
						ID:         shared.ID(2),
						ProductID:  shared.ID(123),
						Name:       "温度",
						IsRequired: false,
						IsMultiple: true,
					},
				},
			},
			validate: func(t *testing.T, resp *dto.ProductResponse) {
				assert.Equal(t, shared.ID(123), resp.ID)
				assert.Equal(t, uint64(456), resp.ShopID)
				assert.Equal(t, "测试商品", resp.Name)
				assert.Equal(t, "测试描述", resp.Description)
				assert.Equal(t, shared.Price(100), resp.Price)
				assert.Equal(t, 10, resp.Stock)
				assert.Equal(t, "test.jpg", resp.ImageURL)
				assert.Equal(t, product.ProductStatusOnline, resp.Status)
				assert.Len(t, resp.OptionCategories, 2)

				// First category
				assert.Equal(t, shared.ID(1), resp.OptionCategories[0].ID)
				assert.Equal(t, "尺寸", resp.OptionCategories[0].Name)
				assert.True(t, resp.OptionCategories[0].IsRequired)
				assert.False(t, resp.OptionCategories[0].IsMultiple)
				assert.Len(t, resp.OptionCategories[0].Options, 2)
				assert.Equal(t, "大", resp.OptionCategories[0].Options[0].Name)
				assert.Equal(t, 10.0, resp.OptionCategories[0].Options[0].PriceAdjustment)
				assert.True(t, resp.OptionCategories[0].Options[0].IsDefault)

				// Second category
				assert.Equal(t, shared.ID(2), resp.OptionCategories[1].ID)
				assert.Equal(t, "温度", resp.OptionCategories[1].Name)
				assert.False(t, resp.OptionCategories[1].IsRequired)
				assert.True(t, resp.OptionCategories[1].IsMultiple)
			},
		},
		{
			name: "product without option categories",
			product: &product.Product{
				ID:               shared.ID(456),
				ShopID:           789,
				Name:             "简单商品",
				Description:      "",
				Price:            shared.Price(50),
				Stock:            5,
				ImageURL:         "",
				Status:           product.ProductStatusPending,
				OptionCategories: []product.ProductOptionCategory{},
			},
			validate: func(t *testing.T, resp *dto.ProductResponse) {
				assert.Equal(t, shared.ID(456), resp.ID)
				assert.Equal(t, "简单商品", resp.Name)
				assert.Empty(t, resp.Description)
				assert.Equal(t, shared.Price(50), resp.Price)
				assert.Len(t, resp.OptionCategories, 0)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := service.toProductResponse(tt.product)
			assert.NotNil(t, got)
			if tt.validate != nil {
				tt.validate(t, got)
			}
		})
	}
}
