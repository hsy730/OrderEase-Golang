package repositories

import (
	"fmt"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/bwmarrin/snowflake"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
	"orderease/models"
)

// ==================== GetOrderByIDAndShopID Tests ====================

func TestOrderRepository_GetOrderByIDAndShopID(t *testing.T) {
	tests := []struct {
		name        string
		orderID     uint64
		shopID      snowflake.ID
		setupMock   func(mock sqlmock.Sqlmock)
		expectedErr bool
		errMsg      string
		validate    func(t *testing.T, order *models.Order)
	}{
		{
			name:    "successfully get order",
			orderID: 123,
			shopID:  456,
			setupMock: func(mock sqlmock.Sqlmock) {
				// Mock order query with User join - GORM generates LEFT JOIN
				orderRows := sqlmock.NewRows([]string{
					"id", "user_id", "shop_id", "total_price", "status", "remark", "created_at", "updated_at",
					"User__id", "User__name", "User__role", "User__password", "User__phone",
				}).
					AddRow(123, 789, 456, 10000, 1, "test remark", time.Now(), time.Now(),
						789, "testuser", "private_user", "hashedpass", "13800138000")
				mock.ExpectQuery("SELECT .* FROM `orders` LEFT JOIN `users`").
					WithArgs(int64(456), int64(123), 1).
					WillReturnRows(orderRows)

				// Mock preload Items
				itemRows := sqlmock.NewRows([]string{"id", "order_id", "product_id", "quantity", "price", "total_price"}).
					AddRow(1, 123, 999, 2, 5000, 10000)
				mock.ExpectQuery("SELECT \\* FROM `order_items`").
					WithArgs(int64(123)).
					WillReturnRows(itemRows)

				// Mock preload Options for item 1
				optionRows := sqlmock.NewRows([]string{"id", "order_item_id", "category_id", "option_id"}).
					AddRow(1, 1, 10, 100)
				mock.ExpectQuery("SELECT \\* FROM `order_item_options`").
					WithArgs(int64(1)).
					WillReturnRows(optionRows)
			},
			expectedErr: false,
			validate: func(t *testing.T, order *models.Order) {
				assert.Equal(t, snowflake.ID(123), order.ID)
				assert.Equal(t, snowflake.ID(456), order.ShopID)
				assert.Equal(t, 1, order.Status)
			},
		},
		{
			name:    "order not found",
			orderID: 999,
			shopID:  456,
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT .* FROM `orders` LEFT JOIN `users`").
					WithArgs(int64(456), int64(999), 1).
					WillReturnError(gorm.ErrRecordNotFound)
			},
			expectedErr: true,
			errMsg:      "订单不存在",
		},
		{
			name:    "database error",
			orderID: 123,
			shopID:  456,
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT .* FROM `orders` LEFT JOIN `users`").
					WithArgs(int64(456), int64(123), 1).
					WillReturnError(fmt.Errorf("database error"))
			},
			expectedErr: true,
			errMsg:      "查询订单失败",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, sqlDB := setupTestDB(t)
			defer sqlDB.Close()

			repo := NewOrderRepository(db)
			tt.setupMock(mock)

			order, err := repo.GetOrderByIDAndShopID(tt.orderID, tt.shopID)

			if tt.expectedErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, order)
				if tt.validate != nil {
					tt.validate(t, order)
				}
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

// ==================== GetOrdersByShop Tests ====================

func TestOrderRepository_GetOrdersByShop(t *testing.T) {
	tests := []struct {
		name        string
		shopID      snowflake.ID
		page        int
		pageSize    int
		setupMock   func(mock sqlmock.Sqlmock)
		expectedErr bool
		validate    func(t *testing.T, orders []models.Order, total int64)
	}{
		{
			name:     "successfully get shop orders",
			shopID:   123,
			page:     1,
			pageSize: 10,
			setupMock: func(mock sqlmock.Sqlmock) {
				// Mock count query
				countRows := sqlmock.NewRows([]string{"count"}).AddRow(25)
				mock.ExpectQuery("SELECT count\\(\\*\\) FROM `orders`").
					WithArgs(int64(123)).
					WillReturnRows(countRows)

				// Mock find query - page 1 has no offset
				orderRows := sqlmock.NewRows([]string{"id", "user_id", "shop_id", "total_price", "status"}).
					AddRow(1, 100, 123, 10000, 1).
					AddRow(2, 101, 123, 20000, 2)
				mock.ExpectQuery("SELECT \\* FROM `orders`").
					WithArgs(int64(123), 10).
					WillReturnRows(orderRows)

				// Mock preload Items for orders 1 and 2 (comes before User)
				itemRows := sqlmock.NewRows([]string{"id", "order_id", "product_id", "quantity"}).
					AddRow(1, 1, 999, 2).
					AddRow(2, 2, 998, 1)
				mock.ExpectQuery("SELECT \\* FROM `order_items`").
					WithArgs(int64(1), int64(2)).
					WillReturnRows(itemRows)

				// Mock preload User - uses user_id foreign key (100, 101), not order id
				userRows := sqlmock.NewRows([]string{"id", "name", "phone"}).
					AddRow(100, "user1", "13800138000").
					AddRow(101, "user2", "13800138001")
				mock.ExpectQuery("SELECT \\* FROM `users`").
					WithArgs(int64(100), int64(101)).
					WillReturnRows(userRows)
			},
			expectedErr: false,
			validate: func(t *testing.T, orders []models.Order, total int64) {
				assert.Equal(t, int64(25), total)
				assert.Len(t, orders, 2)
				assert.Equal(t, snowflake.ID(1), orders[0].ID)
			},
		},
		{
			name:     "pagination - second page",
			shopID:   123,
			page:     2,
			pageSize: 10,
			setupMock: func(mock sqlmock.Sqlmock) {
				countRows := sqlmock.NewRows([]string{"count"}).AddRow(25)
				mock.ExpectQuery("SELECT count\\(\\*\\) FROM `orders`").
					WithArgs(int64(123)).
					WillReturnRows(countRows)

				orderRows := sqlmock.NewRows([]string{"id", "user_id", "shop_id", "total_price", "status"}).
					AddRow(11, 110, 123, 10000, 1)
				mock.ExpectQuery("SELECT \\* FROM `orders`").
					WithArgs(int64(123), 10, 10).
					WillReturnRows(orderRows)

				itemRows := sqlmock.NewRows([]string{"id", "order_id", "product_id", "quantity"}).
					AddRow(11, 11, 999, 2)
				mock.ExpectQuery("SELECT \\* FROM `order_items`").
					WithArgs(int64(11)).
					WillReturnRows(itemRows)

				// Mock preload User - uses user_id foreign key (110)
				userRows := sqlmock.NewRows([]string{"id", "name", "phone"}).
					AddRow(110, "user11", "13800138110")
				mock.ExpectQuery("SELECT \\* FROM `users`").
					WithArgs(int64(110)).
					WillReturnRows(userRows)
			},
			expectedErr: false,
			validate: func(t *testing.T, orders []models.Order, total int64) {
				assert.Equal(t, int64(25), total)
				assert.Len(t, orders, 1)
			},
		},
		{
			name:     "database error on count",
			shopID:   123,
			page:     1,
			pageSize: 10,
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT count\\(\\*\\) FROM `orders`").
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

			repo := NewOrderRepository(db)
			tt.setupMock(mock)

			orders, total, err := repo.GetOrdersByShop(tt.shopID, tt.page, tt.pageSize)

			if tt.expectedErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tt.validate != nil {
					tt.validate(t, orders, total)
				}
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

// ==================== GetOrdersByUser Tests ====================

func TestOrderRepository_GetOrdersByUser(t *testing.T) {
	tests := []struct {
		name        string
		userID      string
		shopID      snowflake.ID
		page        int
		pageSize    int
		setupMock   func(mock sqlmock.Sqlmock)
		expectedErr bool
		validate    func(t *testing.T, orders []models.Order, total int64)
	}{
		{
			name:     "successfully get user orders",
			userID:   "user123",
			shopID:   456,
			page:     1,
			pageSize: 10,
			setupMock: func(mock sqlmock.Sqlmock) {
				countRows := sqlmock.NewRows([]string{"count"}).AddRow(5)
				mock.ExpectQuery("SELECT count\\(\\*\\) FROM `orders`").
					WithArgs("user123", int64(456)).
					WillReturnRows(countRows)

				orderRows := sqlmock.NewRows([]string{"id", "user_id", "shop_id", "total_price", "status"}).
					AddRow(1, 123, 456, 10000, 1)
				mock.ExpectQuery("SELECT \\* FROM `orders`").
					WithArgs("user123", int64(456), 10).
					WillReturnRows(orderRows)

				itemRows := sqlmock.NewRows([]string{"id", "order_id", "product_id", "quantity"}).
					AddRow(1, 1, 999, 2)
				mock.ExpectQuery("SELECT \\* FROM `order_items`").
					WithArgs(int64(1)).
					WillReturnRows(itemRows)

				optionRows := sqlmock.NewRows([]string{"id", "order_item_id", "category_id", "option_id"}).
					AddRow(1, 1, 10, 100)
				mock.ExpectQuery("SELECT \\* FROM `order_item_options`").
					WithArgs(int64(1)).
					WillReturnRows(optionRows)
			},
			expectedErr: false,
			validate: func(t *testing.T, orders []models.Order, total int64) {
				assert.Equal(t, int64(5), total)
				assert.Len(t, orders, 1)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, sqlDB := setupTestDB(t)
			defer sqlDB.Close()

			repo := NewOrderRepository(db)
			tt.setupMock(mock)

			orders, total, err := repo.GetOrdersByUser(tt.userID, tt.shopID, tt.page, tt.pageSize)

			if tt.expectedErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tt.validate != nil {
					tt.validate(t, orders, total)
				}
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

// ==================== CreateOrder Tests ====================

func TestOrderRepository_CreateOrder(t *testing.T) {
	tests := []struct {
		name        string
		order       *models.Order
		setupMock   func(mock sqlmock.Sqlmock)
		expectedErr bool
		errMsg      string
	}{
		{
			name: "successfully create order",
			order: &models.Order{
				ID:     123,
				UserID: 456,
				ShopID: 789,
				Items: []models.OrderItem{
					{ID: 1, OrderID: 123, ProductID: 999, Quantity: 2},
				},
			},
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO `orders`").
					WillReturnResult(sqlmock.NewResult(123, 1))
				// GORM auto-fills OrderItem fields and saves them
				mock.ExpectExec("INSERT INTO `order_items`").
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			expectedErr: false,
		},
		{
			name: "create order fails",
			order: &models.Order{
				UserID: 456,
				ShopID: 789,
			},
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO `orders`").
					WillReturnError(fmt.Errorf("database error"))
				mock.ExpectRollback()
			},
			expectedErr: true,
			errMsg:      "创建订单失败",
		},
		{
			name: "commit fails",
			order: &models.Order{
				ID:     123,
				UserID: 456,
				ShopID: 789,
			},
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO `orders`").
					WillReturnResult(sqlmock.NewResult(123, 1))
				// No items in order, so no INSERT INTO order_items
				mock.ExpectCommit().WillReturnError(fmt.Errorf("commit error"))
			},
			expectedErr: true,
			errMsg:      "创建订单失败",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, sqlDB := setupTestDB(t)
			defer sqlDB.Close()

			repo := NewOrderRepository(db)
			tt.setupMock(mock)

			err := repo.CreateOrder(tt.order)

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

// ==================== UpdateOrder Tests ====================

func TestOrderRepository_UpdateOrder(t *testing.T) {
	tests := []struct {
		name        string
		order       *models.Order
		newItems    []models.OrderItem
		setupMock   func(mock sqlmock.Sqlmock)
		expectedErr bool
		errMsg      string
	}{
		{
			name: "successfully update order",
			order: &models.Order{
				ID:     123,
				UserID: 456,
				ShopID: 789,
			},
			newItems: []models.OrderItem{
				{ID: 1, ProductID: 999, Quantity: 2},
			},
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec("DELETE FROM `order_items`").
					WillReturnResult(sqlmock.NewResult(1, 2))
				mock.ExpectExec("INSERT INTO `order_items`").
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectExec("UPDATE `orders`").
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			expectedErr: false,
		},
		{
			name: "delete old items fails",
			order: &models.Order{
				ID:     123,
				ShopID: 789,
			},
			newItems: []models.OrderItem{},
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec("DELETE FROM `order_items`").
					WillReturnError(fmt.Errorf("database error"))
				mock.ExpectRollback()
			},
			expectedErr: true,
			errMsg:      "删除原有订单项失败",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, sqlDB := setupTestDB(t)
			defer sqlDB.Close()

			repo := NewOrderRepository(db)
			tt.setupMock(mock)

			err := repo.UpdateOrder(tt.order, tt.newItems)

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

// ==================== DeleteOrder Tests ====================

func TestOrderRepository_DeleteOrder(t *testing.T) {
	tests := []struct {
		name        string
		orderID     string
		shopID      snowflake.ID
		setupMock   func(mock sqlmock.Sqlmock)
		expectedErr bool
		errMsg      string
	}{
		{
			name:    "successfully delete order",
			orderID: "123",
			shopID:  456,
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec("DELETE FROM `order_items`").
					WillReturnResult(sqlmock.NewResult(1, 2))
				mock.ExpectExec("DELETE FROM `order_status_logs`").
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectExec("DELETE FROM `orders`").
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			expectedErr: false,
		},
		{
			name:    "order not found",
			orderID: "999",
			shopID:  456,
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec("DELETE FROM `order_items`").
					WillReturnResult(sqlmock.NewResult(1, 0))
				mock.ExpectExec("DELETE FROM `order_status_logs`").
					WillReturnResult(sqlmock.NewResult(1, 0))
				mock.ExpectExec("DELETE FROM `orders`").
					WillReturnResult(sqlmock.NewResult(1, 0))
				mock.ExpectRollback()
			},
			expectedErr: true,
			errMsg:      "订单不存在",
		},
		{
			name:    "delete items fails",
			orderID: "123",
			shopID:  456,
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec("DELETE FROM `order_items`").
					WillReturnError(fmt.Errorf("database error"))
				mock.ExpectRollback()
			},
			expectedErr: true,
			errMsg:      "删除订单项失败",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, sqlDB := setupTestDB(t)
			defer sqlDB.Close()

			repo := NewOrderRepository(db)
			tt.setupMock(mock)

			err := repo.DeleteOrder(tt.orderID, tt.shopID)

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

// ==================== AdvanceSearch Tests ====================

func TestOrderRepository_AdvanceSearch(t *testing.T) {
	tests := []struct {
		name        string
		req         AdvanceSearchOrderRequest
		setupMock   func(mock sqlmock.Sqlmock)
		expectedErr bool
		validate    func(t *testing.T, result *AdvanceSearchResult)
	}{
		{
			name: "search with status filter",
			req: AdvanceSearchOrderRequest{
				ShopID:   123,
				Page:     1,
				PageSize: 10,
				Status:   []int{1, 2},
			},
			setupMock: func(mock sqlmock.Sqlmock) {
				countRows := sqlmock.NewRows([]string{"count"}).AddRow(15)
				mock.ExpectQuery("SELECT count\\(\\*\\) FROM `orders`").
					WithArgs(int64(123), 1, 2).
					WillReturnRows(countRows)

				orderRows := sqlmock.NewRows([]string{"id", "user_id", "shop_id", "total_price", "status"}).
					AddRow(1, 100, 123, 10000, 1)
				mock.ExpectQuery("SELECT \\* FROM `orders`").
					WithArgs(int64(123), 1, 2, 10).
					WillReturnRows(orderRows)
			},
			expectedErr: false,
			validate: func(t *testing.T, result *AdvanceSearchResult) {
				assert.Equal(t, int64(15), result.Total)
				assert.Len(t, result.Orders, 1)
			},
		},
		{
			name: "search with user and status filter",
			req: AdvanceSearchOrderRequest{
				ShopID:   123,
				Page:     1,
				PageSize: 10,
				UserID:   "user123",
				Status:   []int{1},
			},
			setupMock: func(mock sqlmock.Sqlmock) {
				countRows := sqlmock.NewRows([]string{"count"}).AddRow(5)
				mock.ExpectQuery("SELECT count\\(\\*\\) FROM `orders`").
					WithArgs(int64(123), "user123", 1).
					WillReturnRows(countRows)

				orderRows := sqlmock.NewRows([]string{"id", "user_id", "shop_id", "total_price", "status"}).
					AddRow(1, 456, 123, 10000, 1)
				mock.ExpectQuery("SELECT \\* FROM `orders`").
					WithArgs(int64(123), "user123", 1, 10).
					WillReturnRows(orderRows)
			},
			expectedErr: false,
			validate: func(t *testing.T, result *AdvanceSearchResult) {
				assert.Equal(t, int64(5), result.Total)
				assert.Len(t, result.Orders, 1)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, sqlDB := setupTestDB(t)
			defer sqlDB.Close()

			repo := NewOrderRepository(db)
			tt.setupMock(mock)

			result, err := repo.AdvanceSearch(tt.req)

			if tt.expectedErr {
				assert.Error(t, err)
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
