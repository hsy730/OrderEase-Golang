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
	ShopID      snowflake.ID `gorm:"column:shop_id;type:bigint unsigned;index;not null" json:"shop_id" binding:"required"` // 店铺ID
	Name        string       `gorm:"column:name" json:"name" binding:"required,min=1,max=200"`
	Description string       `gorm:"column:description" json:"description" binding:"max=5000"`
	Price       float64      `gorm:"column:price;type:double" json:"price" binding:"required,gt=0"`
	Stock       int          `gorm:"column:stock" json:"stock" binding:"required,min=0"`
	ImageURL    string       `gorm:"column:image_url" json:"image_url"`
	CreatedAt   time.Time    `gorm:"column:created_at" json:"created_at"`
	UpdatedAt   time.Time    `gorm:"column:updated_at" json:"updated_at"`
	Status      string       `gorm:"column:status" json:"status"`

	// 可选：添加参数类别关联，方便查询
	OptionCategories []ProductOptionCategory `gorm:"foreignKey:ProductID;constraint:OnDelete:CASCADE" json:"option_categories,omitempty"`
}

// ProductTag 商品和标签的多对多关系表
type ProductTag struct {
	ProductID snowflake.ID `gorm:"column:product_id;primaryKey" json:"product_id"`
	TagID     int          `gorm:"column:tag_id;primaryKey" json:"tag_id"`
	ShopID    snowflake.ID `gorm:"column:shop_id;type:bigint unsigned;index;not null" json:"shop_id"` // 店铺ID
	CreatedAt time.Time    `gorm:"column:created_at" json:"created_at"`
	UpdatedAt time.Time    `gorm:"column:updated_at" json:"updated_at"`
}
