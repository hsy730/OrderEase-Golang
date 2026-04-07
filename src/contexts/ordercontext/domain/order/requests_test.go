package order

import (
	"testing"

	"github.com/bwmarrin/snowflake"
	"github.com/stretchr/testify/assert"
)

// ==================== CreateOrderRequest Tests ====================

func TestCreateOrderRequest_Validate(t *testing.T) {
	tests := []struct {
		name    string
		req     CreateOrderRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid order request",
			req: CreateOrderRequest{
				UserID: 123,
				ShopID: 456,
				Items: []CreateOrderItemRequest{
					{ProductID: 789, Quantity: 2, Price: 100},
				},
			},
			wantErr: false,
		},
		{
			name: "empty user ID",
			req: CreateOrderRequest{
				UserID: 0,
				ShopID: 456,
				Items: []CreateOrderItemRequest{
					{ProductID: 789, Quantity: 1},
				},
			},
			wantErr: true,
			errMsg:  "用户ID不能为空",
		},
		{
			name: "empty shop ID",
			req: CreateOrderRequest{
				UserID: 123,
				ShopID: 0,
				Items: []CreateOrderItemRequest{
					{ProductID: 789, Quantity: 1},
				},
			},
			wantErr: true,
			errMsg:  "店铺ID不能为空",
		},
		{
			name: "empty items",
			req: CreateOrderRequest{
				UserID: 123,
				ShopID: 456,
				Items:  []CreateOrderItemRequest{},
			},
			wantErr: true,
			errMsg:  "订单项不能为空",
		},
		{
			name: "nil items",
			req: CreateOrderRequest{
				UserID: 123,
				ShopID: 456,
				Items:  nil,
			},
			wantErr: true,
			errMsg:  "订单项不能为空",
		},
		{
			name: "valid request with multiple items and remark",
			req: CreateOrderRequest{
				UserID: 123,
				ShopID: 456,
				Items: []CreateOrderItemRequest{
					{ProductID: 789, Quantity: 2, Price: 100},
					{ProductID: 790, Quantity: 1, Price: 50},
				},
				Remark: "测试备注",
				Status: 1,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.req.Validate()

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// ==================== CreateOrderItemRequest Tests ====================

func TestCreateOrderItemRequest_Validate(t *testing.T) {
	tests := []struct {
		name    string
		req     CreateOrderItemRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid item with options",
			req: CreateOrderItemRequest{
				ProductID: 789,
				Quantity:  2,
				Price:     100.5,
				Options: []CreateOrderItemOption{
					{CategoryID: 1, OptionID: 101},
					{CategoryID: 2, OptionID: 201},
				},
			},
			wantErr: false,
		},
		{
			name: "valid item without options",
			req: CreateOrderItemRequest{
				ProductID: 789,
				Quantity:  1,
				Price:     50,
				Options:   []CreateOrderItemOption{},
			},
			wantErr: false,
		},
		{
			name: "valid item with nil options",
			req: CreateOrderItemRequest{
				ProductID: 789,
				Quantity:  1,
				Price:     50,
				Options:   nil,
			},
			wantErr: false,
		},
		{
			name: "empty product ID",
			req: CreateOrderItemRequest{
				ProductID: 0,
				Quantity:  1,
				Price:     50,
			},
			wantErr: true,
			errMsg:  "商品ID不能为空",
		},
		{
			name: "zero quantity",
			req: CreateOrderItemRequest{
				ProductID: 789,
				Quantity:  0,
				Price:     50,
			},
			wantErr: true,
			errMsg:  "商品数量必须大于0",
		},
		{
			name: "negative quantity",
			req: CreateOrderItemRequest{
				ProductID: 789,
				Quantity:  -1,
				Price:     50,
			},
			wantErr: true,
			errMsg:  "商品数量必须大于0",
		},
		{
			name: "negative price",
			req: CreateOrderItemRequest{
				ProductID: 789,
				Quantity:  1,
				Price:     -10,
			},
			wantErr: true,
			errMsg:  "商品价格不能为负数",
		},
		{
			name: "zero price (allowed)",
			req: CreateOrderItemRequest{
				ProductID: 789,
				Quantity:  1,
				Price:     0,
			},
			wantErr: false,
		},
		{
			name: "large quantity",
			req: CreateOrderItemRequest{
				ProductID: 789,
				Quantity:  9999,
				Price:     100,
			},
			wantErr: false,
		},
		{
			name: "price with decimal",
			req: CreateOrderItemRequest{
				ProductID: 789,
				Quantity:  1,
				Price:     99.99,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.req.Validate()

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// ==================== AdvanceSearchOrderRequest Tests ====================

func TestAdvanceSearchOrderRequest_Validate(t *testing.T) {
	tests := []struct {
		name          string
		req           AdvanceSearchOrderRequest
		wantErr       bool
		errMsg        string
		expectedPage  int
		expectedSize  int
	}{
		{
			name: "valid search request",
			req: AdvanceSearchOrderRequest{
				Page:      1,
				PageSize:  10,
				ShopID:    123,
				UserID:    "user123",
				Status:    []int{0, 1},
				StartTime: "2024-01-01",
				EndTime:   "2024-12-31",
			},
			wantErr:      false,
			expectedPage: 1,
			expectedSize: 10,
		},
		{
			name: "page less than 1 - should default to 1",
			req: AdvanceSearchOrderRequest{
				Page:     0,
				PageSize: 10,
				ShopID:   123,
			},
			wantErr:      false,
			expectedPage: 1,
			expectedSize: 10,
		},
		{
			name: "negative page - should default to 1",
			req: AdvanceSearchOrderRequest{
				Page:     -5,
				PageSize: 10,
				ShopID:   123,
			},
			wantErr:      false,
			expectedPage: 1,
			expectedSize: 10,
		},
		{
			name: "page size too small - should default to 10",
			req: AdvanceSearchOrderRequest{
				Page:     1,
				PageSize: 0,
				ShopID:   123,
			},
			wantErr:      false,
			expectedPage: 1,
			expectedSize: 10,
		},
		{
			name: "page size too large - should default to 10",
			req: AdvanceSearchOrderRequest{
				Page:     1,
				PageSize: 200,
				ShopID:   123,
			},
			wantErr:      false,
			expectedPage: 1,
			expectedSize: 10,
		},
		{
			name: "negative page size - should default to 10",
			req: AdvanceSearchOrderRequest{
				Page:     1,
				PageSize: -10,
				ShopID:   123,
			},
			wantErr:      false,
			expectedPage: 1,
			expectedSize: 10,
		},
		{
			name: "empty shop ID",
			req: AdvanceSearchOrderRequest{
				Page:     1,
				PageSize: 10,
				ShopID:   0,
			},
			wantErr: true,
			errMsg:  "店铺ID不能为空",
		},
		{
			name: "minimal valid request",
			req: AdvanceSearchOrderRequest{
				ShopID: 123,
			},
			wantErr:      false,
			expectedPage: 1,
			expectedSize: 10,
		},
		{
			name: "with empty status array",
			req: AdvanceSearchOrderRequest{
				Page:   1,
				PageSize: 5,
				ShopID: 123,
				Status: []int{},
			},
			wantErr:      false,
			expectedPage: 1,
			expectedSize: 5,
		},
		{
			name: "boundary page size = 1",
			req: AdvanceSearchOrderRequest{
				Page:     1,
				PageSize: 1,
				ShopID:   123,
			},
			wantErr:      false,
			expectedPage: 1,
			expectedSize: 1,
		},
		{
			name: "boundary page size = 100",
			req: AdvanceSearchOrderRequest{
				Page:     1,
				PageSize: 100,
				ShopID:   123,
			},
			wantErr:      false,
			expectedPage: 1,
			expectedSize: 100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.req.Validate()

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedPage, tt.req.Page)
				assert.Equal(t, tt.expectedSize, tt.req.PageSize)
			}
		})
	}
}

// ==================== ToggleOrderStatusRequest Tests ====================

func TestToggleOrderStatusRequest_Validate(t *testing.T) {
	tests := []struct {
		name    string
		req     ToggleOrderStatusRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid status toggle request",
			req: ToggleOrderStatusRequest{
				ID:         789,
				ShopID:     123,
				NextStatus: 1,
			},
			wantErr: false,
		},
		{
			name: "empty order ID",
			req: ToggleOrderStatusRequest{
				ID:         0,
				ShopID:     123,
				NextStatus: 1,
			},
			wantErr: true,
			errMsg:  "订单ID不能为空",
		},
		{
			name: "empty shop ID",
			req: ToggleOrderStatusRequest{
				ID:         789,
				ShopID:     0,
				NextStatus: 1,
			},
			wantErr: true,
			errMsg:  "店铺ID不能为空",
		},
		{
			name: "zero next status (allowed)",
			req: ToggleOrderStatusRequest{
				ID:         789,
				ShopID:     123,
				NextStatus: 0,
			},
			wantErr: false,
		},
		{
			name: "negative next status (allowed)",
			req: ToggleOrderStatusRequest{
				ID:         789,
				ShopID:     123,
				NextStatus: -1,
			},
			wantErr: false,
		},
		{
			name: "all zero values",
			req: ToggleOrderStatusRequest{
				ID:         0,
				ShopID:     0,
				NextStatus: 0,
			},
			wantErr: true,
		},
		{
			name: "large ID values",
			req: ToggleOrderStatusRequest{
				ID:         snowflake.ID(999999999),
				ShopID:     snowflake.ID(888888888),
				NextStatus: 999,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.req.Validate()

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// ==================== Edge Case & Integration Tests ====================

func TestCreateOrderRequest_WithInvalidItems(t *testing.T) {
	tests := []struct {
		name    string
		req     CreateOrderRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "order with valid items structure (item validation is separate)",
			req: CreateOrderRequest{
				UserID: 123,
				ShopID: 456,
				Items: []CreateOrderItemRequest{
					{ProductID: 0, Quantity: 1, Price: 50},
				},
			},
			wantErr: false,
		},
		{
			name: "order with mixed valid/invalid fields",
			req: CreateOrderRequest{
				UserID: 123,
				ShopID: 456,
				Items: []CreateOrderItemRequest{
					{ProductID: 789, Quantity: 2, Price: 100, Options: []CreateOrderItemOption{
						{CategoryID: 1, OptionID: 101},
					}},
				},
				Remark: "",
				Status: 0,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.req.Validate()

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDTOFieldTypes(t *testing.T) {
	t.Run("CreateOrderRequest uses snowflake.ID for IDs", func(t *testing.T) {
		req := CreateOrderRequest{
			ID:     snowflake.ID(123),
			UserID: snowflake.ID(456),
			ShopID: snowflake.ID(789),
		}
		assert.IsType(t, snowflake.ID(0), req.ID)
		assert.IsType(t, snowflake.ID(0), req.UserID)
		assert.IsType(t, snowflake.ID(0), req.ShopID)
	})

	t.Run("CreateOrderItemOption uses snowflake.ID for IDs", func(t *testing.T) {
		opt := CreateOrderItemOption{
			CategoryID: snowflake.ID(1),
			OptionID:   snowflake.ID(2),
		}
		assert.IsType(t, snowflake.ID(0), opt.CategoryID)
		assert.IsType(t, snowflake.ID(0), opt.OptionID)
	})

	t.Run("ToggleOrderStatusRequest uses snowflake.ID for IDs", func(t *testing.T) {
		req := ToggleOrderStatusRequest{
			ID:     snowflake.ID(123),
			ShopID: snowflake.ID(456),
		}
		assert.IsType(t, snowflake.ID(0), req.ID)
		assert.IsType(t, snowflake.ID(0), req.ShopID)
	})
}
