package repositories

import (
	"errors"
	"fmt"
	"time"

	"orderease/contexts/thirdparty/domain/oauth"
	"orderease/models"
	"orderease/utils"

	"gorm.io/gorm"
)

// OAuthStateRepository OAuth State 仓储
type OAuthStateRepository struct {
	db *gorm.DB
}

// NewOAuthStateRepository 创建 OAuth State 仓储
func NewOAuthStateRepository(db *gorm.DB) *OAuthStateRepository {
	return &OAuthStateRepository{db: db}
}

// Save 保存 state
// ttl: 过期时间（秒）
func (r *OAuthStateRepository) Save(state string, provider oauth.Provider, ttl int64) error {
	now := time.Now()
	oauthState := &models.OAuthState{
		State:     state,
		Provider:  provider.String(),
		ExpiresAt: now.Add(time.Duration(ttl) * time.Second),
	}

	if err := r.db.Create(oauthState).Error; err != nil {
		return fmt.Errorf("save oauth state failed: %w", err)
	}

	return nil
}

// Validate 验证 state 是否有效
func (r *OAuthStateRepository) Validate(state string) error {
	var oauthState models.OAuthState
	err := r.db.Where("state = ?", state).First(&oauthState).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("invalid state: not found")
		}
		return fmt.Errorf("validate state failed: %w", err)
	}

	if oauthState.IsExpired() {
		return fmt.Errorf("invalid state: expired")
	}

	return nil
}

// VerifyAndDelete 验证并删除 state（原子操作）
func (r *OAuthStateRepository) VerifyAndDelete(state string) error {
	var oauthState models.OAuthState
	err := r.db.Where("state = ?", state).First(&oauthState).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("invalid state: not found")
		}
		return fmt.Errorf("verify state failed: %w", err)
	}

	if oauthState.IsExpired() {
		return fmt.Errorf("invalid state: expired")
	}

	// 删除已使用的 state
	if err := r.db.Delete(&oauthState).Error; err != nil {
		return fmt.Errorf("delete state failed: %w", err)
	}

	return nil
}

// Delete 删除 state
func (r *OAuthStateRepository) Delete(state string) error {
	result := r.db.Where("state = ?", state).Delete(&models.OAuthState{})
	if result.Error != nil {
		return fmt.Errorf("delete state failed: %w", result.Error)
	}
	return nil
}

// Cleanup 清理过期的 state
func (r *OAuthStateRepository) Cleanup() error {
	result := r.db.Where("expires_at < ?", time.Now()).Delete(&models.OAuthState{})
	if result.Error != nil {
		return fmt.Errorf("cleanup expired states failed: %w", result.Error)
	}
	return nil
}

// Generate 生成新的 state
func (r *OAuthStateRepository) Generate() string {
	return utils.GenerateRandomString(32)
}
