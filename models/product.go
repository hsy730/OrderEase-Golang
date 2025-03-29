package models

import (
	"time"

	"github.com/bwmarrin/snowflake"
)

// 商品状态常量
const (
	ProductStatusPending = "pending" // 待上架
	ProductStatusOnline  = "online"  // 已上架
	ProductStatusOffline = "offline" // 已下架
)

type Product struct {
	ID          snowflake.ID `gorm:"primarykey" json:"id"`
	ShopID      uint64       `gorm:"index;not null" json:"shop_id"` // 新增店铺ID
	Name        string       `json:"name"`
	Description string       `json:"description"`
	Price       float64      `json:"price"`
	Stock       int          `json:"stock"`
	ImageURL    string       `json:"image_url"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
	Status      string       `json:"status"`
}

// ProductTag 商品和标签的多对多关系表
type ProductTag struct {
	ProductID snowflake.ID `gorm:"primaryKey" json:"product_id"`
	TagID     int          `gorm:"primaryKey" json:"tag_id"`
	CreatedAt time.Time    `json:"created_at"`
	UpdatedAt time.Time    `json:"updated_at"`
	// 移除 ShopID（可通过关联查询获取）
	// 通过以下方式保证数据一致性：
	// 1. 添加数据库外键约束
	// 2. 在应用层验证 Product.ShopID == Tag.ShopID
}
