package order

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"orderease/models"
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
				UserID: models.SnowflakeString(123),
				ShopID: models.SnowflakeString(456),
				Items: []CreateOrderItemRequest{
					{ProductID: models.SnowflakeString(789), Quantity: 2, Price: 100},
				},
			},
			wantErr: false,
		},
		{
			name: "empty user ID",
			req: CreateOrderRequest{
				UserID: 0,
				ShopID: models.SnowflakeString(456),
				Items: []CreateOrderItemRequest{
					{ProductID: models.SnowflakeString(789), Quantity: 1},
				},
			},
			wantErr: true,
			errMsg:  "用户ID不能为空",
		},
		{
			name: "empty shop ID",
			req: CreateOrderRequest{
				UserID: models.SnowflakeString(123),
				ShopID: 0,
				Items: []CreateOrderItemRequest{
					{ProductID: models.SnowflakeString(789), Quantity: 1},
				},
			},
			wantErr: true,
			errMsg:  "店铺ID不能为空",
		},
		{
			name: "empty items",
			req: CreateOrderRequest{
				UserID: models.SnowflakeString(123),
				ShopID: models.SnowflakeString(456),
				Items:  []CreateOrderItemRequest{},
			},
			wantErr: true,
			errMsg:  "订单项不能为空",
		},
		{
			name: "nil items",
			req: CreateOrderRequest{
				UserID: models.SnowflakeString(123),
				ShopID: models.SnowflakeString(456),
				Items:  nil,
			},
			wantErr: true,
			errMsg:  "订单项不能为空",
		},
		{
			name: "valid request with multiple items and remark",
			req: CreateOrderRequest{
				UserID: models.SnowflakeString(123),
				ShopID: models.SnowflakeString(456),
				Items: []CreateOrderItemRequest{
					{ProductID: models.SnowflakeString(789), Quantity: 2, Price: 100},
					{ProductID: models.SnowflakeString(790), Quantity: 1, Price: 50},
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
				ProductID: models.SnowflakeString(789),
				Quantity:  2,
				Price:     100.5,
				Options: []CreateOrderItemOption{
					{CategoryID: models.SnowflakeString(1), OptionID: models.SnowflakeString(101)},
					{CategoryID: models.SnowflakeString(2), OptionID: models.SnowflakeString(201)},
				},
			},
			wantErr: false,
		},
		{
			name: "valid item without options",
			req: CreateOrderItemRequest{
				ProductID: models.SnowflakeString(789),
				Quantity:  1,
				Price:     50,
				Options:   []CreateOrderItemOption{},
			},
			wantErr: false,
		},
		{
			name: "valid item with nil options",
			req: CreateOrderItemRequest{
				ProductID: models.SnowflakeString(789),
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
				ProductID: models.SnowflakeString(789),
				Quantity:  0,
				Price:     50,
			},
			wantErr: true,
			errMsg:  "商品数量必须大于0",
		},
		{
			name: "negative quantity",
			req: CreateOrderItemRequest{
				ProductID: models.SnowflakeString(789),
				Quantity:  -1,
				Price:     50,
			},
			wantErr: true,
			errMsg:  "商品数量必须大于0",
		},
		{
			name: "negative price",
			req: CreateOrderItemRequest{
				ProductID: models.SnowflakeString(789),
				Quantity:  1,
				Price:     -10,
			},
			wantErr: true,
			errMsg:  "商品价格不能为负数",
		},
		{
			name: "zero price (allowed)",
			req: CreateOrderItemRequest{
				ProductID: models.SnowflakeString(789),
				Quantity:  1,
				Price:     0,
			},
			wantErr: false,
		},
		{
			name: "large quantity",
			req: CreateOrderItemRequest{
				ProductID: models.SnowflakeString(789),
				Quantity:  9999,
				Price:     100,
			},
			wantErr: false,
		},
		{
			name: "price with decimal",
			req: CreateOrderItemRequest{
				ProductID: models.SnowflakeString(789),
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
				ShopID:    models.SnowflakeString(123),
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
				ShopID:   models.SnowflakeString(123),
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
				ShopID:   models.SnowflakeString(123),
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
				ShopID:   models.SnowflakeString(123),
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
				ShopID:   models.SnowflakeString(123),
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
				ShopID:   models.SnowflakeString(123),
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
				ShopID: models.SnowflakeString(123),
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
				ShopID: models.SnowflakeString(123),
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
				ShopID:   models.SnowflakeString(123),
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
				ShopID:   models.SnowflakeString(123),
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
				ID:         models.SnowflakeString(789),
				ShopID:     models.SnowflakeString(123),
				NextStatus: 1,
			},
			wantErr: false,
		},
		{
			name: "empty order ID",
			req: ToggleOrderStatusRequest{
				ID:         0,
				ShopID:     models.SnowflakeString(123),
				NextStatus: 1,
			},
			wantErr: true,
			errMsg:  "订单ID不能为空",
		},
		{
			name: "empty shop ID",
			req: ToggleOrderStatusRequest{
				ID:         models.SnowflakeString(789),
				ShopID:     0,
				NextStatus: 1,
			},
			wantErr: true,
			errMsg:  "店铺ID不能为空",
		},
		{
			name: "zero next status (allowed)",
			req: ToggleOrderStatusRequest{
				ID:         models.SnowflakeString(789),
				ShopID:     models.SnowflakeString(123),
				NextStatus: 0,
			},
			wantErr: false,
		},
		{
			name: "negative next status (allowed)",
			req: ToggleOrderStatusRequest{
				ID:         models.SnowflakeString(789),
				ShopID:     models.SnowflakeString(123),
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
				ID:         models.SnowflakeString(999999999),
				ShopID:     models.SnowflakeString(888888888),
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
				UserID: models.SnowflakeString(123),
				ShopID: models.SnowflakeString(456),
				Items: []CreateOrderItemRequest{
					{ProductID: 0, Quantity: 1, Price: 50},
				},
			},
			wantErr: false,
		},
		{
			name: "order with mixed valid/invalid fields",
			req: CreateOrderRequest{
				UserID: models.SnowflakeString(123),
				ShopID: models.SnowflakeString(456),
				Items: []CreateOrderItemRequest{
					{ProductID: models.SnowflakeString(789), Quantity: 2, Price: 100, Options: []CreateOrderItemOption{
						{CategoryID: models.SnowflakeString(1), OptionID: models.SnowflakeString(101)},
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
	t.Run("CreateOrderRequest uses SnowflakeString for IDs", func(t *testing.T) {
		req := CreateOrderRequest{
			ID:     models.SnowflakeString(123),
			UserID: models.SnowflakeString(456),
			ShopID: models.SnowflakeString(789),
		}
		assert.IsType(t, models.SnowflakeString(0), req.ID)
		assert.IsType(t, models.SnowflakeString(0), req.UserID)
		assert.IsType(t, models.SnowflakeString(0), req.ShopID)
	})

	t.Run("CreateOrderItemOption uses SnowflakeString for IDs", func(t *testing.T) {
		opt := CreateOrderItemOption{
			CategoryID: models.SnowflakeString(1),
			OptionID:   models.SnowflakeString(2),
		}
		assert.IsType(t, models.SnowflakeString(0), opt.CategoryID)
		assert.IsType(t, models.SnowflakeString(0), opt.OptionID)
	})

	t.Run("ToggleOrderStatusRequest uses SnowflakeString for IDs", func(t *testing.T) {
		req := ToggleOrderStatusRequest{
			ID:     models.SnowflakeString(123),
			ShopID: models.SnowflakeString(456),
		}
		assert.IsType(t, models.SnowflakeString(0), req.ID)
		assert.IsType(t, models.SnowflakeString(0), req.ShopID)
	})
}

// ==================== JSON Serialization Tests ====================

func TestSnowflakeString_JSONSerialization(t *testing.T) {
	// 使用有效的 int64 范围内的雪花 ID（最大值为 9223372036854775807）
	t.Run("marshal to JSON string", func(t *testing.T) {
		req := CreateOrderRequest{
			ID:     models.SnowflakeString(1234567890123456789),
			UserID: models.SnowflakeString(876543210987654321),
			ShopID: models.SnowflakeString(1111111111111111111),
		}
		// 使用 json 包序列化
		data, err := json.Marshal(req)
		assert.NoError(t, err)
		// 验证输出包含字符串格式的 ID
		assert.Contains(t, string(data), `"id":"1234567890123456789"`)
		assert.Contains(t, string(data), `"user_id":"876543210987654321"`)
		assert.Contains(t, string(data), `"shop_id":"1111111111111111111"`)
	})

	t.Run("unmarshal from JSON string", func(t *testing.T) {
		jsonStr := `{"id":"1234567890123456789","user_id":"876543210987654321","shop_id":"1111111111111111111","items":[],"remark":"","status":0}`
		var req CreateOrderRequest
		err := json.Unmarshal([]byte(jsonStr), &req)
		assert.NoError(t, err)
		assert.Equal(t, models.SnowflakeString(1234567890123456789), req.ID)
		assert.Equal(t, models.SnowflakeString(876543210987654321), req.UserID)
		assert.Equal(t, models.SnowflakeString(1111111111111111111), req.ShopID)
	})

	t.Run("unmarshal from JSON number", func(t *testing.T) {
		jsonNum := `{"id":1234567890123456789,"user_id":876543210987654321,"shop_id":1111111111111111111,"items":[],"remark":"","status":0}`
		var req CreateOrderRequest
		err := json.Unmarshal([]byte(jsonNum), &req)
		assert.NoError(t, err)
		assert.Equal(t, models.SnowflakeString(1234567890123456789), req.ID)
		assert.Equal(t, models.SnowflakeString(876543210987654321), req.UserID)
		assert.Equal(t, models.SnowflakeString(1111111111111111111), req.ShopID)
	})
}
