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
	
	// 可选：添加参数类别关联，方便查询
	OptionCategories []ProductOptionCategory `gorm:"foreignKey:ProductID;constraint:OnDelete:CASCADE" json:"option_categories,omitempty"`
}

// ProductTag 商品和标签的多对多关系表
type ProductTag struct {
	ProductID snowflake.ID `gorm:"primaryKey" json:"product_id"`
	TagID     int          `gorm:"primaryKey" json:"tag_id"`
	ShopID    uint64       `gorm:"index;not null" json:"shop_id"` // 恢复 ShopID 字段
	CreatedAt time.Time    `json:"created_at"`
	UpdatedAt time.Time    `json:"updated_at"`
}
