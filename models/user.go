package models

import (
	"time"

	"github.com/bwmarrin/snowflake"
)

type User struct {
	ID        snowflake.ID `gorm:"primarykey" json:"id"`
	IsSystem  bool         `gorm:"default:false" json:"is_system"` // true表示系统自动生成的公共用户
	Password  string       `gorm:"size:255" json:"-"`
	Name      string       `json:"name"`
	Phone     string       `json:"phone"`
	Address   string       `json:"address"`
	Type      string       `json:"type"` // delivery:邮寄, pickup:自提, system:系统用户
	CreatedAt time.Time    `json:"created_at"`
	UpdatedAt time.Time    `json:"updated_at"`
}

const (
	UserTypeDelivery = "delivery" // 邮寄
	UserTypePickup   = "pickup"   // 自提
)
