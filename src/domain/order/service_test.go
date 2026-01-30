package order

import (
	"database/sql"
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

// ==================== ValidateOrder Tests ====================

func TestOrderService_ValidateOrder(t *testing.T) {
	service := &Service{}

	tests := []struct {
		name      string
		order     *models.Order
		wantErr   bool
		errMsg    string
	}{
		{
			name: "valid order",
			order: &models.Order{
				UserID: 123,
				ShopID: 456,
				Items: []models.OrderItem{
					{ProductID: 789, Quantity: 2},
				},
			},
			wantErr: false,
		},
		{
			name: "empty user ID",
			order: &models.Order{
				UserID: 0,
				ShopID: 456,
				Items: []models.OrderItem{
					{ProductID: 789, Quantity: 2},
				},
			},
			wantErr: true,
			errMsg:  "用户ID不能为空",
		},
		{
			name: "empty shop ID",
			order: &models.Order{
				UserID: 123,
				ShopID: 0,
				Items: []models.OrderItem{
					{ProductID: 789, Quantity: 2},
				},
			},
			wantErr: true,
			errMsg:  "店铺ID不能为空",
		},
		{
			name: "empty items",
			order: &models.Order{
				UserID: 123,
				ShopID: 456,
				Items:  []models.OrderItem{},
			},
			wantErr: true,
			errMsg:  "订单项不能为空",
		},
		{
			name: "nil items",
			order: &models.Order{
				UserID: 123,
				ShopID: 456,
			},
			wantErr: true,
			errMsg:  "订单项不能为空",
		},
		{
			name: "empty product ID",
			order: &models.Order{
				UserID: 123,
				ShopID: 456,
				Items: []models.OrderItem{
					{ProductID: 0, Quantity: 2},
				},
			},
			wantErr: true,
			errMsg:  "商品ID不能为空",
		},
		{
			name: "invalid quantity - zero",
			order: &models.Order{
				UserID: 123,
				ShopID: 456,
				Items: []models.OrderItem{
					{ProductID: 789, Quantity: 0},
				},
			},
			wantErr: true,
			errMsg:  "商品数量必须大于0",
		},
		{
			name: "invalid quantity - negative",
			order: &models.Order{
				UserID: 123,
				ShopID: 456,
				Items: []models.OrderItem{
					{ProductID: 789, Quantity: -1},
				},
			},
			wantErr: true,
			errMsg:  "商品数量必须大于0",
		},
		{
			name: "multiple valid items",
			order: &models.Order{
				UserID: 123,
				ShopID: 456,
				Items: []models.OrderItem{
					{ProductID: 789, Quantity: 2},
					{ProductID: 790, Quantity: 1},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.ValidateOrder(tt.order)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// ==================== CalculateTotal Tests ====================

func TestOrderService_CalculateTotal(t *testing.T) {
	service := &Service{}

	tests := []struct {
		name          string
		order         *models.Order
		expectedTotal float64
	}{
		{
			name: "single item without options",
			order: &models.Order{
				Items: []models.OrderItem{
					{
						Quantity: 2,
						Price:    10000, // 100.00
						Options:  []models.OrderItemOption{},
					},
				},
			},
			expectedTotal: 20000, // 2 * 100.00
		},
		{
			name: "single item with options",
			order: &models.Order{
				Items: []models.OrderItem{
					{
						Quantity: 2,
						Price:    10000, // 100.00
						Options: []models.OrderItemOption{
							{PriceAdjustment: 500},  // +5.00
							{PriceAdjustment: -200}, // -2.00
						},
					},
				},
			},
			expectedTotal: 20600, // 2 * (100.00 + 5.00 - 2.00) = 206.00
		},
		{
			name: "multiple items",
			order: &models.Order{
				Items: []models.OrderItem{
					{
						Quantity: 2,
						Price:    10000,
						Options:  []models.OrderItemOption{},
					},
					{
						Quantity: 1,
						Price:    5000,
						Options:  []models.OrderItemOption{},
					},
				},
			},
			expectedTotal: 25000, // 2*100 + 1*50 = 250.00
		},
		{
			name: "multiple items with options",
			order: &models.Order{
				Items: []models.OrderItem{
					{
						Quantity: 2,
						Price:    10000,
						Options: []models.OrderItemOption{
							{PriceAdjustment: 500},
						},
					},
					{
						Quantity: 3,
						Price:    5000,
						Options: []models.OrderItemOption{
							{PriceAdjustment: 200},
							{PriceAdjustment: 300},
						},
					},
				},
			},
			expectedTotal: 37500, // 2*(100+5) + 3*(50+2+3) = 210 + 165 = 375
		},
		{
			name: "empty items",
			order: &models.Order{
				Items: []models.OrderItem{},
			},
			expectedTotal: 0,
		},
		{
			name: "item with negative price adjustment",
			order: &models.Order{
				Items: []models.OrderItem{
					{
						Quantity: 1,
						Price:    10000,
						Options: []models.OrderItemOption{
							{PriceAdjustment: -1000}, // -10.00 discount
						},
					},
				},
			},
			expectedTotal: 9000, // 100 - 10 = 90.00
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			total := service.CalculateTotal(tt.order)
			assert.Equal(t, tt.expectedTotal, total)
		})
	}
}

// ==================== ValidateStatusTransition Tests ====================

func TestOrderService_ValidateStatusTransition(t *testing.T) {
	service := &Service{}

	// 创建一个测试用的状态流转配置
	testFlow := models.OrderStatusFlow{
		Statuses: []models.OrderStatus{
			{
				Value:   0,
				Label:   "待处理",
				Type:    "warning",
				IsFinal: false,
				Actions: []models.OrderStatusAction{
					{Name: "接单", NextStatus: 1, NextStatusLabel: "已接单"},
					{Name: "取消", NextStatus: 9, NextStatusLabel: "已取消"},
				},
			},
			{
				Value:   1,
				Label:   "已接单",
				Type:    "primary",
				IsFinal: false,
				Actions: []models.OrderStatusAction{
					{Name: "发货", NextStatus: 4, NextStatusLabel: "已发货"},
					{Name: "取消", NextStatus: 9, NextStatusLabel: "已取消"},
				},
			},
			{
				Value:   4,
				Label:   "已发货",
				Type:    "info",
				IsFinal: false,
				Actions: []models.OrderStatusAction{
					{Name: "完成", NextStatus: 10, NextStatusLabel: "已完成"},
				},
			},
			{
				Value:   9,
				Label:   "已取消",
				Type:    "danger",
				IsFinal: true,
				Actions: []models.OrderStatusAction{},
			},
			{
				Value:   10,
				Label:   "已完成",
				Type:    "success",
				IsFinal: true,
				Actions: []models.OrderStatusAction{},
			},
		},
	}

	tests := []struct {
		name          string
		currentStatus int
		nextStatus    int
		flow          models.OrderStatusFlow
		wantErr       bool
		errMsg        string
	}{
		{
			name:          "valid transition: pending to accepted",
			currentStatus: 0,
			nextStatus:    1,
			flow:          testFlow,
			wantErr:       false,
		},
		{
			name:          "valid transition: pending to cancelled",
			currentStatus: 0,
			nextStatus:    9,
			flow:          testFlow,
			wantErr:       false,
		},
		{
			name:          "valid transition: accepted to shipped",
			currentStatus: 1,
			nextStatus:    4,
			flow:          testFlow,
			wantErr:       false,
		},
		{
			name:          "valid transition: shipped to completed",
			currentStatus: 4,
			nextStatus:    10,
			flow:          testFlow,
			wantErr:       false,
		},
		{
			name:          "invalid transition: accepted to pending",
			currentStatus: 1,
			nextStatus:    0,
			flow:          testFlow,
			wantErr:       true,
			errMsg:        "当前状态不允许转换到指定的下一个状态",
		},
		{
			name:          "invalid transition: from final state",
			currentStatus: 10,
			nextStatus:    1,
			flow:          testFlow,
			wantErr:       true,
			errMsg:        "当前状态为终态，不允许转换",
		},
		{
			name:          "invalid transition: from cancelled",
			currentStatus: 9,
			nextStatus:    0,
			flow:          testFlow,
			wantErr:       true,
			errMsg:        "当前状态为终态，不允许转换",
		},
		{
			name:          "invalid transition: unknown current status",
			currentStatus: 99,
			nextStatus:    100,
			flow:          testFlow,
			wantErr:       true,
			errMsg:        "当前状态不允许转换",
		},
		{
			name:          "invalid transition: pending to shipped (skip step)",
			currentStatus: 0,
			nextStatus:    4,
			flow:          testFlow,
			wantErr:       true,
			errMsg:        "当前状态不允许转换到指定的下一个状态",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.ValidateStatusTransition(tt.currentStatus, tt.nextStatus, tt.flow)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// ==================== CreateOrder Tests ====================

func TestOrderService_CreateOrder(t *testing.T) {
	tests := []struct {
		name          string
		dto           CreateOrderDTO
		setupMock     func(mock sqlmock.Sqlmock)
		expectedErr   bool
		expectedMsg   string
		validateOrder func(t *testing.T, order *models.Order, total float64)
	}{
		{
			name: "successfully create order without options",
			dto: CreateOrderDTO{
				UserID: 123,
				ShopID: 456,
				Items: []CreateOrderItemDTO{
					{
						ProductID: 789,
						Quantity:  2,
						Options:   []CreateOrderItemOptionDTO{},
					},
				},
				Remark: "test order",
			},
			setupMock: func(mock sqlmock.Sqlmock) {
				// Mock product query
				productRows := sqlmock.NewRows([]string{"id", "name", "description", "image_url", "price", "stock"}).
					AddRow(789, "Test Product", "Test Description", "http://example.com/image.jpg", 10000, 10)
				mock.ExpectQuery("SELECT \\* FROM `products`").
					WithArgs(int64(789), 1).
					WillReturnRows(productRows)

				// Mock product save (stock deduction)
				mock.ExpectBegin()
				mock.ExpectExec("UPDATE `products`").
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			expectedErr: false,
			validateOrder: func(t *testing.T, order *models.Order, total float64) {
				assert.Equal(t, snowflake.ID(123), order.UserID)
				assert.Equal(t, snowflake.ID(456), order.ShopID)
				assert.Equal(t, "test order", order.Remark)
				assert.Equal(t, models.OrderStatusPending, order.Status)
				assert.Len(t, order.Items, 1)
				assert.Equal(t, "Test Product", order.Items[0].ProductName)
				assert.Equal(t, models.Price(10000), order.Items[0].Price)
				assert.Equal(t, float64(20000), total) // 2 * 10000
			},
		},
		{
			name: "product not found",
			dto: CreateOrderDTO{
				UserID: 123,
				ShopID: 456,
				Items: []CreateOrderItemDTO{
					{
						ProductID: 999,
						Quantity:  2,
						Options:   []CreateOrderItemOptionDTO{},
					},
				},
			},
			setupMock: func(mock sqlmock.Sqlmock) {
				// Mock product not found
				mock.ExpectQuery("SELECT \\* FROM `products`").
					WithArgs(int64(999), 1).
					WillReturnError(gorm.ErrRecordNotFound)
			},
			expectedErr: true,
			expectedMsg: "商品不存在",
		},
		{
			name: "insufficient stock",
			dto: CreateOrderDTO{
				UserID: 123,
				ShopID: 456,
				Items: []CreateOrderItemDTO{
					{
						ProductID: 789,
						Quantity:  10,
						Options:   []CreateOrderItemOptionDTO{},
					},
				},
			},
			setupMock: func(mock sqlmock.Sqlmock) {
				// Mock product with insufficient stock
				productRows := sqlmock.NewRows([]string{"id", "name", "description", "image_url", "price", "stock"}).
					AddRow(789, "Test Product", "Test Description", "http://example.com/image.jpg", 10000, 5)
				mock.ExpectQuery("SELECT \\* FROM `products`").
					WithArgs(int64(789), 1).
					WillReturnRows(productRows)
			},
			expectedErr: true,
			expectedMsg: "库存不足",
		},
		{
			name: "order with options",
			dto: CreateOrderDTO{
				UserID: 123,
				ShopID: 456,
				Items: []CreateOrderItemDTO{
					{
						ProductID: 789,
						Quantity:  2,
						Options: []CreateOrderItemOptionDTO{
							{OptionID: 100, CategoryID: 200},
						},
					},
				},
			},
			setupMock: func(mock sqlmock.Sqlmock) {
				// Mock product query
				productRows := sqlmock.NewRows([]string{"id", "name", "description", "image_url", "price", "stock"}).
					AddRow(789, "Test Product", "Test Description", "http://example.com/image.jpg", 10000, 10)
				mock.ExpectQuery("SELECT \\* FROM `products`").
					WithArgs(int64(789), 1).
					WillReturnRows(productRows)

				// Mock option query
				optionRows := sqlmock.NewRows([]string{"id", "name", "category_id", "price_adjustment"}).
					AddRow(100, "Large", 200, 500)
				mock.ExpectQuery("SELECT \\* FROM `product_options`").
					WithArgs(int64(100), 1).
					WillReturnRows(optionRows)

				// Mock category query
				categoryRows := sqlmock.NewRows([]string{"id", "name", "product_id"}).
					AddRow(200, "Size", 789)
				mock.ExpectQuery("SELECT \\* FROM `product_option_categories`").
					WithArgs(int64(200), 1).
					WillReturnRows(categoryRows)

				// Mock product save
				mock.ExpectBegin()
				mock.ExpectExec("UPDATE `products`").
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			expectedErr: false,
			validateOrder: func(t *testing.T, order *models.Order, total float64) {
				assert.Len(t, order.Items[0].Options, 1)
				assert.Equal(t, "Large", order.Items[0].Options[0].OptionName)
				assert.Equal(t, "Size", order.Items[0].Options[0].CategoryName)
				assert.Equal(t, float64(500), order.Items[0].Options[0].PriceAdjustment)
				assert.Equal(t, float64(21000), total)
			},
		},
		{
			name: "option not found",
			dto: CreateOrderDTO{
				UserID: 123,
				ShopID: 456,
				Items: []CreateOrderItemDTO{
					{
						ProductID: 789,
						Quantity:  2,
						Options: []CreateOrderItemOptionDTO{
							{OptionID: 999, CategoryID: 200},
						},
					},
				},
			},
			setupMock: func(mock sqlmock.Sqlmock) {
				// Mock product query
				productRows := sqlmock.NewRows([]string{"id", "name", "description", "image_url", "price", "stock"}).
					AddRow(789, "Test Product", "Test Description", "http://example.com/image.jpg", 10000, 10)
				mock.ExpectQuery("SELECT \\* FROM `products`").
					WithArgs(int64(789), 1).
					WillReturnRows(productRows)

				// Mock option not found
				mock.ExpectQuery("SELECT \\* FROM `product_options`").
					WithArgs(int64(999), 1).
					WillReturnError(gorm.ErrRecordNotFound)
			},
			expectedErr: true,
			expectedMsg: "商品参数选项不存在",
		},
		{
			name: "option belongs to different product",
			dto: CreateOrderDTO{
				UserID: 123,
				ShopID: 456,
				Items: []CreateOrderItemDTO{
					{
						ProductID: 789,
						Quantity:  2,
						Options: []CreateOrderItemOptionDTO{
							{OptionID: 100, CategoryID: 200},
						},
					},
				},
			},
			setupMock: func(mock sqlmock.Sqlmock) {
				// Mock product query
				productRows := sqlmock.NewRows([]string{"id", "name", "description", "image_url", "price", "stock"}).
					AddRow(789, "Test Product", "Test Description", "http://example.com/image.jpg", 10000, 10)
				mock.ExpectQuery("SELECT \\* FROM `products`").
					WithArgs(int64(789), 1).
					WillReturnRows(productRows)

				// Mock option query
				optionRows := sqlmock.NewRows([]string{"id", "name", "category_id", "price_adjustment"}).
					AddRow(100, "Large", 200, 500)
				mock.ExpectQuery("SELECT \\* FROM `product_options`").
					WithArgs(int64(100), 1).
					WillReturnRows(optionRows)

				// Mock category query - belongs to different product
				categoryRows := sqlmock.NewRows([]string{"id", "name", "product_id"}).
					AddRow(200, "Size", 888) // Different product ID
				mock.ExpectQuery("SELECT \\* FROM `product_option_categories`").
					WithArgs(int64(200), 1).
					WillReturnRows(categoryRows)
			},
			expectedErr: true,
			expectedMsg: "参数选项不属于指定商品",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, sqlDB := setupTestDB(t)
			defer sqlDB.Close()

			service := NewService(db)

			tt.setupMock(mock)

			order, total, err := service.CreateOrder(tt.dto)

			if tt.expectedErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedMsg)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, order)
				if tt.validateOrder != nil {
					tt.validateOrder(t, order, total)
				}
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

// ==================== UpdateOrder Tests ====================

func TestOrderService_UpdateOrder(t *testing.T) {
	tests := []struct {
		name          string
		dto           UpdateOrderDTO
		setupMock     func(mock sqlmock.Sqlmock)
		expectedErr   bool
		expectedMsg   string
		validateOrder func(t *testing.T, order *models.Order, total float64)
	}{
		{
			name: "successfully update order",
			dto: UpdateOrderDTO{
				OrderID: 111,
				ShopID:  456,
				Items: []CreateOrderItemDTO{
					{
						ProductID: 789,
						Quantity:  3,
						Options:   []CreateOrderItemOptionDTO{},
					},
				},
				Remark: "updated order",
				Status: models.OrderStatusAccepted,
			},
			setupMock: func(mock sqlmock.Sqlmock) {
				// Mock product query
				productRows := sqlmock.NewRows([]string{"id", "name", "description", "image_url", "price", "stock"}).
					AddRow(789, "Test Product", "Test Description", "http://example.com/image.jpg", 10000, 10)
				mock.ExpectQuery("SELECT \\* FROM `products`").
					WithArgs(int64(789), 1).
					WillReturnRows(productRows)

				// Mock product save
				mock.ExpectBegin()
				mock.ExpectExec("UPDATE `products`").
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			expectedErr: false,
			validateOrder: func(t *testing.T, order *models.Order, total float64) {
				assert.Equal(t, snowflake.ID(111), order.ID)
				assert.Equal(t, snowflake.ID(456), order.ShopID)
				assert.Equal(t, "updated order", order.Remark)
				assert.Equal(t, models.OrderStatusAccepted, order.Status)
				assert.Equal(t, float64(30000), total) // 3 * 10000
			},
		},
		{
			name: "update with insufficient stock",
			dto: UpdateOrderDTO{
				OrderID: 111,
				ShopID:  456,
				Items: []CreateOrderItemDTO{
					{
						ProductID: 789,
						Quantity:  20,
						Options:   []CreateOrderItemOptionDTO{},
					},
				},
			},
			setupMock: func(mock sqlmock.Sqlmock) {
				// Mock product with insufficient stock
				productRows := sqlmock.NewRows([]string{"id", "name", "description", "image_url", "price", "stock"}).
					AddRow(789, "Test Product", "Test Description", "http://example.com/image.jpg", 10000, 5)
				mock.ExpectQuery("SELECT \\* FROM `products`").
					WithArgs(int64(789), 1).
					WillReturnRows(productRows)
			},
			expectedErr: true,
			expectedMsg: "库存不足",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, sqlDB := setupTestDB(t)
			defer sqlDB.Close()

			service := NewService(db)

			tt.setupMock(mock)

			order, total, err := service.UpdateOrder(tt.dto)

			if tt.expectedErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedMsg)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, order)
				if tt.validateOrder != nil {
					tt.validateOrder(t, order, total)
				}
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

// ==================== RestoreStock Tests ====================

func TestOrderService_RestoreStock(t *testing.T) {
	tests := []struct {
		name        string
		order       models.Order
		setupMock   func(mock sqlmock.Sqlmock)
		expectedErr bool
		expectedMsg string
	}{
		{
			name: "successfully restore stock",
			order: models.Order{
				Items: []models.OrderItem{
					{ProductID: 789, Quantity: 2},
				},
			},
			setupMock: func(mock sqlmock.Sqlmock) {
				// Mock product query
				productRows := sqlmock.NewRows([]string{"id", "name", "stock"}).
					AddRow(789, "Test Product", 5)
				mock.ExpectQuery("SELECT \\* FROM `products`").
					WithArgs(int64(789), 1).
					WillReturnRows(productRows)

				// Mock product save
				mock.ExpectBegin()
				mock.ExpectExec("UPDATE `products`").
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			expectedErr: false,
		},
		{
			name: "product not found",
			order: models.Order{
				Items: []models.OrderItem{
					{ProductID: 999, Quantity: 2},
				},
			},
			setupMock: func(mock sqlmock.Sqlmock) {
				// Mock product not found
				mock.ExpectQuery("SELECT \\* FROM `products`").
					WithArgs(int64(999), 1).
					WillReturnError(gorm.ErrRecordNotFound)
			},
			expectedErr: true,
			expectedMsg: "商品不存在",
		},
		{
			name: "multiple items restore stock",
			order: models.Order{
				Items: []models.OrderItem{
					{ProductID: 789, Quantity: 2},
					{ProductID: 790, Quantity: 1},
				},
			},
			setupMock: func(mock sqlmock.Sqlmock) {
				// Mock first product query
				productRows1 := sqlmock.NewRows([]string{"id", "name", "stock"}).
					AddRow(789, "Product 1", 5)
				mock.ExpectQuery("SELECT \\* FROM `products`").
					WithArgs(int64(789), 1).
					WillReturnRows(productRows1)

				// Mock first product save
				mock.ExpectBegin()
				mock.ExpectExec("UPDATE `products`").
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()

				// Mock second product query
				productRows2 := sqlmock.NewRows([]string{"id", "name", "stock"}).
					AddRow(790, "Product 2", 3)
				mock.ExpectQuery("SELECT \\* FROM `products`").
					WithArgs(int64(790), 1).
					WillReturnRows(productRows2)

				// Mock second product save
				mock.ExpectBegin()
				mock.ExpectExec("UPDATE `products`").
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			expectedErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, sqlDB := setupTestDB(t)
			defer sqlDB.Close()

			service := NewService(db)

			tt.setupMock(mock)

			err := service.RestoreStock(db, tt.order)

			if tt.expectedErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedMsg)
			} else {
				assert.NoError(t, err)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
