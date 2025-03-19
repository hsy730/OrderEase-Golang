package models

import (
	"time"

	"github.com/bwmarrin/snowflake"
)

// Tag 商品标签
type Tag struct {
	// 商品标签通常只有几十个，数据量小，不使用雪花ID
	ID          int       `gorm:"primarykey" json:"id"`
	ShopID      int       `gorm:"index;not null" json:"shop_id"` // 新增店铺ID
	Name        string    `gorm:"size:50;not null;uniqueIndex" json:"name"`
	Description string    `gorm:"size:200" json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Products    []Product `gorm:"many2many:product_tags;" json:"products"`
}

// ProductTag 商品和标签的多对多关系表
type ProductTag struct {
	ProductID snowflake.ID `gorm:"primaryKey" json:"product_id"`
	TagID     int          `gorm:"primaryKey" json:"tag_id"`
	ShopID    int          `gorm:"index;not null" json:"shop_id"` // 新增店铺ID
	CreatedAt time.Time    `json:"created_at"`
	UpdatedAt time.Time    `json:"updated_at"`
}

// 新增店铺模型
type Shop struct {
	ID          int       `gorm:"primarykey" json:"id"`
	Name        string    `gorm:"size:100;not null" json:"name"`
	Description string    `gorm:"type:text" json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Products    []Product `gorm:"foreignKey:ShopID" json:"products"`
	Tags        []Tag     `gorm:"foreignKey:ShopID" json:"tags"`
}
