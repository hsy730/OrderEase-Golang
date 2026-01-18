package repositories

import (
	"errors"
	"orderease/models"
	"orderease/utils/log2"

	"gorm.io/gorm"
)

// AdminRepository 管理员数据访问层
type AdminRepository struct {
	DB *gorm.DB
}

// NewAdminRepository 创建AdminRepository实例
func NewAdminRepository(db *gorm.DB) *AdminRepository {
	return &AdminRepository{DB: db}
}

// GetAdminByUsername 根据用户名获取管理员
func (r *AdminRepository) GetAdminByUsername(username string) (*models.Admin, error) {
	var admin models.Admin
	err := r.DB.Where("username = ?", username).First(&admin).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("管理员不存在")
	}
	if err != nil {
		log2.Errorf("GetAdminByUsername failed: %v", err)
		return nil, errors.New("查询管理员失败")
	}
	return &admin, nil
}

// GetFirstAdmin 获取第一个管理员账户
func (r *AdminRepository) GetFirstAdmin() (*models.Admin, error) {
	var admin models.Admin
	err := r.DB.First(&admin).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("管理员账户不存在")
	}
	if err != nil {
		log2.Errorf("GetFirstAdmin failed: %v", err)
		return nil, errors.New("查询管理员失败")
	}
	return &admin, nil
}
