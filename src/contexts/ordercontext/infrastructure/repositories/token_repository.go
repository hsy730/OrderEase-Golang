package repositories

import (
	"errors"
	"orderease/models"
	"orderease/utils/log2"

	"gorm.io/gorm"
)

// TokenRepository Token 数据访问层
type TokenRepository struct {
	DB *gorm.DB
}

// NewTokenRepository 创建 TokenRepository 实例
func NewTokenRepository(db *gorm.DB) *TokenRepository {
	return &TokenRepository{DB: db}
}

// CreateBlacklistedToken 将 token 加入黑名单
func (r *TokenRepository) CreateBlacklistedToken(token *models.BlacklistedToken) error {
	err := r.DB.Create(token).Error
	if err != nil {
		log2.Errorf("CreateBlacklistedToken failed: %v", err)
		return errors.New("添加 token 到黑名单失败")
	}
	return nil
}

// IsTokenBlacklisted 检查 token 是否在黑名单中
func (r *TokenRepository) IsTokenBlacklisted(token string) (bool, error) {
	var count int64
	err := r.DB.Model(&models.BlacklistedToken{}).
		Where("token = ? AND expired_at > NOW()", token).
		Count(&count).Error
	if err != nil {
		log2.Errorf("IsTokenBlacklisted failed: %v", err)
		return false, errors.New("检查 token 黑名单失败")
	}
	return count > 0, nil
}

// CleanExpiredTokens 清理过期的黑名单 token (定期任务调用)
func (r *TokenRepository) CleanExpiredTokens() error {
	err := r.DB.Where("expired_at <= NOW()").
		Delete(&models.BlacklistedToken{}).Error
	if err != nil {
		log2.Errorf("CleanExpiredTokens failed: %v", err)
		return errors.New("清理过期 token 失败")
	}
	return nil
}
