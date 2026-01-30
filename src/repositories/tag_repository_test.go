package repositories

import (
	"fmt"
	"orderease/models"
	"sync"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/bwmarrin/snowflake"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

// TestTagRepository_Create 测试创建标签
func TestTagRepository_Create(t *testing.T) {
	tests := []struct {
		name        string
		tag         *models.Tag
		setupMock   func(mock sqlmock.Sqlmock)
		expectedErr bool
		errMsg      string
		validate    func(t *testing.T, tag *models.Tag)
	}{
		{
			name: "successfully create tag",
			tag: &models.Tag{
				Name:        "Test Tag",
				ShopID:      123,
				Description: "Test Description",
			},
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO `tags`").
					WithArgs(int64(123), "Test Tag", "Test Description", sqlmock.AnyArg(), sqlmock.AnyArg()).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			expectedErr: false,
			validate: func(t *testing.T, tag *models.Tag) {
				assert.Equal(t, "Test Tag", tag.Name)
				assert.Equal(t, snowflake.ID(123), tag.ShopID)
			},
		},
		{
			name: "create fails - database error",
			tag: &models.Tag{
				Name:   "Failed Tag",
				ShopID: 123,
			},
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO `tags`").
					WillReturnError(fmt.Errorf("database error"))
				mock.ExpectRollback()
			},
			expectedErr: true,
			errMsg:      "创建标签失败",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, sqlDB := setupTestDB(t)
			defer sqlDB.Close()

			tt.setupMock(mock)

			repo := NewTagRepository(db)
			err := repo.Create(tt.tag)

			if tt.expectedErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
				if tt.validate != nil {
					tt.validate(t, tt.tag)
				}
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

// TestTagRepository_Update 测试更新标签
func TestTagRepository_Update(t *testing.T) {
	tests := []struct {
		name        string
		tag         *models.Tag
		setupMock   func(mock sqlmock.Sqlmock)
		expectedErr bool
		errMsg      string
	}{
		{
			name: "successfully update tag",
			tag: &models.Tag{
				ID:          1,
				Name:        "Updated Tag",
				ShopID:      123,
				Description: "Updated Description",
			},
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec("UPDATE `tags`").
					WithArgs(int64(123), "Updated Tag", "Updated Description", sqlmock.AnyArg(), sqlmock.AnyArg(), int64(1)).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			expectedErr: false,
		},
		{
			name: "update fails - database error",
			tag: &models.Tag{
				ID:     1,
				Name:   "Failed Update",
				ShopID: 123,
			},
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec("UPDATE `tags`").
					WillReturnError(fmt.Errorf("database error"))
				mock.ExpectRollback()
			},
			expectedErr: true,
			errMsg:      "更新标签失败",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, sqlDB := setupTestDB(t)
			defer sqlDB.Close()

			tt.setupMock(mock)

			repo := NewTagRepository(db)
			err := repo.Update(tt.tag)

			if tt.expectedErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

// TestTagRepository_Delete 测试删除标签
func TestTagRepository_Delete(t *testing.T) {
	tests := []struct {
		name        string
		tag         *models.Tag
		setupMock   func(mock sqlmock.Sqlmock)
		expectedErr bool
		errMsg      string
	}{
		{
			name: "successfully delete tag",
			tag: &models.Tag{
				ID:     1,
				ShopID: 123,
			},
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec("DELETE FROM `tags`").
					WithArgs(int64(1)).
					WillReturnResult(sqlmock.NewResult(0, 1))
				mock.ExpectCommit()
			},
			expectedErr: false,
		},
		{
			name: "delete fails - database error",
			tag: &models.Tag{
				ID:     1,
				ShopID: 123,
			},
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec("DELETE FROM `tags`").
					WillReturnError(fmt.Errorf("database error"))
				mock.ExpectRollback()
			},
			expectedErr: true,
			errMsg:      "删除标签失败",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, sqlDB := setupTestDB(t)
			defer sqlDB.Close()

			tt.setupMock(mock)

			repo := NewTagRepository(db)
			err := repo.Delete(tt.tag)

			if tt.expectedErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

// TestTagRepository_GetByIDAndShopID 测试根据ID和店铺ID获取标签
func TestTagRepository_GetByIDAndShopID(t *testing.T) {
	tests := []struct {
		name        string
		tagID       int
		shopID      snowflake.ID
		setupMock   func(mock sqlmock.Sqlmock)
		expectedErr bool
		errMsg      string
		validate    func(t *testing.T, tag *models.Tag)
	}{
		{
			name:  "successfully get tag",
			tagID: 1,
			shopID: 123,
			setupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "name", "shop_id", "description", "created_at", "updated_at"}).
					AddRow(1, "Test Tag", int64(123), "Test Description", time.Now(), time.Now())
				mock.ExpectQuery("SELECT \\* FROM `tags` WHERE shop_id = \\? AND `tags`\\.\\`id\\` = \\? ORDER BY `tags`\\.\\`id\\` LIMIT \\?").
					WithArgs(int64(123), 1, 1).
					WillReturnRows(rows)
			},
			expectedErr: false,
			validate: func(t *testing.T, tag *models.Tag) {
				assert.NotNil(t, tag)
				assert.Equal(t, 1, tag.ID)
				assert.Equal(t, "Test Tag", tag.Name)
				assert.Equal(t, snowflake.ID(123), tag.ShopID)
			},
		},
		{
			name:  "tag not found",
			tagID: 999,
			shopID: 123,
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT \\* FROM `tags` WHERE shop_id = \\? AND `tags`\\.\\`id\\` = \\? ORDER BY `tags`\\.\\`id\\` LIMIT \\?").
					WithArgs(int64(123), 999, 1).
					WillReturnError(gorm.ErrRecordNotFound)
			},
			expectedErr: true,
			errMsg:      "标签不存在",
		},
		{
			name:  "database error",
			tagID: 1,
			shopID: 123,
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT \\* FROM `tags` WHERE shop_id = \\? AND `tags`\\.\\`id\\` = \\? ORDER BY `tags`\\.\\`id\\` LIMIT \\?").
					WithArgs(int64(123), 1, 1).
					WillReturnError(fmt.Errorf("database error"))
			},
			expectedErr: true,
			errMsg:      "查询标签失败",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, sqlDB := setupTestDB(t)
			defer sqlDB.Close()

			tt.setupMock(mock)

			repo := NewTagRepository(db)
			tag, err := repo.GetByIDAndShopID(tt.tagID, tt.shopID)

			if tt.expectedErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
				if tt.validate != nil {
					tt.validate(t, tag)
				}
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

// TestTagRepository_GetListByShopID 测试获取店铺的标签列表
func TestTagRepository_GetListByShopID(t *testing.T) {
	tests := []struct {
		name        string
		shopID      snowflake.ID
		setupMock   func(mock sqlmock.Sqlmock)
		expectedErr bool
		errMsg      string
		validate    func(t *testing.T, tags []models.Tag)
	}{
		{
			name:   "successfully get tag list",
			shopID: 123,
			setupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "name", "shop_id", "description", "created_at", "updated_at"}).
					AddRow(1, "Tag 1", int64(123), "Description 1", time.Now(), time.Now()).
					AddRow(2, "Tag 2", int64(123), "Description 2", time.Now(), time.Now())
				mock.ExpectQuery("SELECT \\* FROM `tags` WHERE shop_id = \\? ORDER BY created_at DESC").
					WithArgs(int64(123)).
					WillReturnRows(rows)
			},
			expectedErr: false,
			validate: func(t *testing.T, tags []models.Tag) {
				assert.Len(t, tags, 2)
				assert.Equal(t, "Tag 1", tags[0].Name)
				assert.Equal(t, "Tag 2", tags[1].Name)
			},
		},
		{
			name:   "empty tag list",
			shopID: 123,
			setupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "name", "shop_id", "description", "created_at", "updated_at"})
				mock.ExpectQuery("SELECT \\* FROM `tags` WHERE shop_id = \\? ORDER BY created_at DESC").
					WithArgs(int64(123)).
					WillReturnRows(rows)
			},
			expectedErr: false,
			validate: func(t *testing.T, tags []models.Tag) {
				assert.Empty(t, tags)
			},
		},
		{
			name:   "database error",
			shopID: 123,
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT \\* FROM `tags` WHERE shop_id = \\? ORDER BY created_at DESC").
					WithArgs(int64(123)).
					WillReturnError(fmt.Errorf("database error"))
			},
			expectedErr: true,
			errMsg:      "查询标签列表失败",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, sqlDB := setupTestDB(t)
			defer sqlDB.Close()

			tt.setupMock(mock)

			repo := NewTagRepository(db)
			tags, err := repo.GetListByShopID(tt.shopID)

			if tt.expectedErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
				if tt.validate != nil {
					tt.validate(t, tags)
				}
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

// TestTagRepository_GetUnboundProductsCount 测试获取未绑定任何标签的商品数量
func TestTagRepository_GetUnboundProductsCount(t *testing.T) {
	tests := []struct {
		name        string
		shopID      snowflake.ID
		setupMock   func(mock sqlmock.Sqlmock)
		expectedErr bool
		errMsg      string
		validate    func(t *testing.T, count int64)
	}{
		{
			name:   "successfully get count",
			shopID: 123,
			setupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"count"}).AddRow(5)
				mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM products").
					WithArgs(int64(123)).
					WillReturnRows(rows)
			},
			expectedErr: false,
			validate: func(t *testing.T, count int64) {
				assert.Equal(t, int64(5), count)
			},
		},
		{
			name:   "zero unbound products",
			shopID: 123,
			setupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"count"}).AddRow(0)
				mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM products").
					WithArgs(int64(123)).
					WillReturnRows(rows)
			},
			expectedErr: false,
			validate: func(t *testing.T, count int64) {
				assert.Equal(t, int64(0), count)
			},
		},
		{
			name:   "database error",
			shopID: 123,
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM products").
					WithArgs(int64(123)).
					WillReturnError(fmt.Errorf("database error"))
			},
			expectedErr: true,
			errMsg:      "查询未绑定商品数量失败",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, sqlDB := setupTestDB(t)
			defer sqlDB.Close()

			tt.setupMock(mock)

			repo := NewTagRepository(db)
			count, err := repo.GetUnboundProductsCount(tt.shopID)

			if tt.expectedErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
				if tt.validate != nil {
					tt.validate(t, count)
				}
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

// TestTagRepository_GetTagBoundProductIDs 测试获取已绑定到指定标签的商品ID列表
func TestTagRepository_GetTagBoundProductIDs(t *testing.T) {
	tests := []struct {
		name        string
		tagID       int
		shopID      snowflake.ID
		setupMock   func(mock sqlmock.Sqlmock)
		expectedErr bool
		errMsg      string
		validate    func(t *testing.T, productIDs []uint)
	}{
		{
			name:  "successfully get product IDs",
			tagID: 1,
			shopID: 123,
			setupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"product_id"}).
					AddRow(100).
					AddRow(200).
					AddRow(300)
				mock.ExpectQuery("SELECT product_id FROM product_tags").
					WithArgs(1, int64(123)).
					WillReturnRows(rows)
			},
			expectedErr: false,
			validate: func(t *testing.T, productIDs []uint) {
				assert.Len(t, productIDs, 3)
				assert.Equal(t, uint(100), productIDs[0])
				assert.Equal(t, uint(200), productIDs[1])
				assert.Equal(t, uint(300), productIDs[2])
			},
		},
		{
			name:  "empty product IDs",
			tagID: 1,
			shopID: 123,
			setupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"product_id"})
				mock.ExpectQuery("SELECT product_id FROM product_tags").
					WithArgs(1, int64(123)).
					WillReturnRows(rows)
			},
			expectedErr: false,
			validate: func(t *testing.T, productIDs []uint) {
				assert.Empty(t, productIDs)
			},
		},
		{
			name:  "database error",
			tagID: 1,
			shopID: 123,
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT product_id FROM product_tags").
					WithArgs(1, int64(123)).
					WillReturnError(fmt.Errorf("database error"))
			},
			expectedErr: true,
			errMsg:      "获取绑定商品ID列表失败",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, sqlDB := setupTestDB(t)
			defer sqlDB.Close()

			tt.setupMock(mock)

			repo := NewTagRepository(db)
			productIDs, err := repo.GetTagBoundProductIDs(tt.tagID, tt.shopID)

			if tt.expectedErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
				if tt.validate != nil {
					tt.validate(t, productIDs)
				}
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

// TestTagRepository_GetOnlineProductsByTag 测试获取标签关联的在线商品列表
func TestTagRepository_GetOnlineProductsByTag(t *testing.T) {
	tests := []struct {
		name        string
		tagID       int
		shopID      snowflake.ID
		setupMock   func(mock sqlmock.Sqlmock)
		expectedErr bool
		errMsg      string
		validate    func(t *testing.T, products []models.Product)
	}{
		{
			name:  "successfully get online products",
			tagID: 1,
			shopID: 123,
			setupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "shop_id", "name", "description", "price", "stock", "image_url", "status", "created_at", "updated_at"}).
					AddRow(100, int64(123), "Product 1", "Description 1", 100.0, 10, "image1.jpg", "online", time.Now(), time.Now()).
					AddRow(200, int64(123), "Product 2", "Description 2", 200.0, 20, "image2.jpg", "online", time.Now(), time.Now())
				mock.ExpectQuery("SELECT `products`\\.\\S+ FROM `products` JOIN product_tags").
					WithArgs(1, "online", int64(123)).
					WillReturnRows(rows)
			},
			expectedErr: false,
			validate: func(t *testing.T, products []models.Product) {
				assert.Len(t, products, 2)
				assert.Equal(t, snowflake.ID(100), products[0].ID)
				assert.Equal(t, "Product 1", products[0].Name)
				assert.Equal(t, "online", products[0].Status)
			},
		},
		{
			name:  "empty online products",
			tagID: 1,
			shopID: 123,
			setupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "shop_id", "name", "description", "price", "stock", "image_url", "status", "created_at", "updated_at"})
				mock.ExpectQuery("SELECT `products`\\.\\S+ FROM `products` JOIN product_tags").
					WithArgs(1, "online", int64(123)).
					WillReturnRows(rows)
			},
			expectedErr: false,
			validate: func(t *testing.T, products []models.Product) {
				assert.Empty(t, products)
			},
		},
		{
			name:  "database error",
			tagID: 1,
			shopID: 123,
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT `products`\\.\\S+ FROM `products` JOIN product_tags").
					WithArgs(1, "online", int64(123)).
					WillReturnError(fmt.Errorf("database error"))
			},
			expectedErr: true,
			errMsg:      "查询标签关联商品失败",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, sqlDB := setupTestDB(t)
			defer sqlDB.Close()

			tt.setupMock(mock)

			repo := NewTagRepository(db)
			products, err := repo.GetOnlineProductsByTag(tt.tagID, tt.shopID)

			if tt.expectedErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
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

// TestTagRepository_BatchTagProducts 测试批量打标签
func TestTagRepository_BatchTagProducts(t *testing.T) {
	tests := []struct {
		name        string
		productIDs  []snowflake.ID
		tagID       int
		shopID      snowflake.ID
		setupMock   func(mock sqlmock.Sqlmock)
		expectedErr bool
		errMsg      string
		validate    func(t *testing.T, result *BatchTagProductsResult)
	}{
		{
			name:       "successfully batch tag products",
			productIDs: []snowflake.ID{100, 200, 300},
			tagID:      1,
			shopID:     123,
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				// Query valid products
				rows := sqlmock.NewRows([]string{"id"}).
					AddRow(int64(100)).
					AddRow(int64(200))
				mock.ExpectQuery("SELECT `id` FROM `products`").
					WithArgs(int64(100), int64(200), int64(300), int64(123)).
					WillReturnRows(rows)
				// Insert product_tags
				mock.ExpectExec("INSERT INTO `product_tags`").
					WillReturnResult(sqlmock.NewResult(0, 2))
				mock.ExpectCommit()
			},
			expectedErr: false,
			validate: func(t *testing.T, result *BatchTagProductsResult) {
				assert.Equal(t, 3, result.Total)
				assert.Equal(t, 2, result.Successful)
			},
		},
		{
			name:       "all products valid",
			productIDs: []snowflake.ID{100, 200},
			tagID:      1,
			shopID:     123,
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				rows := sqlmock.NewRows([]string{"id"}).
					AddRow(int64(100)).
					AddRow(int64(200))
				mock.ExpectQuery("SELECT `id` FROM `products`").
					WithArgs(int64(100), int64(200), int64(123)).
					WillReturnRows(rows)
				mock.ExpectExec("INSERT INTO `product_tags`").
					WillReturnResult(sqlmock.NewResult(0, 2))
				mock.ExpectCommit()
			},
			expectedErr: false,
			validate: func(t *testing.T, result *BatchTagProductsResult) {
				assert.Equal(t, 2, result.Total)
				assert.Equal(t, 2, result.Successful)
			},
		},
		{
			name:       "no valid products",
			productIDs: []snowflake.ID{100, 200},
			tagID:      1,
			shopID:     123,
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				rows := sqlmock.NewRows([]string{"id"})
				mock.ExpectQuery("SELECT `id` FROM `products`").
					WithArgs(int64(100), int64(200), int64(123)).
					WillReturnRows(rows)
				// No INSERT when no valid products
				mock.ExpectCommit()
			},
			expectedErr: false,
			validate: func(t *testing.T, result *BatchTagProductsResult) {
				assert.Equal(t, 2, result.Total)
				assert.Equal(t, 0, result.Successful)
			},
		},
		{
			name:       "query products fails",
			productIDs: []snowflake.ID{100, 200},
			tagID:      1,
			shopID:     123,
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectQuery("SELECT `id` FROM `products`").
					WithArgs(int64(100), int64(200), int64(123)).
					WillReturnError(fmt.Errorf("database error"))
				mock.ExpectRollback()
			},
			expectedErr: true,
			errMsg:      "批量查询商品失败",
		},
		{
			name:       "insert fails",
			productIDs: []snowflake.ID{100},
			tagID:      1,
			shopID:     123,
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				rows := sqlmock.NewRows([]string{"id"}).AddRow(int64(100))
				mock.ExpectQuery("SELECT `id` FROM `products`").
					WithArgs(int64(100), int64(123)).
					WillReturnRows(rows)
				mock.ExpectExec("INSERT INTO `product_tags`").
					WillReturnError(fmt.Errorf("insert error"))
				mock.ExpectRollback()
			},
			expectedErr: true,
			errMsg:      "批量打标签失败",
		},
		{
			name:       "commit fails",
			productIDs: []snowflake.ID{100},
			tagID:      1,
			shopID:     123,
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				rows := sqlmock.NewRows([]string{"id"}).AddRow(int64(100))
				mock.ExpectQuery("SELECT `id` FROM `products`").
					WithArgs(int64(100), int64(123)).
					WillReturnRows(rows)
				mock.ExpectExec("INSERT INTO `product_tags`").
					WillReturnResult(sqlmock.NewResult(0, 1))
				mock.ExpectCommit().WillReturnError(fmt.Errorf("commit error"))
			},
			expectedErr: true,
			errMsg:      "批量打标签失败",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, sqlDB := setupTestDB(t)
			defer sqlDB.Close()

			tt.setupMock(mock)

			repo := NewTagRepository(db)
			result, err := repo.BatchTagProducts(tt.productIDs, tt.tagID, tt.shopID)

			if tt.expectedErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
				if tt.validate != nil {
					tt.validate(t, result)
				}
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

// TestTagRepository_BatchUntagProducts 测试批量解绑商品标签
func TestTagRepository_BatchUntagProducts(t *testing.T) {
	tests := []struct {
		name        string
		productIDs  []snowflake.ID
		tagID       uint
		shopID      snowflake.ID
		setupMock   func(mock sqlmock.Sqlmock)
		expectedErr bool
		errMsg      string
		validate    func(t *testing.T, result *BatchUntagProductsResult)
	}{
		{
			name:       "successfully batch untag products",
			productIDs: []snowflake.ID{100, 200, 300},
			tagID:      1,
			shopID:     123,
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec("DELETE FROM `product_tags`").
					WithArgs(int64(123), 1, int64(100), int64(200), int64(300)).
					WillReturnResult(sqlmock.NewResult(0, 2))
				mock.ExpectCommit()
			},
			expectedErr: false,
			validate: func(t *testing.T, result *BatchUntagProductsResult) {
				assert.Equal(t, 3, result.Total)
				assert.Equal(t, int64(2), result.Successful)
			},
		},
		{
			name:       "delete fails",
			productIDs: []snowflake.ID{100, 200},
			tagID:      1,
			shopID:     123,
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec("DELETE FROM `product_tags`").
					WillReturnError(fmt.Errorf("delete error"))
				mock.ExpectRollback()
			},
			expectedErr: true,
			errMsg:      "批量解绑标签失败",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, sqlDB := setupTestDB(t)
			defer sqlDB.Close()

			tt.setupMock(mock)

			repo := NewTagRepository(db)
			result, err := repo.BatchUntagProducts(tt.productIDs, tt.tagID, tt.shopID)

			if tt.expectedErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
				if tt.validate != nil {
					tt.validate(t, result)
				}
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

// TestTagRepository_GetUnboundTagsList 测试获取未绑定任何商品的标签列表
func TestTagRepository_GetUnboundTagsList(t *testing.T) {
	tests := []struct {
		name        string
		shopID      snowflake.ID
		page        int
		pageSize    int
		setupMock   func(mock sqlmock.Sqlmock)
		expectedErr bool
		errMsg      string
		validate    func(t *testing.T, tags []models.Tag, total int64)
	}{
		{
			name:     "successfully get unbound tags",
			shopID:   123,
			page:     1,
			pageSize: 10,
			setupMock: func(mock sqlmock.Sqlmock) {
				// Main query
				tagRows := sqlmock.NewRows([]string{"id", "name", "shop_id", "description", "created_at", "updated_at"}).
					AddRow(1, "Tag 1", int64(123), "Description 1", time.Now(), time.Now()).
					AddRow(2, "Tag 2", int64(123), "Description 2", time.Now(), time.Now())
				mock.ExpectQuery("SELECT \\* FROM tags").
					WithArgs(int64(123), 10, 0).
					WillReturnRows(tagRows)
				// Count query
				countRows := sqlmock.NewRows([]string{"count"}).AddRow(2)
				mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM tags").
					WithArgs(int64(123)).
					WillReturnRows(countRows)
			},
			expectedErr: false,
			validate: func(t *testing.T, tags []models.Tag, total int64) {
				assert.Len(t, tags, 2)
				assert.Equal(t, int64(2), total)
			},
		},
		{
			name:     "empty unbound tags",
			shopID:   123,
			page:     1,
			pageSize: 10,
			setupMock: func(mock sqlmock.Sqlmock) {
				tagRows := sqlmock.NewRows([]string{"id", "name", "shop_id", "description", "created_at", "updated_at"})
				mock.ExpectQuery("SELECT \\* FROM tags").
					WithArgs(int64(123), 10, 0).
					WillReturnRows(tagRows)
				countRows := sqlmock.NewRows([]string{"count"}).AddRow(0)
				mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM tags").
					WithArgs(int64(123)).
					WillReturnRows(countRows)
			},
			expectedErr: false,
			validate: func(t *testing.T, tags []models.Tag, total int64) {
				assert.Empty(t, tags)
				assert.Equal(t, int64(0), total)
			},
		},
		{
			name:     "database error",
			shopID:   123,
			page:     1,
			pageSize: 10,
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT \\* FROM tags").
					WithArgs(int64(123), 10, 0).
					WillReturnError(fmt.Errorf("database error"))
			},
			expectedErr: true,
			errMsg:      "查询未绑定标签失败",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, sqlDB := setupTestDB(t)
			defer sqlDB.Close()

			tt.setupMock(mock)

			repo := NewTagRepository(db)
			tags, total, err := repo.GetUnboundTagsList(tt.shopID, tt.page, tt.pageSize)

			if tt.expectedErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
				if tt.validate != nil {
					tt.validate(t, tags, total)
				}
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

// TestTagRepository_GetUnboundProductsForTag 测试获取可绑定到指定标签的商品列表
func TestTagRepository_GetUnboundProductsForTag(t *testing.T) {
	tests := []struct {
		name        string
		tagID       int
		shopID      snowflake.ID
		page        int
		pageSize    int
		setupMock   func(mock sqlmock.Sqlmock)
		expectedErr bool
		errMsg      string
		validate    func(t *testing.T, products []models.Product, total int64)
	}{
		{
			name:     "successfully get unbound products",
			tagID:    1,
			shopID:   123,
			page:     1,
			pageSize: 10,
			setupMock: func(mock sqlmock.Sqlmock) {
				// Main query
				productRows := sqlmock.NewRows([]string{"id", "shop_id", "name", "description", "price", "stock", "image_url", "status", "created_at", "updated_at"}).
					AddRow(100, int64(123), "Product 1", "Description 1", 100.0, 10, "image1.jpg", "online", time.Now(), time.Now()).
					AddRow(200, int64(123), "Product 2", "Description 2", 200.0, 20, "image2.jpg", "online", time.Now(), time.Now())
				mock.ExpectQuery("SELECT \\* FROM products").
					WithArgs(1, int64(123), 10, 0).
					WillReturnRows(productRows)
				// Count query
				countRows := sqlmock.NewRows([]string{"count"}).AddRow(2)
				mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM products").
					WithArgs(1, int64(123)).
					WillReturnRows(countRows)
			},
			expectedErr: false,
			validate: func(t *testing.T, products []models.Product, total int64) {
				assert.Len(t, products, 2)
				assert.Equal(t, int64(2), total)
			},
		},
		{
			name:     "empty unbound products",
			tagID:    1,
			shopID:   123,
			page:     1,
			pageSize: 10,
			setupMock: func(mock sqlmock.Sqlmock) {
				productRows := sqlmock.NewRows([]string{"id", "shop_id", "name", "description", "price", "stock", "image_url", "status", "created_at", "updated_at"})
				mock.ExpectQuery("SELECT \\* FROM products").
					WithArgs(1, int64(123), 10, 0).
					WillReturnRows(productRows)
				countRows := sqlmock.NewRows([]string{"count"}).AddRow(0)
				mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM products").
					WithArgs(1, int64(123)).
					WillReturnRows(countRows)
			},
			expectedErr: false,
			validate: func(t *testing.T, products []models.Product, total int64) {
				assert.Empty(t, products)
				assert.Equal(t, int64(0), total)
			},
		},
		{
			name:     "database error",
			tagID:    1,
			shopID:   123,
			page:     1,
			pageSize: 10,
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT \\* FROM products").
					WithArgs(1, int64(123), 10, 0).
					WillReturnError(fmt.Errorf("database error"))
			},
			expectedErr: true,
			errMsg:      "查询未绑定商品失败",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, sqlDB := setupTestDB(t)
			defer sqlDB.Close()

			tt.setupMock(mock)

			repo := NewTagRepository(db)
			products, total, err := repo.GetUnboundProductsForTag(tt.tagID, tt.shopID, tt.page, tt.pageSize)

			if tt.expectedErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
				if tt.validate != nil {
					tt.validate(t, products, total)
				}
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

// TestTagRepository_GetBoundProductsWithPagination 测试获取标签绑定的商品（分页）
func TestTagRepository_GetBoundProductsWithPagination(t *testing.T) {
	tests := []struct {
		name        string
		tagID       int
		shopID      snowflake.ID
		page        int
		pageSize    int
		onlyOnline  bool
		setupMock   func(mock sqlmock.Sqlmock)
		expectedErr bool
		errMsg      string
		validate    func(t *testing.T, result *BoundProductsResult)
	}{
		{
			name:       "successfully get bound products - all",
			tagID:      1,
			shopID:     123,
			page:       1,
			pageSize:   10,
			onlyOnline: false,
			setupMock: func(mock sqlmock.Sqlmock) {
				// Get bound product IDs
				idRows := sqlmock.NewRows([]string{"product_id"}).
					AddRow(100).
					AddRow(200)
				mock.ExpectQuery("SELECT product_id FROM product_tags").
					WithArgs(1, int64(123)).
					WillReturnRows(idRows)
				// Count query
				countRows := sqlmock.NewRows([]string{"count"}).AddRow(2)
				mock.ExpectQuery("SELECT count\\(\\*\\) FROM `products`").
					WithArgs(100, 200).
					WillReturnRows(countRows)
				// Find products
				productRows := sqlmock.NewRows([]string{"id", "shop_id", "name", "description", "price", "stock", "image_url", "status", "created_at", "updated_at"}).
					AddRow(100, int64(123), "Product 1", "Description 1", 100.0, 10, "image1.jpg", "online", time.Now(), time.Now())
				mock.ExpectQuery("SELECT \\* FROM `products`").
					WithArgs(100, 200, 10).
					WillReturnRows(productRows)
				// Preload OptionCategories for product 100
				categoryRows := sqlmock.NewRows([]string{"id", "name", "product_id"})
				mock.ExpectQuery("SELECT \\* FROM `product_option_categories`").
					WithArgs(int64(100)).
					WillReturnRows(categoryRows)
			},
			expectedErr: false,
			validate: func(t *testing.T, result *BoundProductsResult) {
				assert.NotNil(t, result)
				assert.Len(t, result.Products, 1)
				assert.Equal(t, int64(2), result.Total)
			},
		},
		{
			name:       "successfully get bound products - online only",
			tagID:      1,
			shopID:     123,
			page:       1,
			pageSize:   10,
			onlyOnline: true,
			setupMock: func(mock sqlmock.Sqlmock) {
				// Get bound product IDs
				idRows := sqlmock.NewRows([]string{"product_id"}).
					AddRow(100).
					AddRow(200)
				mock.ExpectQuery("SELECT product_id FROM product_tags").
					WithArgs(1, int64(123)).
					WillReturnRows(idRows)
				// Count query with status filter
				countRows := sqlmock.NewRows([]string{"count"}).AddRow(1)
				mock.ExpectQuery("SELECT count\\(\\*\\) FROM `products`").
					WithArgs(100, 200, "online").
					WillReturnRows(countRows)
				// Find products with status filter
				productRows := sqlmock.NewRows([]string{"id", "shop_id", "name", "description", "price", "stock", "image_url", "status", "created_at", "updated_at"}).
					AddRow(100, int64(123), "Product 1", "Description 1", 100.0, 10, "image1.jpg", "online", time.Now(), time.Now())
				mock.ExpectQuery("SELECT \\* FROM `products`").
					WithArgs(100, 200, "online", 10).
					WillReturnRows(productRows)
				// Preload OptionCategories for product 100
				categoryRows := sqlmock.NewRows([]string{"id", "name", "product_id"})
				mock.ExpectQuery("SELECT \\* FROM `product_option_categories`").
					WithArgs(int64(100)).
					WillReturnRows(categoryRows)
			},
			expectedErr: false,
			validate: func(t *testing.T, result *BoundProductsResult) {
				assert.NotNil(t, result)
				assert.Len(t, result.Products, 1)
				assert.Equal(t, int64(1), result.Total)
			},
		},
		{
			name:       "get product IDs fails",
			tagID:      1,
			shopID:     123,
			page:       1,
			pageSize:   10,
			onlyOnline: false,
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT product_id FROM product_tags").
					WithArgs(1, int64(123)).
					WillReturnError(fmt.Errorf("database error"))
			},
			expectedErr: true,
			errMsg:      "获取绑定商品ID列表失败",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, sqlDB := setupTestDB(t)
			defer sqlDB.Close()

			tt.setupMock(mock)

			repo := NewTagRepository(db)
			result, err := repo.GetBoundProductsWithPagination(tt.tagID, tt.shopID, tt.page, tt.pageSize, tt.onlyOnline)

			if tt.expectedErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
				if tt.validate != nil {
					tt.validate(t, result)
				}
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

// TestTagRepository_GetUnboundProductsWithPagination 测试获取未绑定任何标签的商品（分页）
func TestTagRepository_GetUnboundProductsWithPagination(t *testing.T) {
	tests := []struct {
		name        string
		shopID      snowflake.ID
		page        int
		pageSize    int
		onlyOnline  bool
		setupMock   func(mock sqlmock.Sqlmock)
		expectedErr bool
		errMsg      string
		validate    func(t *testing.T, result *BoundProductsResult)
	}{
		{
			name:       "successfully get unbound products - all",
			shopID:     123,
			page:       1,
			pageSize:   10,
			onlyOnline: false,
			setupMock: func(mock sqlmock.Sqlmock) {
				// Count query
				countRows := sqlmock.NewRows([]string{"count"}).AddRow(5)
				mock.ExpectQuery("SELECT count\\(\\*\\) FROM `products`").
					WithArgs(int64(123)).
					WillReturnRows(countRows)
				// Find products
				productRows := sqlmock.NewRows([]string{"id", "shop_id", "name", "description", "price", "stock", "image_url", "status", "created_at", "updated_at"}).
					AddRow(100, int64(123), "Product 1", "Description 1", 100.0, 10, "image1.jpg", "online", time.Now(), time.Now())
				mock.ExpectQuery("SELECT \\* FROM `products`").
					WithArgs(int64(123), 10).
					WillReturnRows(productRows)
				// Preload OptionCategories for product 100
				categoryRows := sqlmock.NewRows([]string{"id", "name", "product_id"})
				mock.ExpectQuery("SELECT \\* FROM `product_option_categories`").
					WithArgs(int64(100)).
					WillReturnRows(categoryRows)
			},
			expectedErr: false,
			validate: func(t *testing.T, result *BoundProductsResult) {
				assert.NotNil(t, result)
				assert.Len(t, result.Products, 1)
				assert.Equal(t, int64(5), result.Total)
			},
		},
		{
			name:       "successfully get unbound products - online only",
			shopID:     123,
			page:       1,
			pageSize:   10,
			onlyOnline: true,
			setupMock: func(mock sqlmock.Sqlmock) {
				// Count query with status filter
				countRows := sqlmock.NewRows([]string{"count"}).AddRow(3)
				mock.ExpectQuery("SELECT count\\(\\*\\) FROM `products`").
					WithArgs(int64(123), "online").
					WillReturnRows(countRows)
				// Find products with status filter
				productRows := sqlmock.NewRows([]string{"id", "shop_id", "name", "description", "price", "stock", "image_url", "status", "created_at", "updated_at"}).
					AddRow(100, int64(123), "Product 1", "Description 1", 100.0, 10, "image1.jpg", "online", time.Now(), time.Now())
				mock.ExpectQuery("SELECT \\* FROM `products`").
					WithArgs(int64(123), "online", 10).
					WillReturnRows(productRows)
				// Preload OptionCategories for product 100
				categoryRows := sqlmock.NewRows([]string{"id", "name", "product_id"})
				mock.ExpectQuery("SELECT \\* FROM `product_option_categories`").
					WithArgs(int64(100)).
					WillReturnRows(categoryRows)
			},
			expectedErr: false,
			validate: func(t *testing.T, result *BoundProductsResult) {
				assert.NotNil(t, result)
				assert.Len(t, result.Products, 1)
				assert.Equal(t, int64(3), result.Total)
			},
		},
		{
			name:       "count query fails",
			shopID:     123,
			page:       1,
			pageSize:   10,
			onlyOnline: false,
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT count\\(\\*\\) FROM `products`").
					WithArgs(int64(123)).
					WillReturnError(fmt.Errorf("database error"))
			},
			expectedErr: true,
			errMsg:      "获取未绑定商品总数失败",
		},
		{
			name:       "find query fails",
			shopID:     123,
			page:       1,
			pageSize:   10,
			onlyOnline: false,
			setupMock: func(mock sqlmock.Sqlmock) {
				countRows := sqlmock.NewRows([]string{"count"}).AddRow(5)
				mock.ExpectQuery("SELECT count\\(\\*\\) FROM `products`").
					WithArgs(int64(123)).
					WillReturnRows(countRows)
				mock.ExpectQuery("SELECT \\* FROM `products`").
					WithArgs(int64(123), 10).
					WillReturnError(fmt.Errorf("database error"))
			},
			expectedErr: true,
			errMsg:      "查询未绑定商品失败",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, sqlDB := setupTestDB(t)
			defer sqlDB.Close()

			tt.setupMock(mock)

			repo := NewTagRepository(db)
			result, err := repo.GetUnboundProductsWithPagination(tt.shopID, tt.page, tt.pageSize, tt.onlyOnline)

			if tt.expectedErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
				if tt.validate != nil {
					tt.validate(t, result)
				}
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

// TestTagRepository_ConcurrentAccess 测试并发访问
func TestTagRepository_ConcurrentAccess(t *testing.T) {
	db, mock, sqlDB := setupTestDB(t)
	defer sqlDB.Close()

	// Set up multiple mock expectations
	for i := 0; i < 5; i++ {
		rows := sqlmock.NewRows([]string{"id", "name", "shop_id", "description", "created_at", "updated_at"}).
			AddRow(i+1, fmt.Sprintf("Tag %d", i+1), int64(123), fmt.Sprintf("Description %d", i+1), time.Now(), time.Now())
		mock.ExpectQuery("SELECT \\* FROM `tags` WHERE shop_id = \\? ORDER BY created_at DESC").
			WithArgs(int64(123)).
			WillReturnRows(rows)
	}

	repo := NewTagRepository(db)

	var wg sync.WaitGroup
	errors := make(chan error, 5)

	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := repo.GetListByShopID(123)
			if err != nil {
				errors <- err
			}
		}()
	}

	wg.Wait()
	close(errors)

	for err := range errors {
		t.Errorf("Concurrent access error: %v", err)
	}

	assert.NoError(t, mock.ExpectationsWereMet())
}
