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
	ID          snowflake.ID `gorm:"column:id;primarykey;type:bigint unsigned" json:"id"`
	ShopID      uint64       `gorm:"column:shop_id;index;not null" json:"shop_id"` // 新增店铺ID
	Name        string       `gorm:"column:name" json:"name"`
	Description string       `gorm:"column:description" json:"description"`
	Price       float64      `gorm:"column:price;type:double" json:"price"`
	Stock       int          `gorm:"column:stock" json:"stock"`
	ImageURL    string       `gorm:"column:image_url" json:"image_url"`
	CreatedAt   time.Time    `gorm:"column:created_at" json:"created_at"`
	UpdatedAt   time.Time    `gorm:"column:updated_at" json:"updated_at"`
	Status      string       `gorm:"column:status" json:"status"`

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
