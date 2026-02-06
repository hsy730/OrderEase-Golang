package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

// UserThirdpartyBinding 用户第三方平台绑定
type UserThirdpartyBinding struct {
	ID             uint      `gorm:"primarykey" json:"id"`
	UserID         uint64    `gorm:"column:user_id;not null;index:idx_user_id;comment:用户ID" json:"user_id"`
	Provider       string    `gorm:"column:provider;type:varchar(20);not null;index:idx_provider;comment:平台类型" json:"provider"`
	ProviderUserID string    `gorm:"column:provider_user_id;type:varchar(128);not null;comment:第三方平台用户ID" json:"provider_user_id"`
	UnionID        string    `gorm:"column:union_id;type:varchar(128);index:idx_union_id;comment:开放平台统一ID" json:"union_id,omitempty"`

	// 第三方用户信息
	Nickname  string `gorm:"column:nickname;type:varchar(100);comment:第三方昵称" json:"nickname,omitempty"`
	AvatarURL string `gorm:"column:avatar_url;type:varchar(500);comment:第三方头像URL" json:"avatar_url,omitempty"`
	Gender    int    `gorm:"column:gender;type:tinyint;default:0;comment:性别" json:"gender,omitempty"`
	Country   string `gorm:"column:country;type:varchar(50);comment:国家" json:"country,omitempty"`
	Province  string `gorm:"column:province;type:varchar(50);comment:省份" json:"province,omitempty"`
	City      string `gorm:"column:city;type:varchar(50);comment:城市" json:"city,omitempty"`

	// 扩展字段
	Metadata Metadata `gorm:"column:metadata;type:json;comment:平台特有数据" json:"metadata,omitempty"`

	// 状态字段
	IsActive     bool       `gorm:"column:is_active;type:tinyint(1);default:1;index:idx_is_active;comment:是否激活" json:"is_active"`
	LastLoginAt  *time.Time `gorm:"column:last_login_at;comment:最后登录时间" json:"last_login_at,omitempty"`

	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime;comment:绑定时间" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime;comment:更新时间" json:"updated_at"`
}

// TableName 指定表名
func (UserThirdpartyBinding) TableName() string {
	return "user_thirdparty_bindings"
}

// Metadata JSON 类型，存储平台特有数据
type Metadata map[string]interface{}

// Scan 实现 sql.Scanner 接口（从数据库读取）
func (m *Metadata) Scan(value interface{}) error {
	if value == nil {
		*m = make(Metadata)
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to unmarshal Metadata value: %v", value)
	}
	return json.Unmarshal(bytes, m)
}

// Value 实现 driver.Valuer 接口（写入数据库）
func (m Metadata) Value() (driver.Value, error) {
	if len(m) == 0 {
		return nil, nil
	}
	return json.Marshal(m)
}

// GetAccessToken 从 metadata 获取 access_token
func (m Metadata) GetAccessToken() string {
	if v, ok := m["access_token"].(string); ok {
		return v
	}
	return ""
}

// GetRefreshToken 从 metadata 获取 refresh_token
func (m Metadata) GetRefreshToken() string {
	if v, ok := m["refresh_token"].(string); ok {
		return v
	}
	return ""
}

// SetAccessToken 设置 access_token 到 metadata
func (m Metadata) SetAccessToken(token string) {
	if m == nil {
		m = make(Metadata)
	}
	m["access_token"] = token
}

// SetRefreshToken 设置 refresh_token 到 metadata
func (m Metadata) SetRefreshToken(token string) {
	if m == nil {
		m = make(Metadata)
	}
	m["refresh_token"] = token
}

// IsActiveBinding 检查绑定是否激活
func (b *UserThirdpartyBinding) IsActiveBinding() bool {
	return b.IsActive
}

// UpdateLastLogin 更新最后登录时间
func (b *UserThirdpartyBinding) UpdateLastLogin() {
	now := time.Now()
	b.LastLoginAt = &now
}
