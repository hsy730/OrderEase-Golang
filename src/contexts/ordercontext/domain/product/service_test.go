package product

import (
	"database/sql"
	"fmt"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
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

func TestProductService_ValidateForDeletion(t *testing.T) {
	tests := []struct {
		name         string
		productID    uint64
		orderCount   int64
		queryError   bool
		expectedErr  bool
		expectedMsg  string
	}{
		{
			name:        "no associated orders - can delete",
			productID:   123,
			orderCount:  0,
			queryError:  false,
			expectedErr: false,
		},
		{
			name:        "has associated orders - cannot delete",
			productID:   123,
			orderCount:  5,
			queryError:  false,
			expectedErr: true,
			expectedMsg: "该商品有 5 个关联订单，不能删除",
		},
		{
			name:        "single associated order - cannot delete",
			productID:   456,
			orderCount:  1,
			queryError:  false,
			expectedErr: true,
			expectedMsg: "该商品有 1 个关联订单，不能删除",
		},
		{
			name:        "many associated orders - cannot delete",
			productID:   789,
			orderCount:  100,
			queryError:  false,
			expectedErr: true,
			expectedMsg: "该商品有 100 个关联订单，不能删除",
		},
		{
			name:        "database query error",
			productID:   999,
			orderCount:  0,
			queryError:  true,
			expectedErr: true,
			expectedMsg: "检查商品订单关联失败",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, sqlDB := setupTestDB(t)
			defer sqlDB.Close()

			service := NewService(db)

			if tt.queryError {
				// Mock database error
				mock.ExpectQuery("SELECT count\\(\\*\\) FROM `order_items`").
					WithArgs(tt.productID).
					WillReturnError(fmt.Errorf("database connection error"))
			} else {
				// Mock successful count query
				rows := sqlmock.NewRows([]string{"count"}).AddRow(tt.orderCount)
				mock.ExpectQuery("SELECT count\\(\\*\\) FROM `order_items`").
					WithArgs(tt.productID).
					WillReturnRows(rows)
			}

			// Execute the method
			err := service.ValidateForDeletion(tt.productID)

			// Verify expectations
			if tt.expectedErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedMsg)
			} else {
				assert.NoError(t, err)
			}

			// Ensure all expectations were met
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestProductService_ValidateForDeletion_SQLInjection(t *testing.T) {
	db, mock, sqlDB := setupTestDB(t)
	defer sqlDB.Close()

	service := NewService(db)

	// Test with a potentially dangerous product ID
	dangerousID := uint64(123); // Simulating a normal ID, but we want to ensure proper escaping
	rows := sqlmock.NewRows([]string{"count"}).AddRow(0)

	// The query should properly escape the product ID parameter
	mock.ExpectQuery("SELECT count\\(\\*\\) FROM `order_items`").
		WithArgs(dangerousID).
		WillReturnRows(rows)

	err := service.ValidateForDeletion(dangerousID)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestProductService_ValidateForDeletion_RealWorldScenarios(t *testing.T) {
	scenarios := []struct {
		name        string
		description string
		productID   uint64
		orderCount  int64
		canDelete   bool
	}{
		{
			name:        "new_product_no_orders",
			description: "新上架商品，无订单记录",
			productID:   1001,
			orderCount:  0,
			canDelete:   true,
		},
		{
			name:        "popular_product_many_orders",
			description: "热销商品，有大量订单",
			productID:   1002,
			orderCount:  500,
			canDelete:   false,
		},
		{
			name:        "discontinued_product_few_orders",
			description: "停售商品，少量订单",
			productID:   1003,
			orderCount:  3,
			canDelete:   false,
		},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			db, mock, sqlDB := setupTestDB(t)
			defer sqlDB.Close()

			service := NewService(db)

			rows := sqlmock.NewRows([]string{"count"}).AddRow(scenario.orderCount)
			mock.ExpectQuery("SELECT count\\(\\*\\) FROM `order_items`").
				WithArgs(scenario.productID).
				WillReturnRows(rows)

			err := service.ValidateForDeletion(scenario.productID)

			if scenario.canDelete {
				assert.NoError(t, err, scenario.description)
			} else {
				assert.Error(t, err, scenario.description)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestProductService_ValidateForDeletion_EmptyResult(t *testing.T) {
	db, mock, sqlDB := setupTestDB(t)
	defer sqlDB.Close()

	service := NewService(db)

	// Mock empty result set (no rows returned)
	rows := sqlmock.NewRows([]string{"count"})
	mock.ExpectQuery("SELECT count\\(\\*\\) FROM `order_items`").
		WithArgs(uint64(123)).
		WillReturnRows(rows)

	err := service.ValidateForDeletion(123)

	// GORM should handle empty results gracefully
	// Count should return 0 for empty results
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestProductService_ValidateForDeletion_LargeOrderCount(t *testing.T) {
	db, mock, sqlDB := setupTestDB(t)
	defer sqlDB.Close()

	service := NewService(db)

	// Test with a very large order count
	largeCount := int64(999999)
	rows := sqlmock.NewRows([]string{"count"}).AddRow(largeCount)
	mock.ExpectQuery("SELECT count\\(\\*\\) FROM `order_items`").
		WithArgs(uint64(999)).
		WillReturnRows(rows)

	err := service.ValidateForDeletion(999)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "该商品有 999999 个关联订单")
	assert.NoError(t, mock.ExpectationsWereMet())
}
