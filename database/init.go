package database

import (
	"orderease/models"
	"orderease/utils/log2"

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

		log2.Infof("已创建默认管理员账户，请及时修改密码")
	}

	return nil
}

func Init() (*gorm.DB, error) {
	// 获取数据库连接
	db := GetDB()

	// 数据库迁移
	tables := []interface{}{
		&models.Product{},
		&models.Order{},
		&models.OrderItem{},
		&models.User{},
		&models.Tag{},
		&models.ProductTag{},
		&models.Shop{},

		&models.OrderStatusLog{},   // 不需要迁移数据
		&models.Admin{},            // 不需要迁移数据
		&models.BlacklistedToken{}, // 不需要迁移数据
	}
	// 自动迁移数据库表结构
	for _, table := range tables {
		if err := db.AutoMigrate(table); err != nil {
			log2.Logger.Fatalf("迁移表 %T 失败: %v", table, err)
		}
		log2.Logger.Printf("表 %T 迁移成功", table)
	}

	log2.Logger.Println("所有数据库表迁移完成")

	// 初始化管理员账户
	if err := InitAdminAccount(db); err != nil {
		return nil, fmt.Errorf("初始化管理员账户失败: %v", err)
	}

	return db, nil
}
