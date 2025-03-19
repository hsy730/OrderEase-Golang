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
	ShopID      int          `gorm:"index;not null" json:"shop_id"` // 新增店铺ID
	Name        string       `json:"name"`
	Description string       `json:"description"`
	Price       float64      `json:"price"`
	Stock       int          `json:"stock"`
	ImageURL    string       `json:"image_url"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
	Status      string       `json:"status"`
}
