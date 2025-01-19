package models

import "time"

// Tag 商品标签
type Tag struct {
	ID          uint      `gorm:"primarykey" json:"id"`
	Name        string    `gorm:"size:50;not null;uniqueIndex" json:"name"`
	Description string    `gorm:"size:200" json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Products    []Product `gorm:"many2many:product_tags;" json:"products"`
}

// ProductTag 商品和标签的多对多关系表
type ProductTag struct {
	ProductID uint      `gorm:"primaryKey" json:"product_id"`
	TagID     uint      `gorm:"primaryKey" json:"tag_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
