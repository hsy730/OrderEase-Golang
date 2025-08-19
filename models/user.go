package models

import (
	"strings"
	"time"

	"github.com/bwmarrin/snowflake"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type User struct {
	ID        snowflake.ID `gorm:"primarykey" json:"id"`
	Role      string       `gorm:"size:50;default:'user'" json:"role"` // 使用UserRole枚举值
	Password  string       `gorm:"size:255" json:"-"`
	Name      string       `json:"name"`
	Phone     string       `json:"phone"`
	Address   string       `json:"address"`
	Type      string       `json:"type"` // delivery:邮寄, pickup:自提, system:系统用户
	CreatedAt time.Time    `json:"created_at"`
	UpdatedAt time.Time    `json:"updated_at"`
	Orders    []Order      `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE;"`
}

const (
	UserTypeDelivery = "delivery"
	UserTypePickup   = "pickup"

	// 用户角色枚举
	UserRoleRegular = "user"   // 普通用户
	UserRoleSystem  = "system" // 系统用户
)

func (u *User) BeforeSave(tx *gorm.DB) error {
	if u.Password != "" && !strings.HasPrefix(u.Password, "$2a$") {
		hashed, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		u.Password = string(hashed)
	}
	return nil
}
