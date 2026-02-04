package shop

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/bwmarrin/snowflake"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
	"orderease/models"
)

// ==================== Constructor Tests ====================

func TestNewShop(t *testing.T) {
	name := "测试店铺"
	ownerUsername := "test_owner"
	validUntil := time.Now().AddDate(1, 0, 0)

	shop := NewShop(name, ownerUsername, validUntil)

	assert.NotNil(t, shop)
	assert.Equal(t, name, shop.Name())
	assert.Equal(t, ownerUsername, shop.OwnerUsername())
	assert.Equal(t, validUntil, shop.ValidUntil())
	assert.Equal(t, "", shop.ContactPhone())
	assert.Equal(t, "", shop.ContactEmail())
	assert.Equal(t, "", shop.Address())
	assert.Equal(t, "", shop.ImageURL())
	assert.Equal(t, "", shop.Description())
	assert.Empty(t, shop.Settings())
	assert.Empty(t, shop.OrderStatusFlow().Statuses)
	assert.False(t, shop.createdAt.IsZero())
	assert.False(t, shop.updatedAt.IsZero())
}

// ==================== Getter Tests ====================

func TestShop_Getters(t *testing.T) {
	shop := NewShop("测试店铺", "test_owner", time.Now().AddDate(1, 0, 0))
	testID := snowflake.ID(456)
	shop.SetID(testID)

	assert.Equal(t, testID, shop.ID())
	assert.Equal(t, "测试店铺", shop.Name())
	assert.Equal(t, "test_owner", shop.OwnerUsername())
	assert.Equal(t, "", shop.OwnerPassword())
	assert.Equal(t, "", shop.ContactPhone())
	assert.Equal(t, "", shop.ContactEmail())
	assert.Equal(t, "", shop.Address())
	assert.Equal(t, "", shop.ImageURL())
	assert.Equal(t, "", shop.Description())
	assert.False(t, shop.ValidUntil().IsZero())
	assert.Empty(t, shop.Settings())
	assert.Empty(t, shop.OrderStatusFlow().Statuses)
	assert.False(t, shop.CreatedAt().IsZero())
	assert.False(t, shop.UpdatedAt().IsZero())
}

// ==================== Setter Tests ====================

func TestShop_Setters(t *testing.T) {
	shop := NewShop("测试店铺", "test_owner", time.Now().AddDate(1, 0, 0))

	// Test SetID
	testID := snowflake.ID(456)
	shop.SetID(testID)
	assert.Equal(t, testID, shop.ID())

	// Test SetName
	shop.SetName("新店铺名")
	assert.Equal(t, "新店铺名", shop.Name())

	// Test SetOwnerUsername
	shop.SetOwnerUsername("new_owner")
	assert.Equal(t, "new_owner", shop.OwnerUsername())

	// Test SetOwnerPassword
	shop.SetOwnerPassword("hashed_password")
	assert.Equal(t, "hashed_password", shop.OwnerPassword())

	// Test SetContactPhone
	shop.SetContactPhone("13800138000")
	assert.Equal(t, "13800138000", shop.ContactPhone())

	// Test SetContactEmail
	shop.SetContactEmail("test@example.com")
	assert.Equal(t, "test@example.com", shop.ContactEmail())

	// Test SetAddress
	shop.SetAddress("测试地址")
	assert.Equal(t, "测试地址", shop.Address())

	// Test SetImageURL
	shop.SetImageURL("http://example.com/image.jpg")
	assert.Equal(t, "http://example.com/image.jpg", shop.ImageURL())

	// Test SetDescription
	shop.SetDescription("测试描述")
	assert.Equal(t, "测试描述", shop.Description())

	// Test SetValidUntil
	newValidUntil := time.Now().AddDate(2, 0, 0)
	shop.SetValidUntil(newValidUntil)
	assert.Equal(t, newValidUntil, shop.ValidUntil())

	// Test SetSettings
	settings := []byte(`{"key": "value"}`)
	shop.SetSettings(settings)
	assert.Equal(t, settings, shop.Settings())

	// Test SetOrderStatusFlow
	flow := models.OrderStatusFlow{Statuses: []models.OrderStatus{
		{Value: 1, Label: "待处理", Type: "warning", IsFinal: false, Actions: []models.OrderStatusAction{}},
	}}
	shop.SetOrderStatusFlow(flow)
	assert.Len(t, shop.OrderStatusFlow().Statuses, 1)

	// Test SetCreatedAt
	now := time.Now()
	shop.SetCreatedAt(now)
	assert.Equal(t, now, shop.CreatedAt())

	// Test SetUpdatedAt
	updated := time.Now()
	shop.SetUpdatedAt(updated)
	assert.Equal(t, updated, shop.UpdatedAt())
}

// ==================== Business Logic Tests ====================

func TestShop_CheckPassword(t *testing.T) {
	// Create a hashed password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("test_password"), bcrypt.DefaultCost)
	assert.NoError(t, err)

	shop := NewShop("测试店铺", "test_owner", time.Now().AddDate(1, 0, 0))
	shop.SetOwnerPassword(string(hashedPassword))

	tests := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{
			name:     "correct password",
			password: "test_password",
			wantErr:  false,
		},
		{
			name:     "incorrect password",
			password: "wrong_password",
			wantErr:  true,
		},
		{
			name:     "empty password",
			password: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := shop.CheckPassword(tt.password)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestShop_IsExpired(t *testing.T) {
	tests := []struct {
		name      string
		validUntil time.Time
		expected  bool
	}{
		{
			name:      "valid future date",
			validUntil: time.Now().AddDate(1, 0, 0),
			expected:  false,
		},
		{
			name:      "expired past date",
			validUntil: time.Now().AddDate(-1, 0, 0),
			expected:  true,
		},
		{
			name:      "expires now - boundary",
			validUntil: time.Now().UTC().Add(-time.Second),
			expected:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			shop := NewShop("测试店铺", "test_owner", tt.validUntil)
			got := shop.IsExpired()
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestShop_IsActive(t *testing.T) {
	tests := []struct {
		name      string
		validUntil time.Time
		expected  bool
	}{
		{
			name:      "valid future - more than 7 days",
			validUntil: time.Now().AddDate(1, 0, 0),
			expected:  true,
		},
		{
			name:      "expired - not active",
			validUntil: time.Now().AddDate(-1, 0, 0),
			expected:  false,
		},
		{
			name:      "expiring soon - less than 7 days",
			validUntil: time.Now().AddDate(0, 0, 3),
			expected:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			shop := NewShop("测试店铺", "test_owner", tt.validUntil)
			got := shop.IsActive()
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestShop_IsExpiringSoon(t *testing.T) {
	tests := []struct {
		name      string
		validUntil time.Time
		expected  bool
	}{
		{
			name:      "expiring in 3 days - true",
			validUntil: time.Now().AddDate(0, 0, 3),
			expected:  true,
		},
		{
			name:      "expiring in 6 days - true",
			validUntil: time.Now().AddDate(0, 0, 6),
			expected:  true,
		},
		{
			name:      "expiring in 7 days - boundary, should be false",
			validUntil: time.Now().AddDate(0, 0, 7).Add(time.Hour), // Add 1 hour to ensure it's just over 7 days
			expected:  false,
		},
		{
			name:      "expiring in 30 days - false",
			validUntil: time.Now().AddDate(0, 0, 30),
			expected:  false,
		},
		{
			name:      "already expired - false",
			validUntil: time.Now().AddDate(-1, 0, 0),
			expected:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			shop := NewShop("测试店铺", "test_owner", tt.validUntil)
			got := shop.IsExpiringSoon()
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestShop_CanDelete(t *testing.T) {
	shop := NewShop("测试店铺", "test_owner", time.Now().AddDate(1, 0, 0))

	tests := []struct {
		name         string
		productCount int
		orderCount   int
		wantErr      bool
		errMsg       string
	}{
		{
			name:         "can delete - no products or orders",
			productCount: 0,
			orderCount:   0,
			wantErr:      false,
		},
		{
			name:         "cannot delete - has products",
			productCount: 5,
			orderCount:   0,
			wantErr:      true,
			errMsg:       "店铺存在 5 个关联商品",
		},
		{
			name:         "cannot delete - has orders",
			productCount: 0,
			orderCount:   3,
			wantErr:      true,
			errMsg:       "店铺存在 3 个关联订单",
		},
		{
			name:         "cannot delete - has both",
			productCount: 5,
			orderCount:   3,
			wantErr:      true,
			errMsg:       "店铺存在 5 个关联商品",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := shop.CanDelete(tt.productCount, tt.orderCount)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestShop_UpdateValidUntil(t *testing.T) {
	tests := []struct {
		name          string
		newValidUntil time.Time
		wantErr       bool
		errMsg        string
	}{
		{
			name:          "valid future date",
			newValidUntil: time.Now().AddDate(2, 0, 0),
			wantErr:       false,
		},
		{
			name:          "past date - error",
			newValidUntil: time.Now().AddDate(-1, 0, 0),
			wantErr:       true,
			errMsg:        "新有效期不能早于当前时间",
		},
		{
			name:          "current time - boundary",
			newValidUntil: time.Now().UTC(),
			wantErr:       true,
			errMsg:        "新有效期不能早于当前时间",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a fresh shop for each test
			shop := NewShop("测试店铺", "test_owner", time.Now().AddDate(1, 0, 0))
			originalUpdatedAt := shop.UpdatedAt()

			// Add a small delay to ensure time difference
			time.Sleep(time.Millisecond)

			err := shop.UpdateValidUntil(tt.newValidUntil)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				assert.Equal(t, originalUpdatedAt, shop.UpdatedAt())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.newValidUntil, shop.ValidUntil())
				assert.True(t, shop.UpdatedAt().After(originalUpdatedAt) || shop.UpdatedAt().Equal(originalUpdatedAt))
			}
		})
	}
}

func TestShop_ValidateOrderStatusFlow(t *testing.T) {
	shop := NewShop("测试店铺", "test_owner", time.Now().AddDate(1, 0, 0))

	tests := []struct {
		name    string
		flow    models.OrderStatusFlow
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid flow with statuses",
			flow: models.OrderStatusFlow{
				Statuses: []models.OrderStatus{
					{Value: 1, Label: "待处理", Type: "warning", IsFinal: false, Actions: []models.OrderStatusAction{}},
					{Value: 2, Label: "已接单", Type: "primary", IsFinal: false, Actions: []models.OrderStatusAction{}},
				},
			},
			wantErr: false,
		},
		{
			name:    "empty flow - error",
			flow:    models.OrderStatusFlow{Statuses: []models.OrderStatus{}},
			wantErr: true,
			errMsg:  "订单流转配置不能为空",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := shop.ValidateOrderStatusFlow(tt.flow)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// ==================== Model Conversion Tests ====================

func TestShop_ToModel(t *testing.T) {
	shopID := snowflake.ID(123)
	validUntil := time.Now().AddDate(1, 0, 0)
	settings := []byte(`{"key": "value"}`)
	flow := models.OrderStatusFlow{Statuses: []models.OrderStatus{
		{Value: 1, Label: "待处理", Type: "warning", IsFinal: false, Actions: []models.OrderStatusAction{}},
	}}

	shop := NewShop("测试店铺", "test_owner", validUntil)
	shop.SetID(shopID)
	shop.SetOwnerPassword("plain_password")
	shop.SetContactPhone("13800138000")
	shop.SetContactEmail("test@example.com")
	shop.SetAddress("测试地址")
	shop.SetImageURL("http://example.com/image.jpg")
	shop.SetDescription("测试描述")
	shop.SetSettings(settings)
	shop.SetOrderStatusFlow(flow)

	model := shop.ToModel()

	assert.Equal(t, shopID, model.ID)
	assert.Equal(t, "测试店铺", model.Name)
	assert.Equal(t, "test_owner", model.OwnerUsername)
	// Password should be hashed, not plain
	assert.NotEqual(t, "plain_password", model.OwnerPassword)
	// Verify the hash is valid by checking password matches
	assert.NoError(t, bcrypt.CompareHashAndPassword([]byte(model.OwnerPassword), []byte("plain_password")))
	assert.Equal(t, "13800138000", model.ContactPhone)
	assert.Equal(t, "test@example.com", model.ContactEmail)
	assert.Equal(t, "测试地址", model.Address)
	assert.Equal(t, "http://example.com/image.jpg", model.ImageURL)
	assert.Equal(t, "测试描述", model.Description)
	assert.Equal(t, validUntil, model.ValidUntil)
	assert.Equal(t, json.RawMessage(settings), model.Settings)
	assert.Len(t, model.OrderStatusFlow.Statuses, 1)
}

func TestShop_ToModel_WithHashedPassword(t *testing.T) {
	shop := NewShop("测试店铺", "test_owner", time.Now().AddDate(1, 0, 0))

	// Already hashed password (bcrypt format starts with $2a$ or $2b$)
	hashedPassword := "$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy"
	shop.SetOwnerPassword(hashedPassword)

	model := shop.ToModel()

	// Should not re-hash an already hashed password
	assert.Equal(t, hashedPassword, model.OwnerPassword)
}

func TestShopFromModel(t *testing.T) {
	shopID := snowflake.ID(123)
	validUntil := time.Now().AddDate(1, 0, 0)
	settings := []byte(`{"key": "value"}`)
	flow := models.OrderStatusFlow{Statuses: []models.OrderStatus{
		{Value: 1, Label: "待处理", Type: "warning", IsFinal: false, Actions: []models.OrderStatusAction{}},
	}}

	model := &models.Shop{
		ID:              shopID,
		Name:            "测试店铺",
		OwnerUsername:   "test_owner",
		OwnerPassword:   "hashed_password",
		ContactPhone:    "13800138000",
		ContactEmail:    "test@example.com",
		Address:         "测试地址",
		ImageURL:        "http://example.com/image.jpg",
		Description:     "测试描述",
		ValidUntil:      validUntil,
		Settings:        settings,
		OrderStatusFlow: flow,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	shop := ShopFromModel(model)

	assert.Equal(t, shopID, shop.ID())
	assert.Equal(t, "测试店铺", shop.Name())
	assert.Equal(t, "test_owner", shop.OwnerUsername())
	assert.Equal(t, "hashed_password", shop.OwnerPassword())
	assert.Equal(t, "13800138000", shop.ContactPhone())
	assert.Equal(t, "test@example.com", shop.ContactEmail())
	assert.Equal(t, "测试地址", shop.Address())
	assert.Equal(t, "http://example.com/image.jpg", shop.ImageURL())
	assert.Equal(t, "测试描述", shop.Description())
	assert.Equal(t, validUntil, shop.ValidUntil())
	// Settings() returns []byte, model.Settings is json.RawMessage (alias for []byte)
	// Use Bytes() for comparison
	assert.Equal(t, settings, shop.Settings())
	assert.Len(t, shop.OrderStatusFlow().Statuses, 1)
}

// ==================== Service Tests ====================

func TestService_ProcessValidUntil(t *testing.T) {
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

func TestService_ParseOrderStatusFlow(t *testing.T) {
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

func TestService_ParseOrderStatusFlow_DefaultFlow(t *testing.T) {
	service := &Service{}

	// Parse default flow
	flow, err := service.ParseOrderStatusFlow(nil)
	assert.NoError(t, err)

	// Verify default flow can be parsed from JSON constant
	var defaultFlow models.OrderStatusFlow
	err = json.Unmarshal([]byte(models.DefaultOrderStatusFlow), &defaultFlow)
	assert.NoError(t, err)

	// Should have the same number of statuses
	assert.Equal(t, len(defaultFlow.Statuses), len(flow.Statuses))
}
