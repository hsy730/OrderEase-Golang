package repositories

import (
	"database/sql"
	"fmt"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/bwmarrin/snowflake"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"orderease/models"
)

// setupTestDB 创建测试用的 mock 数据库
func setupTestDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock, *sql.DB) {
	t.Helper()

	sqlDB, mock, err := sqlmock.New()
	assert.NoError(t, err)

	dialector := mysql.New(mysql.Config{
		Conn:                      sqlDB,
		SkipInitializeWithVersion: true,
	})

	db, err := gorm.Open(dialector, &gorm.Config{})
	assert.NoError(t, err)

	return db, mock, sqlDB
}

// ==================== GetProductByID Tests ====================

func TestProductRepository_GetProductByID(t *testing.T) {
	tests := []struct {
		name        string
		productID   uint64
		shopID      snowflake.ID
		setupMock   func(mock sqlmock.Sqlmock)
		expectedErr bool
		errMsg      string
		validate    func(t *testing.T, product *models.Product)
	}{
		{
			name:      "successfully get product",
			productID: 123,
			shopID:    456,
			setupMock: func(mock sqlmock.Sqlmock) {
				// Mock product query
				rows := sqlmock.NewRows([]string{"id", "name", "shop_id", "price", "stock", "status"}).
					AddRow(123, "Test Product", 456, 10000, 50, "online")
				mock.ExpectQuery("SELECT \\* FROM `products`").
					WithArgs(456, int64(123), 1).
					WillReturnRows(rows)

				// Mock preload categories query
				categoryRows := sqlmock.NewRows([]string{"id", "name", "product_id"}).
					AddRow(1, "Size", 123)
				mock.ExpectQuery("SELECT \\* FROM `product_option_categories`").
					WithArgs(int64(123)).
					WillReturnRows(categoryRows)

				// Mock preload options query
				optionRows := sqlmock.NewRows([]string{"id", "name", "category_id", "price_adjustment"}).
					AddRow(1, "Large", 1, 500)
				mock.ExpectQuery("SELECT \\* FROM `product_options`").
					WithArgs(int64(1)).
					WillReturnRows(optionRows)
			},
			expectedErr: false,
			validate: func(t *testing.T, product *models.Product) {
				assert.Equal(t, snowflake.ID(123), product.ID)
				assert.Equal(t, "Test Product", product.Name)
				assert.Equal(t, snowflake.ID(456), product.ShopID)
			},
		},
		{
			name:      "product not found",
			productID: 999,
			shopID:    456,
			setupMock: func(mock sqlmock.Sqlmock) {
				// Mock product not found
				mock.ExpectQuery("SELECT \\* FROM `products`").
					WithArgs(456, int64(999), 1).
					WillReturnError(gorm.ErrRecordNotFound)
			},
			expectedErr: true,
			errMsg:      "商品不存在",
		},
		{
			name:      "database error",
			productID: 123,
			shopID:    456,
			setupMock: func(mock sqlmock.Sqlmock) {
				// Mock database error
				mock.ExpectQuery("SELECT \\* FROM `products`").
					WithArgs(456, int64(123), 1).
					WillReturnError(fmt.Errorf("database connection error"))
			},
			expectedErr: true,
			errMsg:      "服务器内部错误",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, sqlDB := setupTestDB(t)
			defer sqlDB.Close()

			repo := NewProductRepository(db)
			tt.setupMock(mock)

			product, err := repo.GetProductByID(tt.productID, tt.shopID)

			if tt.expectedErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, product)
				if tt.validate != nil {
					tt.validate(t, product)
				}
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

// ==================== GetProductsByIDs Tests ====================

func TestProductRepository_GetProductsByIDs(t *testing.T) {
	tests := []struct {
		name        string
		ids         []snowflake.ID
		shopID      snowflake.ID
		setupMock   func(mock sqlmock.Sqlmock)
		expectedErr bool
		validate    func(t *testing.T, products []models.Product)
	}{
		{
			name:   "successfully get products",
			ids:    []snowflake.ID{123, 456, 789},
			shopID: 100,
			setupMock: func(mock sqlmock.Sqlmock) {
				// Mock products query - GORM expands IN clause individually
				rows := sqlmock.NewRows([]string{"id"}).
					AddRow(123).
					AddRow(456).
					AddRow(789)
				mock.ExpectQuery("SELECT `id` FROM `products`").
					WithArgs(int64(123), int64(456), int64(789), int64(100)).
					WillReturnRows(rows)
			},
			expectedErr: false,
			validate: func(t *testing.T, products []models.Product) {
				assert.Len(t, products, 3)
				assert.Equal(t, snowflake.ID(123), products[0].ID)
			},
		},
		{
			name:   "single product ID",
			ids:    []snowflake.ID{123},
			shopID: 100,
			setupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id"}).
					AddRow(123)
				mock.ExpectQuery("SELECT `id` FROM `products`").
					WithArgs(int64(123), int64(100)).
					WillReturnRows(rows)
			},
			expectedErr: false,
			validate: func(t *testing.T, products []models.Product) {
				assert.Len(t, products, 1)
				assert.Equal(t, snowflake.ID(123), products[0].ID)
			},
		},
		{
			name:   "database error",
			ids:    []snowflake.ID{123},
			shopID: 100,
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT `id` FROM `products`").
					WithArgs(int64(123), int64(100)).
					WillReturnError(fmt.Errorf("database error"))
			},
			expectedErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, sqlDB := setupTestDB(t)
			defer sqlDB.Close()

			repo := NewProductRepository(db)
			tt.setupMock(mock)

			products, err := repo.GetProductsByIDs(tt.ids, tt.shopID)

			if tt.expectedErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tt.validate != nil {
					tt.validate(t, products)
				}
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

// ==================== CheckShopExists Tests ====================

func TestProductRepository_CheckShopExists(t *testing.T) {
	tests := []struct {
		name          string
		shopID        snowflake.ID
		setupMock     func(mock sqlmock.Sqlmock)
		expectedErr   bool
		expectedExists bool
		errMsg        string
	}{
		{
			name:   "shop exists",
			shopID: 123,
			setupMock: func(mock sqlmock.Sqlmock) {
				// Mock count query
				rows := sqlmock.NewRows([]string{"count"}).AddRow(1)
				mock.ExpectQuery("SELECT count\\(\\*\\) FROM `shops`").
					WithArgs(int64(123)).
					WillReturnRows(rows)
			},
			expectedErr:    false,
			expectedExists: true,
		},
		{
			name:   "shop does not exist",
			shopID: 999,
			setupMock: func(mock sqlmock.Sqlmock) {
				// Mock count query
				rows := sqlmock.NewRows([]string{"count"}).AddRow(0)
				mock.ExpectQuery("SELECT count\\(\\*\\) FROM `shops`").
					WithArgs(int64(999)).
					WillReturnRows(rows)
			},
			expectedErr:    false,
			expectedExists: false,
		},
		{
			name:   "database error",
			shopID: 123,
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT count\\(\\*\\) FROM `shops`").
					WithArgs(int64(123)).
					WillReturnError(fmt.Errorf("database error"))
			},
			expectedErr: true,
			errMsg:      "店铺校验失败",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, sqlDB := setupTestDB(t)
			defer sqlDB.Close()

			repo := NewProductRepository(db)
			tt.setupMock(mock)

			exists, err := repo.CheckShopExists(tt.shopID)

			if tt.expectedErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedExists, exists)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

// ==================== UpdateStatus Tests ====================

func TestProductRepository_UpdateStatus(t *testing.T) {
	tests := []struct {
		name        string
		productID   uint64
		shopID      snowflake.ID
		status      string
		setupMock   func(mock sqlmock.Sqlmock)
		expectedErr bool
		errMsg      string
	}{
		{
			name:      "successfully update status",
			productID: 123,
			shopID:    456,
			status:    "online",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec("UPDATE `products`").
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			expectedErr: false,
		},
		{
			name:      "product not found",
			productID: 999,
			shopID:    456,
			status:    "online",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec("UPDATE `products`").
					WillReturnResult(sqlmock.NewResult(1, 0)) // RowsAffected = 0
				mock.ExpectCommit()
			},
			expectedErr: true,
			errMsg:      "商品不存在",
		},
		{
			name:      "database error",
			productID: 123,
			shopID:    456,
			status:    "online",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec("UPDATE `products`").
					WillReturnError(fmt.Errorf("database error"))
				mock.ExpectRollback()
			},
			expectedErr: true,
			errMsg:      "更新商品状态失败",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, sqlDB := setupTestDB(t)
			defer sqlDB.Close()

			repo := NewProductRepository(db)
			tt.setupMock(mock)

			err := repo.UpdateStatus(tt.productID, tt.shopID, tt.status)

			if tt.expectedErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

// ==================== UpdateImageURL Tests ====================

func TestProductRepository_UpdateImageURL(t *testing.T) {
	tests := []struct {
		name        string
		productID   uint64
		shopID      snowflake.ID
		imageURL    string
		setupMock   func(mock sqlmock.Sqlmock)
		expectedErr bool
		errMsg      string
	}{
		{
			name:      "successfully update image URL",
			productID: 123,
			shopID:    456,
			imageURL:  "http://example.com/image.jpg",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec("UPDATE `products`").
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			expectedErr: false,
		},
		{
			name:      "product not found",
			productID: 999,
			shopID:    456,
			imageURL:  "http://example.com/image.jpg",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec("UPDATE `products`").
					WillReturnResult(sqlmock.NewResult(1, 0))
				mock.ExpectCommit()
			},
			expectedErr: true,
			errMsg:      "商品不存在",
		},
		{
			name:      "database error",
			productID: 123,
			shopID:    456,
			imageURL:  "http://example.com/image.jpg",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec("UPDATE `products`").
					WillReturnError(fmt.Errorf("database error"))
				mock.ExpectRollback()
			},
			expectedErr: true,
			errMsg:      "更新商品图片失败",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, sqlDB := setupTestDB(t)
			defer sqlDB.Close()

			repo := NewProductRepository(db)
			tt.setupMock(mock)

			err := repo.UpdateImageURL(tt.productID, tt.shopID, tt.imageURL)

			if tt.expectedErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

// ==================== CreateWithCategories Tests ====================

func TestProductRepository_CreateWithCategories(t *testing.T) {
	tests := []struct {
		name        string
		product     *models.Product
		categories  []models.ProductOptionCategory
		setupMock   func(mock sqlmock.Sqlmock)
		expectedErr bool
		errMsg      string
	}{
		{
			name: "successfully create product with categories",
			product: &models.Product{
				ID:   123,
				Name: "Test Product",
			},
			categories: []models.ProductOptionCategory{
				{ID: 1, Name: "Size"},
				{ID: 2, Name: "Color"},
			},
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				// Mock create product
				mock.ExpectExec("INSERT INTO `products`").
					WillReturnResult(sqlmock.NewResult(123, 1))
				// Mock create categories
				mock.ExpectExec("INSERT INTO `product_option_categories`").
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectExec("INSERT INTO `product_option_categories`").
					WillReturnResult(sqlmock.NewResult(2, 1))
				mock.ExpectCommit()
			},
			expectedErr: false,
		},
		{
			name: "create product fails",
			product: &models.Product{
				Name: "Test Product",
			},
			categories: []models.ProductOptionCategory{
				{Name: "Size"},
			},
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO `products`").
					WillReturnError(fmt.Errorf("database error"))
				mock.ExpectRollback()
			},
			expectedErr: true,
			errMsg:      "创建商品失败",
		},
		{
			name: "create category fails after product created",
			product: &models.Product{
				ID:   123,
				Name: "Test Product",
			},
			categories: []models.ProductOptionCategory{
				{Name: "Size"},
			},
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO `products`").
					WillReturnResult(sqlmock.NewResult(123, 1))
				mock.ExpectExec("INSERT INTO `product_option_categories`").
					WillReturnError(fmt.Errorf("database error"))
				mock.ExpectRollback()
			},
			expectedErr: true,
			errMsg:      "创建商品参数失败",
		},
		{
			name: "commit fails",
			product: &models.Product{
				ID:   123,
				Name: "Test Product",
			},
			categories: []models.ProductOptionCategory{
				{Name: "Size"},
			},
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO `products`").
					WillReturnResult(sqlmock.NewResult(123, 1))
				mock.ExpectExec("INSERT INTO `product_option_categories`").
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit().WillReturnError(fmt.Errorf("commit error"))
			},
			expectedErr: true,
			errMsg:      "创建商品失败",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, sqlDB := setupTestDB(t)
			defer sqlDB.Close()

			repo := NewProductRepository(db)
			tt.setupMock(mock)

			err := repo.CreateWithCategories(tt.product, tt.categories)

			if tt.expectedErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

// ==================== UpdateWithCategories Tests ====================

func TestProductRepository_UpdateWithCategories(t *testing.T) {
	tests := []struct {
		name        string
		product     *models.Product
		categories  []models.ProductOptionCategory
		setupMock   func(mock sqlmock.Sqlmock)
		expectedErr bool
		errMsg      string
	}{
		{
			name: "successfully update product with categories",
			product: &models.Product{
				ID:   123,
				Name: "Updated Product",
			},
			categories: []models.ProductOptionCategory{
				{Name: "New Size"},
			},
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec("UPDATE `products`").
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectExec("DELETE FROM `product_option_categories`").
					WillReturnResult(sqlmock.NewResult(1, 2))
				mock.ExpectExec("INSERT INTO `product_option_categories`").
					WillReturnResult(sqlmock.NewResult(3, 1))
				mock.ExpectCommit()
			},
			expectedErr: false,
		},
		{
			name: "update product fails",
			product: &models.Product{
				ID:   123,
				Name: "Updated Product",
			},
			categories: []models.ProductOptionCategory{
				{Name: "New Size"},
			},
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec("UPDATE `products`").
					WillReturnError(fmt.Errorf("database error"))
				mock.ExpectRollback()
			},
			expectedErr: true,
			errMsg:      "更新商品失败",
		},
		{
			name: "delete old categories fails",
			product: &models.Product{
				ID:   123,
				Name: "Updated Product",
			},
			categories: []models.ProductOptionCategory{
				{Name: "New Size"},
			},
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec("UPDATE `products`").
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectExec("DELETE FROM `product_option_categories`").
					WillReturnError(fmt.Errorf("database error"))
				mock.ExpectRollback()
			},
			expectedErr: true,
			errMsg:      "更新商品参数失败",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, sqlDB := setupTestDB(t)
			defer sqlDB.Close()

			repo := NewProductRepository(db)
			tt.setupMock(mock)

			err := repo.UpdateWithCategories(tt.product, tt.categories)

			if tt.expectedErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

// ==================== DeleteWithDependencies Tests ====================

func TestProductRepository_DeleteWithDependencies(t *testing.T) {
	tests := []struct {
		name        string
		productID   uint64
		shopID      snowflake.ID
		setupMock   func(mock sqlmock.Sqlmock)
		expectedErr bool
		errMsg      string
	}{
		{
			name:      "successfully delete product with dependencies",
			productID: 123,
			shopID:    456,
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec("DELETE FROM `product_options`").
					WillReturnResult(sqlmock.NewResult(1, 5))
				mock.ExpectExec("DELETE FROM `product_option_categories`").
					WillReturnResult(sqlmock.NewResult(1, 2))
				mock.ExpectExec("DELETE FROM `products`").
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			expectedErr: false,
		},
		{
			name:      "product not found",
			productID: 999,
			shopID:    456,
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec("DELETE FROM `product_options`").
					WillReturnResult(sqlmock.NewResult(1, 0))
				mock.ExpectExec("DELETE FROM `product_option_categories`").
					WillReturnResult(sqlmock.NewResult(1, 0))
				mock.ExpectExec("DELETE FROM `products`").
					WillReturnResult(sqlmock.NewResult(1, 0)) // RowsAffected = 0
				mock.ExpectRollback()
			},
			expectedErr: true,
			errMsg:      "商品不存在",
		},
		{
			name:      "delete options fails",
			productID: 123,
			shopID:    456,
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec("DELETE FROM `product_options`").
					WillReturnError(fmt.Errorf("database error"))
				mock.ExpectRollback()
			},
			expectedErr: true,
			errMsg:      "删除商品参数选项失败",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, sqlDB := setupTestDB(t)
			defer sqlDB.Close()

			repo := NewProductRepository(db)
			tt.setupMock(mock)

			err := repo.DeleteWithDependencies(tt.productID, tt.shopID)

			if tt.expectedErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

// ==================== GetProductsByShop Tests ====================

func TestProductRepository_GetProductsByShop(t *testing.T) {
	tests := []struct {
		name        string
		shopID      snowflake.ID
		page        int
		pageSize    int
		search      string
		onlyOnline  bool
		setupMock   func(mock sqlmock.Sqlmock)
		expectedErr bool
		validate    func(t *testing.T, result *ProductListResult)
	}{
		{
			name:       "get all products for admin",
			shopID:     123,
			page:       1,
			pageSize:   10,
			search:     "",
			onlyOnline: false,
			setupMock: func(mock sqlmock.Sqlmock) {
				// Mock count query
				countRows := sqlmock.NewRows([]string{"count"}).AddRow(25)
				mock.ExpectQuery("SELECT count\\(\\*\\) FROM `products`").
					WithArgs(int64(123)).
					WillReturnRows(countRows)

				// Mock find query
				productRows := sqlmock.NewRows([]string{"id", "name", "shop_id"}).
					AddRow(1, "Product 1", 123).
					AddRow(2, "Product 2", 123)
				mock.ExpectQuery("SELECT \\* FROM `products`").
					WithArgs(int64(123), 10).
					WillReturnRows(productRows)

				// Mock preload categories using IN clause
				categoryRows := sqlmock.NewRows([]string{"id", "name", "product_id"}).
					AddRow(1, "Size", 1).
					AddRow(2, "Color", 2)
				mock.ExpectQuery("SELECT \\* FROM `product_option_categories`").
					WithArgs(int64(1), int64(2)).
					WillReturnRows(categoryRows)

				// Mock preload options using IN clause
				optionRows := sqlmock.NewRows([]string{"id", "name", "category_id", "price_adjustment"}).
					AddRow(1, "Large", 1, 500).
					AddRow(2, "Red", 2, 0)
				mock.ExpectQuery("SELECT \\* FROM `product_options`").
					WithArgs(int64(1), int64(2)).
					WillReturnRows(optionRows)
			},
			expectedErr: false,
			validate: func(t *testing.T, result *ProductListResult) {
				assert.Equal(t, int64(25), result.Total)
				assert.Len(t, result.Products, 2)
			},
		},
		{
			name:       "get online products for user",
			shopID:     123,
			page:       1,
			pageSize:   10,
			search:     "",
			onlyOnline: true,
			setupMock: func(mock sqlmock.Sqlmock) {
				// Mock count query with status filter
				countRows := sqlmock.NewRows([]string{"count"}).AddRow(15)
				mock.ExpectQuery("SELECT count\\(\\*\\) FROM `products`").
					WithArgs(int64(123), "online").
					WillReturnRows(countRows)

				// Mock find query - offset=0 is omitted by GORM
				productRows := sqlmock.NewRows([]string{"id", "name", "shop_id"}).
					AddRow(1, "Product 1", 123)
				mock.ExpectQuery("SELECT \\* FROM `products`").
					WithArgs(int64(123), "online", 10).
					WillReturnRows(productRows)

				// Mock preload categories for product 1
				categoryRows := sqlmock.NewRows([]string{"id", "name", "product_id"}).
					AddRow(1, "Size", 1)
				mock.ExpectQuery("SELECT \\* FROM `product_option_categories`").
					WithArgs(int64(1)).
					WillReturnRows(categoryRows)

				// Mock preload options for category 1
				optionRows := sqlmock.NewRows([]string{"id", "name", "category_id", "price_adjustment"}).
					AddRow(1, "Large", 1, 500)
				mock.ExpectQuery("SELECT \\* FROM `product_options`").
					WithArgs(int64(1)).
					WillReturnRows(optionRows)
			},
			expectedErr: false,
			validate: func(t *testing.T, result *ProductListResult) {
				assert.Equal(t, int64(15), result.Total)
				assert.Len(t, result.Products, 1)
			},
		},
		{
			name:       "search products",
			shopID:     123,
			page:       1,
			pageSize:   10,
			search:     "Test",
			onlyOnline: false,
			setupMock: func(mock sqlmock.Sqlmock) {
				// Mock count query with search
				countRows := sqlmock.NewRows([]string{"count"}).AddRow(5)
				mock.ExpectQuery("SELECT count\\(\\*\\) FROM `products`").
					WithArgs(int64(123), "%Test%").
					WillReturnRows(countRows)

				// Mock find query - offset=0 is omitted by GORM
				productRows := sqlmock.NewRows([]string{"id", "name", "shop_id"}).
					AddRow(1, "Test Product", 123)
				mock.ExpectQuery("SELECT \\* FROM `products`").
					WithArgs(int64(123), "%Test%", 10).
					WillReturnRows(productRows)

				// Mock preload categories for product 1
				categoryRows := sqlmock.NewRows([]string{"id", "name", "product_id"}).
					AddRow(1, "Size", 1)
				mock.ExpectQuery("SELECT \\* FROM `product_option_categories`").
					WithArgs(int64(1)).
					WillReturnRows(categoryRows)

				// Mock preload options for category 1
				optionRows := sqlmock.NewRows([]string{"id", "name", "category_id", "price_adjustment"}).
					AddRow(1, "Large", 1, 500)
				mock.ExpectQuery("SELECT \\* FROM `product_options`").
					WithArgs(int64(1)).
					WillReturnRows(optionRows)
			},
			expectedErr: false,
			validate: func(t *testing.T, result *ProductListResult) {
				assert.Equal(t, int64(5), result.Total)
				assert.Len(t, result.Products, 1)
			},
		},
		{
			name:       "pagination - second page",
			shopID:     123,
			page:       2,
			pageSize:   10,
			search:     "",
			onlyOnline: false,
			setupMock: func(mock sqlmock.Sqlmock) {
				// Mock count query
				countRows := sqlmock.NewRows([]string{"count"}).AddRow(25)
				mock.ExpectQuery("SELECT count\\(\\*\\) FROM `products`").
					WithArgs(int64(123)).
					WillReturnRows(countRows)

				// Mock find query with offset (offset = (2-1)*10 = 10)
				productRows := sqlmock.NewRows([]string{"id", "name", "shop_id"}).
					AddRow(11, "Product 11", 123).
					AddRow(12, "Product 12", 123)
				mock.ExpectQuery("SELECT \\* FROM `products`").
					WithArgs(int64(123), 10, 10).
					WillReturnRows(productRows)

				// Mock preload categories for products 11 and 12
				categoryRows := sqlmock.NewRows([]string{"id", "name", "product_id"}).
					AddRow(11, "Size", 11).
					AddRow(12, "Color", 12)
				mock.ExpectQuery("SELECT \\* FROM `product_option_categories`").
					WithArgs(int64(11), int64(12)).
					WillReturnRows(categoryRows)

				// Mock preload options for categories 11 and 12
				optionRows := sqlmock.NewRows([]string{"id", "name", "category_id", "price_adjustment"}).
					AddRow(1, "Large", 11, 500).
					AddRow(2, "Red", 12, 0)
				mock.ExpectQuery("SELECT \\* FROM `product_options`").
					WithArgs(int64(11), int64(12)).
					WillReturnRows(optionRows)
			},
			expectedErr: false,
			validate: func(t *testing.T, result *ProductListResult) {
				assert.Equal(t, int64(25), result.Total)
				assert.Len(t, result.Products, 2)
			},
		},
		{
			name:       "database error on count",
			shopID:     123,
			page:       1,
			pageSize:   10,
			search:     "",
			onlyOnline: false,
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT count\\(\\*\\) FROM `products`").
					WithArgs(int64(123)).
					WillReturnError(fmt.Errorf("database error"))
			},
			expectedErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, sqlDB := setupTestDB(t)
			defer sqlDB.Close()

			repo := NewProductRepository(db)
			tt.setupMock(mock)

			result, err := repo.GetProductsByShop(tt.shopID, tt.page, tt.pageSize, tt.search, tt.onlyOnline)

			if tt.expectedErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				if tt.validate != nil {
					tt.validate(t, result)
				}
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
