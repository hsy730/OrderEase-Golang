package product

import (
	"testing"
	"time"

	"orderease/domain/shared"

	"github.com/stretchr/testify/assert"
)

func TestNewProduct(t *testing.T) {
	tests := []struct {
		name        string
		shopID      uint64
		productName string
		description string
		price       shared.Price
		stock       int
		wantErr     bool
		errMsg      string
		validate    func(*testing.T, *Product)
	}{
		{
			name:        "valid product",
			shopID:      123,
			productName: "测试商品",
			description: "这是一个测试商品",
			price:       shared.Price(99.99),
			stock:       100,
			wantErr:     false,
			validate: func(t *testing.T, p *Product) {
				assert.Equal(t, shared.ID(0), p.ID)
				assert.Equal(t, uint64(123), p.ShopID)
				assert.Equal(t, "测试商品", p.Name)
				assert.Equal(t, "这是一个测试商品", p.Description)
				assert.Equal(t, shared.Price(99.99), p.Price)
				assert.Equal(t, 100, p.Stock)
				assert.Equal(t, ProductStatusPending, p.Status)
				assert.False(t, p.CreatedAt.IsZero())
				assert.False(t, p.UpdatedAt.IsZero())
			},
		},
		{
			name:        "empty shopID",
			shopID:      0,
			productName: "测试商品",
			description: "描述",
			price:       shared.Price(99.99),
			stock:       100,
			wantErr:     true,
			errMsg:      "店铺ID不能为空",
		},
		{
			name:        "empty name",
			shopID:      123,
			productName: "",
			description: "描述",
			price:       shared.Price(99.99),
			stock:       100,
			wantErr:     true,
			errMsg:      "商品名称不能为空",
		},
		{
			name:        "zero price",
			shopID:      123,
			productName: "测试商品",
			description: "描述",
			price:       shared.Price(0),
			stock:       100,
			wantErr:     true,
			errMsg:      "商品价格不能为零",
		},
		{
			name:        "negative stock",
			shopID:      123,
			productName: "测试商品",
			description: "描述",
			price:       shared.Price(99.99),
			stock:       -1,
			wantErr:     true,
			errMsg:      "商品库存不能为负数",
		},
		{
			name:        "zero stock valid",
			shopID:      123,
			productName: "测试商品",
			description: "描述",
			price:       shared.Price(99.99),
			stock:       0,
			wantErr:     false,
		},
		{
			name:        "empty description valid",
			shopID:      123,
			productName: "测试商品",
			description: "",
			price:       shared.Price(99.99),
			stock:       100,
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewProduct(tt.shopID, tt.productName, tt.description, tt.price, tt.stock)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, got)
				if tt.validate != nil {
					tt.validate(t, got)
				}
			}
		})
	}
}

func TestProduct_UpdateStock(t *testing.T) {
	tests := []struct {
		name        string
		product     *Product
		newStock    int
		wantErr     bool
		errMsg      string
		validateNew int
	}{
		{
			name: "update stock successfully",
			product: &Product{
				Stock: 100,
			},
			newStock:    50,
			wantErr:     false,
			validateNew: 50,
		},
		{
			name: "update to zero",
			product: &Product{
				Stock: 100,
			},
			newStock:    0,
			wantErr:     false,
			validateNew: 0,
		},
		{
			name: "negative stock",
			product: &Product{
				Stock: 100,
			},
			newStock: -1,
			wantErr:  true,
			errMsg:   "库存不能为负数",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			oldUpdatedAt := tt.product.UpdatedAt
			err := tt.product.UpdateStock(tt.newStock)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.validateNew, tt.product.Stock)
				assert.True(t, tt.product.UpdatedAt.After(oldUpdatedAt) || tt.product.UpdatedAt.Equal(oldUpdatedAt))
			}
		})
	}
}

func TestProduct_DecreaseStock(t *testing.T) {
	tests := []struct {
		name          string
		initialStock  int
		quantity      int
		wantErr       bool
		errMsg        string
		validateStock int
	}{
		{
			name:          "decrease successfully",
			initialStock:  100,
			quantity:      30,
			wantErr:       false,
			validateStock: 70,
		},
		{
			name:          "decrease all stock",
			initialStock:  50,
			quantity:      50,
			wantErr:       false,
			validateStock: 0,
		},
		{
			name:          "insufficient stock",
			initialStock:  10,
			quantity:      20,
			wantErr:       true,
			errMsg:        "库存不足",
			validateStock: 10,
		},
		{
			name:          "exact match",
			initialStock:  25,
			quantity:      25,
			wantErr:       false,
			validateStock: 0,
		},
		{
			name:          "decrease by zero",
			initialStock:  100,
			quantity:      0,
			wantErr:       false,
			validateStock: 100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Product{
				Stock:     tt.initialStock,
				UpdatedAt: time.Now(),
			}
			oldUpdatedAt := p.UpdatedAt

			err := p.DecreaseStock(tt.quantity)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				assert.Equal(t, tt.initialStock, p.Stock, "stock should not change on error")
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.validateStock, p.Stock)
				assert.True(t, p.UpdatedAt.After(oldUpdatedAt) || p.UpdatedAt.Equal(oldUpdatedAt))
			}
		})
	}
}

func TestProduct_IncreaseStock(t *testing.T) {
	tests := []struct {
		name          string
		initialStock  int
		quantity      int
		wantErr       bool
		errMsg        string
		validateStock int
	}{
		{
			name:          "increase successfully",
			initialStock:  100,
			quantity:      30,
			wantErr:       false,
			validateStock: 130,
		},
		{
			name:          "increase zero stock",
			initialStock:  0,
			quantity:      50,
			wantErr:       false,
			validateStock: 50,
		},
		{
			name:          "increase by zero",
			initialStock:  100,
			quantity:      0,
			wantErr:       false,
			validateStock: 100,
		},
		{
			name:          "negative quantity",
			initialStock:  100,
			quantity:      -10,
			wantErr:       true,
			errMsg:        "增加的库存数量不能为负数",
			validateStock: 100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Product{
				Stock:     tt.initialStock,
				UpdatedAt: time.Now(),
			}
			oldUpdatedAt := p.UpdatedAt

			err := p.IncreaseStock(tt.quantity)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				assert.Equal(t, tt.initialStock, p.Stock, "stock should not change on error")
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.validateStock, p.Stock)
				assert.True(t, p.UpdatedAt.After(oldUpdatedAt) || p.UpdatedAt.Equal(oldUpdatedAt))
			}
		})
	}
}

func TestProduct_ChangeStatus(t *testing.T) {
	tests := []struct {
		name        string
		initialStatus ProductStatus
		newStatus   ProductStatus
		wantErr     bool
		errMsg      string
	}{
		// Valid transitions
		{"pending to online", ProductStatusPending, ProductStatusOnline, false, ""},
		{"online to offline", ProductStatusOnline, ProductStatusOffline, false, ""},
		{"offline to online", ProductStatusOffline, ProductStatusOnline, false, ""},

		// Invalid transitions
		{"pending to offline", ProductStatusPending, ProductStatusOffline, true, "不允许的状态转换"},
		{"online to pending", ProductStatusOnline, ProductStatusPending, true, "不允许的状态转换"},
		{"offline to pending", ProductStatusOffline, ProductStatusPending, true, "不允许的状态转换"},
		{"same status pending", ProductStatusPending, ProductStatusPending, true, "不允许的状态转换"},
		{"same status online", ProductStatusOnline, ProductStatusOnline, true, "不允许的状态转换"},
		{"same status offline", ProductStatusOffline, ProductStatusOffline, true, "不允许的状态转换"},

		// Invalid status
		{"to invalid status", ProductStatusPending, ProductStatus("unknown"), true, "无效的商品状态"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Product{
				Status:    tt.initialStatus,
				UpdatedAt: time.Now(),
			}
			oldUpdatedAt := p.UpdatedAt

			err := p.ChangeStatus(tt.newStatus)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				assert.Equal(t, tt.initialStatus, p.Status, "status should not change on error")
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.newStatus, p.Status)
				assert.True(t, p.UpdatedAt.After(oldUpdatedAt) || p.UpdatedAt.Equal(oldUpdatedAt))
			}
		})
	}
}

func TestProduct_IsAvailable(t *testing.T) {
	tests := []struct {
		name     string
		status   ProductStatus
		expected bool
	}{
		{"pending is available", ProductStatusPending, true},
		{"online is available", ProductStatusOnline, true},
		{"offline is not available", ProductStatusOffline, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Product{Status: tt.status}
			got := p.IsAvailable()
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestProduct_HasStock(t *testing.T) {
	tests := []struct {
		name     string
		stock    int
		quantity int
		expected bool
	}{
		{"sufficient stock", 100, 50, true},
		{"exact stock", 100, 100, true},
		{"insufficient stock", 50, 100, false},
		{"zero stock", 0, 1, false},
		{"zero quantity", 100, 0, true},
		{"large stock", 10000, 5000, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Product{Stock: tt.stock}
			got := p.HasStock(tt.quantity)
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestProduct_StatusWorkflow(t *testing.T) {
	// Test typical product status workflow
	t.Run("typical workflow: pending -> online -> offline -> online", func(t *testing.T) {
		p, _ := NewProduct(123, "测试商品", "描述", shared.Price(99.99), 100)

		// Initial status should be pending
		assert.Equal(t, ProductStatusPending, p.Status)

		// Transition to online
		err := p.ChangeStatus(ProductStatusOnline)
		assert.NoError(t, err)
		assert.Equal(t, ProductStatusOnline, p.Status)

		// Transition to offline
		err = p.ChangeStatus(ProductStatusOffline)
		assert.NoError(t, err)
		assert.Equal(t, ProductStatusOffline, p.Status)

		// Back to online
		err = p.ChangeStatus(ProductStatusOnline)
		assert.NoError(t, err)
		assert.Equal(t, ProductStatusOnline, p.Status)
	})
}
