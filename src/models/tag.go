package models

import (
	"time"

	"github.com/bwmarrin/snowflake"
)

// Tag 商品标签
type Tag struct {
	// 商品标签通常只有几十个，数据量小，不使用雪花ID
	ID          int          `gorm:"column:id;primarykey" json:"id"`
	ShopID      snowflake.ID `gorm:"column:shop_id;type:bigint unsigned;index;not null" json:"shop_id"` // 店铺ID
	Name        string    `gorm:"column:name;size:50;not null;uniqueIndex" json:"name"`
	Description string    `gorm:"column:description;size:200" json:"description"`
	CreatedAt   time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at" json:"updated_at"`
	Products    []Product `gorm:"many2many:product_tags;" json:"products"`
}
