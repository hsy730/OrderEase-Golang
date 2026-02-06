package models

import "time"

// OAuthState OAuth State 数据模型
// 用于防止 CSRF 攻击，每次授权都需要生成唯一的 state
type OAuthState struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	State     string    `gorm:"type:varchar(64);uniqueIndex;not null;comment:OAuth State 参数" json:"state"`
	Provider  string    `gorm:"type:varchar(20);not null;comment:平台类型" json:"provider"`
	ExpiresAt time.Time `gorm:"not null;comment:过期时间" json:"expires_at"`
	CreatedAt time.Time `gorm:"autoCreateTime;comment:创建时间" json:"created_at"`
}

// TableName 指定表名
func (OAuthState) TableName() string {
	return "oauth_states"
}

// IsExpired 检查是否已过期
func (s *OAuthState) IsExpired() bool {
	return time.Now().After(s.ExpiresAt)
}
