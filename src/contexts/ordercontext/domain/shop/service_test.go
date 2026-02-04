package shop

import (
	"database/sql"
	"fmt"
	"testing"
	"time"

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

func TestShopService_DeleteShop(t *testing.T) {
	tests := []struct {
		name          string
		shopID        snowflake.ID
		productCount  int64
		orderCount    int64
		setupShop     bool
		shopExists    bool
		expectedErr   bool
		expectedMsg   string
	}{
		{
			name:         "successfully delete shop - no products or orders",
			shopID:       123,
			productCount: 0,
			orderCount:   0,
			setupShop:    true,
			shopExists:   true,
			expectedErr:  false,
		},
		{
			name:         "cannot delete - has products",
			shopID:       456,
			productCount: 5,
			orderCount:   0,
			setupShop:    true,
			shopExists:   true,
			expectedErr:  true,
			expectedMsg: "店铺存在 5 个关联商品",
		},
		{
			name:         "cannot delete - has orders",
			shopID:       789,
			productCount: 0,
			orderCount:   3,
			setupShop:    true,
			shopExists:   true,
			expectedErr:  true,
			expectedMsg: "店铺存在 3 个关联订单",
		},
		{
			name:         "cannot delete - has both products and orders",
			shopID:       999,
			productCount: 10,
			orderCount:   5,
			setupShop:    true,
			shopExists:   true,
			expectedErr:  true,
			expectedMsg: "店铺存在 10 个关联商品",
		},
		{
			name:         "shop not found",
			shopID:       111,
			productCount: 0,
			orderCount:   0,
			setupShop:    false,
			shopExists:   false,
			expectedErr:  true,
			expectedMsg: "店铺不存在",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, sqlDB := setupTestDB(t)
			defer sqlDB.Close()

			service := NewService(db)

			if tt.shopExists {
				// Mock shop query - GORM adds LIMIT 1 when using First
				shopRows := sqlmock.NewRows([]string{"id", "name", "owner_username", "owner_password",
					"contact_phone", "contact_email", "address", "image_url", "description",
					"created_at", "updated_at", "valid_until", "settings", "order_status_flow"}).
					AddRow(tt.shopID, "Test Shop", "owner", "hashed_pass", "13800138000",
						"test@example.com", "test address", "http://example.com/image.jpg", "test description",
						time.Now(), time.Now(), time.Now().AddDate(1, 0, 0), []byte("{}"), models.OrderStatusFlow{})

				mock.ExpectQuery("SELECT \\* FROM `shops` WHERE").
					WithArgs(tt.shopID, 1).
					WillReturnRows(shopRows)
			} else {
				// Mock shop not found
				mock.ExpectQuery("SELECT \\* FROM `shops` WHERE").
					WithArgs(tt.shopID, 1).
					WillReturnError(gorm.ErrRecordNotFound)
			}

			if tt.shopExists {
				// Mock product count query
				productRows := sqlmock.NewRows([]string{"count"}).AddRow(tt.productCount)
				mock.ExpectQuery("SELECT count\\(\\*\\) FROM `products`").
					WithArgs(tt.shopID).
					WillReturnRows(productRows)

				// Mock order count query
				orderRows := sqlmock.NewRows([]string{"count"}).AddRow(tt.orderCount)
				mock.ExpectQuery("SELECT count\\(\\*\\) FROM `orders`").
					WithArgs(tt.shopID).
					WillReturnRows(orderRows)
			}

			if tt.shopExists && tt.productCount == 0 && tt.orderCount == 0 {
				// Mock delete
				mock.ExpectBegin()
				mock.ExpectExec("DELETE FROM `shops`").
					WithArgs(tt.shopID).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			}

			// Execute the method
			err := service.DeleteShop(tt.shopID)

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

func TestShopService_ProcessValidUntil(t *testing.T) {
	service := &Service{}

	tests := []struct {
		name          string
		validUntilStr string
		wantErr       bool
		errMsg        string
		validate      func(t *testing.T, result time.Time)
	}{
		{
			name:          "empty string - use default 1 year",
			validUntilStr: "",
			wantErr:       false,
			validate: func(t *testing.T, result time.Time) {
				// Should be approximately 1 year from now
				expected := time.Now().AddDate(1, 0, 0)
				diff := result.Sub(expected)
				assert.Less(t, diff.Abs(), time.Minute)
			},
		},
		{
			name:          "valid RFC3339 format",
			validUntilStr: "2025-12-31T23:59:59Z",
			wantErr:       false,
			validate: func(t *testing.T, result time.Time) {
				expected, _ := time.Parse(time.RFC3339, "2025-12-31T23:59:59Z")
				assert.Equal(t, expected, result)
			},
		},
		{
			name:          "valid RFC3339 with timezone",
			validUntilStr: "2025-12-31T23:59:59+08:00",
			wantErr:       false,
			validate: func(t *testing.T, result time.Time) {
				expected, _ := time.Parse(time.RFC3339, "2025-12-31T23:59:59+08:00")
				assert.Equal(t, expected, result)
			},
		},
		{
			name:          "invalid format - error",
			validUntilStr: "2024-01-01",
			wantErr:       true,
			errMsg:        "无效的有效期格式",
		},
		{
			name:          "invalid date - error",
			validUntilStr: "invalid-date",
			wantErr:       true,
			errMsg:        "无效的有效期格式",
		},
		{
			name:          "invalid format - Chinese characters",
			validUntilStr: "2024年1月1日",
			wantErr:       true,
			errMsg:        "无效的有效期格式",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := service.ProcessValidUntil(tt.validUntilStr)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
				tt.validate(t, result)
			}
		})
	}
}

func TestShopService_ProcessValidUntil_EdgeCases(t *testing.T) {
	service := &Service{}

	t.Run("past date - still parses correctly", func(t *testing.T) {
		pastDate := "2020-01-01T00:00:00Z"
		result, err := service.ProcessValidUntil(pastDate)

		assert.NoError(t, err)
		expected, _ := time.Parse(time.RFC3339, pastDate)
		assert.Equal(t, expected, result)
	})

	t.Run("far future date", func(t *testing.T) {
		futureDate := "2030-12-31T23:59:59Z"
		result, err := service.ProcessValidUntil(futureDate)

		assert.NoError(t, err)
		expected, _ := time.Parse(time.RFC3339, futureDate)
		assert.Equal(t, expected, result)
	})

	t.Run("leap year date", func(t *testing.T) {
		leapDate := "2024-02-29T00:00:00Z"
		result, err := service.ProcessValidUntil(leapDate)

		assert.NoError(t, err)
		expected, _ := time.Parse(time.RFC3339, leapDate)
		assert.Equal(t, expected, result)
	})
}

func TestShopService_ParseOrderStatusFlow(t *testing.T) {
	service := &Service{}

	tests := []struct {
		name    string
		input   *models.OrderStatusFlow
		wantErr bool
		validate func(t *testing.T, result models.OrderStatusFlow)
	}{
		{
			name:  "nil input - use default",
			input: nil,
			wantErr: false,
			validate: func(t *testing.T, result models.OrderStatusFlow) {
				// Default flow should have statuses
				assert.NotEmpty(t, result.Statuses)
			},
		},
		{
			name: "custom flow provided",
			input: &models.OrderStatusFlow{
				Statuses: []models.OrderStatus{
					{Value: 1, Label: "自定义状态1", Type: "warning", IsFinal: false, Actions: []models.OrderStatusAction{}},
					{Value: 2, Label: "自定义状态2", Type: "primary", IsFinal: false, Actions: []models.OrderStatusAction{}},
				},
			},
			wantErr: false,
			validate: func(t *testing.T, result models.OrderStatusFlow) {
				assert.Len(t, result.Statuses, 2)
				assert.Equal(t, "自定义状态1", result.Statuses[0].Label)
			},
		},
		{
			name:  "empty flow provided - keeps empty",
			input: &models.OrderStatusFlow{},
			wantErr: false,
			validate: func(t *testing.T, result models.OrderStatusFlow) {
				// Empty input (even if non-nil) replaces the default with empty
				assert.Empty(t, result.Statuses)
			},
		},
		{
			name: "complex flow with multiple statuses",
			input: &models.OrderStatusFlow{
				Statuses: []models.OrderStatus{
					{Value: 1, Label: "待处理", Type: "warning", IsFinal: false, Actions: []models.OrderStatusAction{}},
					{Value: 2, Label: "已接单", Type: "primary", IsFinal: false, Actions: []models.OrderStatusAction{}},
					{Value: 4, Label: "已发货", Type: "info", IsFinal: false, Actions: []models.OrderStatusAction{}},
					{Value: 10, Label: "已完成", Type: "success", IsFinal: true, Actions: []models.OrderStatusAction{}},
				},
			},
			wantErr: false,
			validate: func(t *testing.T, result models.OrderStatusFlow) {
				assert.Len(t, result.Statuses, 4)
				assert.Equal(t, "已完成", result.Statuses[3].Label)
				assert.True(t, result.Statuses[3].IsFinal)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := service.ParseOrderStatusFlow(tt.input)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				tt.validate(t, result)
			}
		})
	}
}

func TestShopService_DeleteShop_TransactionHandling(t *testing.T) {
	tests := []struct {
		name         string
		shopID       snowflake.ID
		deleteError  bool
		commitError  bool
		expectedErr  bool
		expectedMsg  string
	}{
		{
			name:        "delete succeeds - transaction commits",
			shopID:      123,
			deleteError: false,
			commitError: false,
			expectedErr: false,
		},
		{
			name:        "delete fails - transaction rolls back",
			shopID:      456,
			deleteError: true,
			commitError: false,
			expectedErr: true,
			expectedMsg: "删除店铺失败",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, sqlDB := setupTestDB(t)
			defer sqlDB.Close()

			service := NewService(db)

			// Mock shop query
			shopRows := sqlmock.NewRows([]string{"id", "name", "owner_username", "owner_password",
				"contact_phone", "contact_email", "address", "image_url", "description",
				"created_at", "updated_at", "valid_until", "settings", "order_status_flow"}).
				AddRow(tt.shopID, "Test Shop", "owner", "hashed_pass", "13800138000",
					"test@example.com", "test address", "http://example.com/image.jpg", "test description",
					time.Now(), time.Now(), time.Now().AddDate(1, 0, 0), []byte("{}"), models.OrderStatusFlow{})

			mock.ExpectQuery("SELECT \\* FROM `shops` WHERE").
				WithArgs(tt.shopID, 1).
				WillReturnRows(shopRows)

			// Mock product count query (0 products)
			productRows := sqlmock.NewRows([]string{"count"}).AddRow(0)
			mock.ExpectQuery("SELECT count\\(\\*\\) FROM `products`").
				WithArgs(tt.shopID).
				WillReturnRows(productRows)

			// Mock order count query (0 orders)
			orderRows := sqlmock.NewRows([]string{"count"}).AddRow(0)
			mock.ExpectQuery("SELECT count\\(\\*\\) FROM `orders`").
				WithArgs(tt.shopID).
				WillReturnRows(orderRows)

			// Mock transaction
			mock.ExpectBegin()

			if tt.deleteError {
				mock.ExpectExec("DELETE FROM `shops`").
					WithArgs(tt.shopID).
					WillReturnError(fmt.Errorf("database error"))
				mock.ExpectRollback()
			} else {
				mock.ExpectExec("DELETE FROM `shops`").
					WithArgs(tt.shopID).
					WillReturnResult(sqlmock.NewResult(1, 1))

				if tt.commitError {
					mock.ExpectCommit().WillReturnError(fmt.Errorf("commit error"))
				} else {
					mock.ExpectCommit()
				}
			}

			// Execute the method
			err := service.DeleteShop(tt.shopID)

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
