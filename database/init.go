package database

import (
	"orderease/models"
	"orderease/utils"

	"fmt"

	"gorm.io/gorm"
)

// 初始化管理员账户
func InitAdminAccount(db *gorm.DB) error {
	var count int64
	db.Model(&models.Admin{}).Count(&count)

	// 如果没有管理员账户，创建默认账户
	if count == 0 {
		admin := models.Admin{
			Username: "admin",
			Password: "Admin@123456", // 初始密码
		}

		if err := admin.HashPassword(); err != nil {
			return err
		}

		if err := db.Create(&admin).Error; err != nil {
			return err
		}

		utils.Logger.Println("已创建默认管理员账户，请及时修改密码")
	}

	return nil
}

func Init() (*gorm.DB, error) {
	// 获取数据库连接
	db := GetDB()

	// 自动迁移数据库表结构
	if err := db.AutoMigrate(
		&models.Admin{},
		// ... 其他模型
	); err != nil {
		return nil, fmt.Errorf("数据库迁移失败: %v", err)
	}

	// 初始化管理员账户
	if err := InitAdminAccount(db); err != nil {
		return nil, fmt.Errorf("初始化管理员账户失败: %v", err)
	}

	return db, nil
}
