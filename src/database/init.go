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
		// 基础表（无依赖）
		&models.User{},             // User 必须在 Order 之前，因为 Order 有外键引用 User
		&models.Shop{},             // Shop 必须在 Product 之前，因为 Product 有外键引用 Shop
		&models.OAuthState{},       // OAuth State 管理表
		&models.Admin{},            // 不需要迁移数据
		&models.BlacklistedToken{}, // 不需要迁移数据

		// 依赖基础表的表
		&models.Product{},               // 依赖 Shop
		&models.Tag{},                   // 依赖 Shop
		&models.TempToken{},             // 依赖 Shop
		&models.UserThirdpartyBinding{}, // 依赖 User
		&models.Order{},                 // 依赖 User 和 Shop

		// 依赖上述表的表
		&models.ProductOptionCategory{}, // 依赖 Product
		&models.ProductOption{},         // 依赖 ProductOptionCategory
		&models.ProductTag{},            // 依赖 Product 和 Tag
		&models.OrderItem{},             // 依赖 Order 和 Product
		&models.OrderItemOption{},       // 依赖 OrderItem
		&models.OrderStatusLog{},        // 依赖 Order
	}
	// 自动迁移数据库表结构
	for _, table := range tables {
		if err := db.AutoMigrate(table); err != nil {
			log2.Fatalf("迁移表 %T 失败: %v", table, err)
		}
		log2.Infof("表 %T 迁移成功", table)
	}

	log2.Infof("所有数据库表迁移完成")

	// 初始化管理员账户
	if err := InitAdminAccount(db); err != nil {
		return nil, fmt.Errorf("初始化管理员账户失败: %v", err)
	}

	return db, nil
}
