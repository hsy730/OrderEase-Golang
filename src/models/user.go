package models

import (
	"time"

	"github.com/bwmarrin/snowflake"
)

type User struct {
	ID        snowflake.ID `gorm:"column:id;primarykey" json:"id"`
	Name      string       `gorm:"column:name;size:100;uniqueIndex" json:"name"`
	Role      string       `gorm:"column:role;size:50;default:'user'" json:"role"` // 使用UserRole枚举值
	Password  string       `gorm:"column:password;size:255" json:"-"`
	Phone     string       `gorm:"column:phone" json:"phone"`
	Address   string       `gorm:"column:address" json:"address"`
	Type      string       `gorm:"column:type" json:"type"` // delivery:邮寄, pickup:自提, system:系统用户
	CreatedAt time.Time    `gorm:"column:created_at" json:"created_at"`
	UpdatedAt time.Time    `gorm:"column:updated_at" json:"updated_at"`
	Orders    []Order      `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE;"`
}

const (
	UserTypeDelivery = "delivery"
	UserTypePickup   = "pickup"

	// 用户角色枚举
	UserRolePrivate = "private_user" // 普通用户
	UserRolePublic  = "public_user"  // 公共用户
)
