package repositories

import (
	"fmt"
	"time"

	"orderease/contexts/thirdparty/domain/oauth"
	"orderease/models"

	"gorm.io/gorm"
)

// UserThirdpartyBindingRepository 用户第三方平台绑定仓储
type UserThirdpartyBindingRepository struct {
	db *gorm.DB
}

// NewUserThirdpartyBindingRepository 创建用户第三方平台绑定仓储
func NewUserThirdpartyBindingRepository(db *gorm.DB) *UserThirdpartyBindingRepository {
	return &UserThirdpartyBindingRepository{db: db}
}

// FindByProviderAndUserID 通过平台和第三方用户ID查找绑定
func (r *UserThirdpartyBindingRepository) FindByProviderAndUserID(provider oauth.Provider, providerUserID string) (*models.UserThirdpartyBinding, error) {
	var binding models.UserThirdpartyBinding
	err := r.db.Where("provider = ? AND provider_user_id = ?", provider.String(), providerUserID).
		Where("is_active = ?", true).
		First(&binding).Error
	if err != nil {
		return nil, err
	}
	return &binding, nil
}

// FindByUserID 通过系统用户ID查找所有绑定
func (r *UserThirdpartyBindingRepository) FindByUserID(userID uint64) ([]models.UserThirdpartyBinding, error) {
	var bindings []models.UserThirdpartyBinding
	err := r.db.Where("user_id = ? AND is_active = ?", userID, true).
		Find(&bindings).Error
	return bindings, err
}

// FindByUserIDAndProvider 通过系统用户ID和平台查找绑定
func (r *UserThirdpartyBindingRepository) FindByUserIDAndProvider(userID uint64, provider oauth.Provider) (*models.UserThirdpartyBinding, error) {
	var binding models.UserThirdpartyBinding
	err := r.db.Where("user_id = ? AND provider = ? AND is_active = ?", userID, provider.String(), true).
		First(&binding).Error
	if err != nil {
		return nil, err
	}
	return &binding, nil
}

// Create 创建绑定
func (r *UserThirdpartyBindingRepository) Create(binding *models.UserThirdpartyBinding) error {
	if err := r.db.Create(binding).Error; err != nil {
		return fmt.Errorf("create binding failed: %w", err)
	}
	return nil
}

// Update 更新绑定
func (r *UserThirdpartyBindingRepository) Update(binding *models.UserThirdpartyBinding) error {
	if err := r.db.Save(binding).Error; err != nil {
		return fmt.Errorf("update binding failed: %w", err)
	}
	return nil
}

// UpdateLastLogin 更新最后登录时间
func (r *UserThirdpartyBindingRepository) UpdateLastLogin(id uint) error {
	now := time.Now()
	return r.db.Model(&models.UserThirdpartyBinding{}).
		Where("id = ?", id).
		Update("last_login_at", now).Error
}

// Deactivate 解绑（软删除）
func (r *UserThirdpartyBindingRepository) Deactivate(id uint) error {
	return r.db.Model(&models.UserThirdpartyBinding{}).
		Where("id = ?", id).
		Update("is_active", false).Error
}

// Delete 硬删除绑定
func (r *UserThirdpartyBindingRepository) Delete(id uint) error {
	return r.db.Delete(&models.UserThirdpartyBinding{}, id).Error
}

// ListActive 获取所有激活的绑定（用于管理）
func (r *UserThirdpartyBindingRepository) ListActive(limit, offset int) ([]models.UserThirdpartyBinding, error) {
	var bindings []models.UserThirdpartyBinding
	query := r.db.Where("is_active = ?", true)
	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}
	err := query.Find(&bindings).Error
	return bindings, err
}

// CountByProvider 统计各平台的绑定数量
func (r *UserThirdpartyBindingRepository) CountByProvider() (map[string]int64, error) {
	type Result struct {
		Provider string
		Count    int64
	}
	var results []Result
	err := r.db.Model(&models.UserThirdpartyBinding{}).
		Select("provider, COUNT(*) as count").
		Where("is_active = ?", true).
		Group("provider").
		Scan(&results).Error
	if err != nil {
		return nil, err
	}

	counts := make(map[string]int64)
	for _, r := range results {
		counts[r.Provider] = r.Count
	}
	return counts, nil
}
