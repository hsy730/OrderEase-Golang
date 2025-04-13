package models

import (
	"time"

	"github.com/bwmarrin/snowflake"
)

type User struct {
	ID        snowflake.ID `gorm:"primarykey" json:"id"`
	ShopID    uint64       `gorm:"index;not null" json:"shop_id"` // 新增店铺ID字段
	Name      string       `json:"name"`
	Phone     string       `json:"phone"`
	Address   string       `json:"address"`
	Type      string       `json:"type"` // delivery: 邮寄, pickup: 自提
	CreatedAt time.Time    `json:"created_at"`
	UpdatedAt time.Time    `json:"updated_at"`
}

const (
	UserTypeDelivery = "delivery" // 邮寄
	UserTypePickup   = "pickup"   // 自提
)
