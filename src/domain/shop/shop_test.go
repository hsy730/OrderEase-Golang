package shop

import (
	"testing"
	"time"

	"orderease/domain/order"

	"github.com/stretchr/testify/assert"
)

func TestNewShop(t *testing.T) {
	tests := []struct {
		name          string
		shopName      string
		ownerUsername string
		ownerPassword string
		validUntil    time.Time
		wantErr       bool
		errMsg        string
		validate      func(*testing.T, *Shop)
	}{
		{
			name:          "valid shop with validUntil",
			shopName:      "测试店铺",
			ownerUsername: "test_user",
			ownerPassword: "test_pass",
			validUntil:    time.Now().AddDate(1, 0, 0),
			wantErr:       false,
			validate: func(t *testing.T, s *Shop) {
				assert.Equal(t, "测试店铺", s.Name)
				assert.Equal(t, "test_user", s.OwnerUsername)
				assert.Equal(t, "test_pass", s.OwnerPassword)
				assert.False(t, s.CreatedAt.IsZero())
				assert.False(t, s.UpdatedAt.IsZero())
				assert.False(t, s.ValidUntil.IsZero())
				assert.NotEmpty(t, s.OrderStatusFlow.Statuses)
			},
		},
		{
			name:          "valid shop with zero validUntil (defaults to 1 year)",
			shopName:      "测试店铺",
			ownerUsername: "test_user",
			ownerPassword: "test_pass",
			validUntil:    time.Time{},
			wantErr:       false,
			validate: func(t *testing.T, s *Shop) {
				expectedMin := time.Now().AddDate(1, 0, 0).Add(-time.Second)
				expectedMax := time.Now().AddDate(1, 0, 0).Add(time.Second)
				assert.True(t, s.ValidUntil.After(expectedMin) || s.ValidUntil.Equal(expectedMin))
				assert.True(t, s.ValidUntil.Before(expectedMax) || s.ValidUntil.Equal(expectedMax))
			},
		},
		{
			name:          "empty shop name",
			shopName:      "",
			ownerUsername: "test_user",
			ownerPassword: "test_pass",
			validUntil:    time.Now().AddDate(1, 0, 0),
			wantErr:       true,
			errMsg:        "店铺名称不能为空",
		},
		{
			name:          "empty owner username",
			shopName:      "测试店铺",
			ownerUsername: "",
			ownerPassword: "test_pass",
			validUntil:    time.Now().AddDate(1, 0, 0),
			wantErr:       true,
			errMsg:        "店主用户名不能为空",
		},
		{
			name:          "empty owner password",
			shopName:      "测试店铺",
			ownerUsername: "test_user",
			ownerPassword: "",
			validUntil:    time.Now().AddDate(1, 0, 0),
			wantErr:       true,
			errMsg:        "店主密码不能为空",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewShop(tt.shopName, tt.ownerUsername, tt.ownerPassword, tt.validUntil)

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

func TestShop_IsExpired(t *testing.T) {
	tests := []struct {
		name       string
		validUntil time.Time
		want       bool
	}{
		{
			name:       "not expired - future date",
			validUntil: time.Now().UTC().AddDate(1, 0, 0),
			want:       false,
		},
		{
			name:       "not expired - tomorrow",
			validUntil: time.Now().UTC().AddDate(0, 0, 1),
			want:       false,
		},
		{
			name:       "not expired - later today",
			validUntil: time.Now().UTC().Add(time.Hour),
			want:       false,
		},
		{
			name:       "expired - past date",
			validUntil: time.Now().UTC().AddDate(-1, 0, 0),
			want:       true,
		},
		{
			name:       "expired - yesterday",
			validUntil: time.Now().UTC().AddDate(0, 0, -1),
			want:       true,
		},
		{
			name:       "expired - earlier today",
			validUntil: time.Now().UTC().Add(-time.Hour),
			want:       true,
		},
		{
			name:       "boundary - exact now (not expired, Before returns false)",
			validUntil: time.Now().UTC().Add(time.Second),
			want:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Shop{ValidUntil: tt.validUntil}
			got := s.IsExpired()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestShop_RemainingDays(t *testing.T) {
	tests := []struct {
		name         string
		validUntil   time.Time
		minRemaining int
		maxRemaining int
	}{
		{
			name:         "365 days remaining",
			validUntil:   time.Now().UTC().AddDate(1, 0, 0),
			minRemaining: 364,
			maxRemaining: 365,
		},
		{
			name:         "30 days remaining",
			validUntil:   time.Now().UTC().AddDate(0, 0, 30),
			minRemaining: 29,
			maxRemaining: 30,
		},
		{
			name:         "1 day remaining",
			validUntil:   time.Now().UTC().AddDate(0, 0, 1),
			minRemaining: 0,
			maxRemaining: 1,
		},
		{
			name:         "expired by 1 day",
			validUntil:   time.Now().UTC().AddDate(0, 0, -1),
			minRemaining: -1,
			maxRemaining: -1,
		},
		{
			name:         "expired by 30 days",
			validUntil:   time.Now().UTC().AddDate(0, 0, -30),
			minRemaining: -30,
			maxRemaining: -30,
		},
		{
			name:         "expired by 365 days",
			validUntil:   time.Now().UTC().AddDate(-1, 0, 0),
			minRemaining: -365,
			maxRemaining: -365,
		},
		{
			name:         "less than a day remaining",
			validUntil:   time.Now().UTC().Add(time.Hour * 12),
			minRemaining: 0,
			maxRemaining: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Shop{ValidUntil: tt.validUntil}
			got := s.RemainingDays()
			assert.GreaterOrEqual(t, got, tt.minRemaining)
			assert.LessOrEqual(t, got, tt.maxRemaining)
		})
	}
}

func TestShop_UpdateBasicInfo(t *testing.T) {
	tests := []struct {
		name         string
		shop         *Shop
		shopName     string
		contactPhone string
		contactEmail string
		address      string
		description  string
		wantErr      bool
		validate     func(*testing.T, *Shop)
	}{
		{
			name: "update all fields",
			shop: &Shop{
				Name:         "原店名",
				ContactPhone: "",
				ContactEmail: "",
				Address:      "",
				Description:  "",
			},
			shopName:     "新店名",
			contactPhone: "123456789",
			contactEmail: "test@example.com",
			address:      "测试地址",
			description:  "测试描述",
			wantErr:      false,
			validate: func(t *testing.T, s *Shop) {
				assert.Equal(t, "新店名", s.Name)
				assert.Equal(t, "123456789", s.ContactPhone)
				assert.Equal(t, "test@example.com", s.ContactEmail)
				assert.Equal(t, "测试地址", s.Address)
				assert.Equal(t, "测试描述", s.Description)
			},
		},
		{
			name: "update only some fields",
			shop: &Shop{
				Name:         "原店名",
				ContactPhone: "原电话",
				ContactEmail: "",
				Address:      "",
				Description:  "",
			},
			shopName:     "",
			contactPhone: "新电话",
			contactEmail: "test@example.com",
			address:      "",
			description:  "",
			wantErr:      false,
			validate: func(t *testing.T, s *Shop) {
				assert.Equal(t, "原店名", s.Name, "name should not change")
				assert.Equal(t, "新电话", s.ContactPhone)
				assert.Equal(t, "test@example.com", s.ContactEmail)
			},
		},
		{
			name: "update with empty strings keeps original",
			shop: &Shop{
				Name:         "原店名",
				ContactPhone: "原电话",
				ContactEmail: "原邮箱",
				Address:      "原地址",
				Description:  "原描述",
			},
			shopName:     "",
			contactPhone: "",
			contactEmail: "",
			address:      "",
			description:  "",
			wantErr:      false,
			validate: func(t *testing.T, s *Shop) {
				assert.Equal(t, "原店名", s.Name)
				assert.Equal(t, "原电话", s.ContactPhone)
				assert.Equal(t, "原邮箱", s.ContactEmail)
				assert.Equal(t, "原地址", s.Address)
				assert.Equal(t, "原描述", s.Description)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			oldUpdatedAt := tt.shop.UpdatedAt
			err := tt.shop.UpdateBasicInfo(tt.shopName, tt.contactPhone, tt.contactEmail, tt.address, tt.description)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.True(t, tt.shop.UpdatedAt.After(oldUpdatedAt) || tt.shop.UpdatedAt.Equal(oldUpdatedAt))
				if tt.validate != nil {
					tt.validate(t, tt.shop)
				}
			}
		})
	}
}

func TestShop_UpdateOrderStatusFlow(t *testing.T) {
	tests := []struct {
		name    string
		shop    *Shop
		flow    order.OrderStatusFlow
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid flow",
			shop: &Shop{},
			flow: order.OrderStatusFlow{
				Statuses: []order.OrderStatusConfig{
					{Value: 0, Label: "待处理", IsFinal: false},
				},
			},
			wantErr: false,
		},
		{
			name:    "empty flow",
			shop:    &Shop{},
			flow:    order.OrderStatusFlow{Statuses: []order.OrderStatusConfig{}},
			wantErr: true,
			errMsg:  "订单流转配置不能为空",
		},
		{
			name: "flow with multiple statuses",
			shop: &Shop{},
			flow: order.OrderStatusFlow{
				Statuses: []order.OrderStatusConfig{
					{Value: 0, Label: "待处理", IsFinal: false},
					{Value: 1, Label: "已完成", IsFinal: true},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			oldUpdatedAt := tt.shop.UpdatedAt
			err := tt.shop.UpdateOrderStatusFlow(tt.flow)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
				assert.True(t, tt.shop.UpdatedAt.After(oldUpdatedAt) || tt.shop.UpdatedAt.Equal(oldUpdatedAt))
				assert.Equal(t, tt.flow, tt.shop.OrderStatusFlow)
			}
		})
	}
}

func TestShop_UpdateValidUntil(t *testing.T) {
	tests := []struct {
		name       string
		shop       *Shop
		validUntil time.Time
		wantErr    bool
		errMsg     string
	}{
		{
			name:       "valid future date",
			shop:       &Shop{},
			validUntil: time.Now().AddDate(1, 0, 0),
			wantErr:    false,
		},
		{
			name:       "valid past date",
			shop:       &Shop{},
			validUntil: time.Now().AddDate(-1, 0, 0),
			wantErr:    false,
		},
		{
			name:       "zero time",
			shop:       &Shop{},
			validUntil: time.Time{},
			wantErr:    true,
			errMsg:     "有效期不能为空",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			oldUpdatedAt := tt.shop.UpdatedAt
			err := tt.shop.UpdateValidUntil(tt.validUntil)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
				assert.True(t, tt.shop.UpdatedAt.After(oldUpdatedAt) || tt.shop.UpdatedAt.Equal(oldUpdatedAt))
				assert.Equal(t, tt.validUntil, tt.shop.ValidUntil)
			}
		})
	}
}

func TestShop_UpdatePassword(t *testing.T) {
	tests := []struct {
		name        string
		shop        *Shop
		newPassword string
		wantErr     bool
		errMsg      string
	}{
		{
			name:        "valid password",
			shop:        &Shop{OwnerPassword: "oldpass"},
			newPassword: "newpass123",
			wantErr:     false,
		},
		{
			name:        "empty password",
			shop:        &Shop{OwnerPassword: "oldpass"},
			newPassword: "",
			wantErr:     true,
			errMsg:      "新密码不能为空",
		},
		{
			name:        "password with spaces",
			shop:        &Shop{OwnerPassword: "oldpass"},
			newPassword: "new pass",
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			oldUpdatedAt := tt.shop.UpdatedAt
			err := tt.shop.UpdatePassword(tt.newPassword)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
				assert.True(t, tt.shop.UpdatedAt.After(oldUpdatedAt) || tt.shop.UpdatedAt.Equal(oldUpdatedAt))
				assert.Equal(t, tt.newPassword, tt.shop.OwnerPassword)
			}
		})
	}
}

func TestShop_DefaultOrderStatusFlow(t *testing.T) {
	shop, err := NewShop("测试店铺", "user", "pass", time.Now().AddDate(1, 0, 0))
	assert.NoError(t, err)

	assert.NotEmpty(t, shop.OrderStatusFlow.Statuses)
	assert.Greater(t, len(shop.OrderStatusFlow.Statuses), 0)

	// Verify the structure
	for _, status := range shop.OrderStatusFlow.Statuses {
		assert.NotEmpty(t, status.Label)
	}
}
