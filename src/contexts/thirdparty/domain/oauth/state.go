package oauth

import "time"

// State OAuth State 管理接口
// State 用于防止 CSRF 攻击，每次授权都需要生成唯一的 state
type State interface {
	// Generate 生成新的 state
	Generate() string

	// Validate 验证 state 是否有效
	Validate(state string) bool

	// Delete 删除已使用的 state
	Delete(state string) error

	// Cleanup 清理过期的 state
	Cleanup() error
}

// StateEntity State 实体
type StateEntity struct {
	State     string    // State 字符串
	Provider  Provider  // 平台提供者
	ExpiresAt time.Time // 过期时间
	CreatedAt time.Time // 创建时间
}

// IsExpired 检查 state 是否已过期
func (s *StateEntity) IsExpired() bool {
	return time.Now().After(s.ExpiresAt)
}
