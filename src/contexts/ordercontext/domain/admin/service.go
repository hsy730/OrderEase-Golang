package admin

import (
	"fmt"

	"gorm.io/gorm"
	"orderease/models"
)

type Service struct {
	db *gorm.DB
}

func NewService(db *gorm.DB) *Service {
	return &Service{db: db}
}

// UpdatePassword 更新管理员密码
func (s *Service) UpdatePassword(username, oldPassword, newPassword string) error {
	var admin models.Admin
	if err := s.db.Where("username = ?", username).First(&admin).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("管理员账户不存在")
		}
		return fmt.Errorf("查询管理员失败: %w", err)
	}

	if !admin.CheckPassword(oldPassword) {
		return fmt.Errorf("旧密码错误")
	}

	admin.Password = newPassword
	if err := admin.HashPassword(); err != nil {
		return fmt.Errorf("密码加密失败: %w", err)
	}

	if err := s.db.Save(&admin).Error; err != nil {
		return fmt.Errorf("保存新密码失败: %w", err)
	}

	return nil
}
