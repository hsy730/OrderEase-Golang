package models

import (
	"time"
)

// Tag 商品标签
type Tag struct {
	// 商品标签通常只有几十个，数据量小，不使用雪花ID
	ID          int       `gorm:"primarykey" json:"id"`
	ShopID      uint64    `gorm:"index;not null" json:"shop_id"` // 新增店铺ID
	Name        string    `gorm:"size:50;not null;uniqueIndex" json:"name"`
	Description string    `gorm:"size:200" json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Products    []Product `gorm:"many2many:product_tags;" json:"products"`
}
