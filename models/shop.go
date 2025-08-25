package models

import (
	"strings"
	"time"

	"gorm.io/datatypes" // 新增导入
	"gorm.io/gorm"

	"golang.org/x/crypto/bcrypt"
)

type Shop struct {
	ID            uint64 `gorm:"primarykey" json:"id"`
	Name          string `gorm:"size:100;not null" json:"name"`                      //店名
	OwnerUsername string `gorm:"size:50;not null;uniqueIndex" json:"owner_username"` // 店主登录用户
	OwnerPassword string `gorm:"size:255;not null" json:"-"`                         // 店主登录密码

	ContactPhone string `gorm:"size:20" json:"contact_phone"`
	ContactEmail string `gorm:"size:100" json:"contact_email"`
	Address      string `gorm:"size:100" json:"address"`
	ImageURL     string `gorm:"size:255" json:"image_url"`                           // 店铺图片URL

	Description string    `gorm:"type:text" json:"description"` // 店铺描述
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	ValidUntil  time.Time `gorm:"index" json:"valid_until"` // 有效期
	// 假设使用 gorm.io/datatypes 包中的 JSON 类型
	Settings datatypes.JSON `gorm:"type:json" json:"settings"` // 店铺设置
	Products []Product      `gorm:"foreignKey:ShopID" json:"products"`
	Tags     []Tag          `gorm:"foreignKey:ShopID" json:"tags"`
}

func (s *Shop) CheckPassword(password string) error {
	return bcrypt.CompareHashAndPassword([]byte(s.OwnerPassword), []byte(password))
}

func (s *Shop) HashPassword() error {
	hashed, err := bcrypt.GenerateFromPassword([]byte(s.OwnerPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	s.OwnerPassword = string(hashed)
	return nil
}

// 在创建/更新钩子中添加
func (s *Shop) BeforeSave(tx *gorm.DB) error {
	if s.OwnerPassword != "" && !strings.HasPrefix(s.OwnerPassword, "$2a$") {
		return s.HashPassword()
	}
	return nil
}

// IsExpired 判断店铺是否到期
func (s *Shop) IsExpired() bool {
	now := time.Now().UTC()
	return s.ValidUntil.Before(now)
}

// RemainingDays 获取剩余有效天数（负数表示已过期）
func (s *Shop) RemainingDays() int {
	hours := time.Until(s.ValidUntil.UTC()).Hours()
	return int(hours / 24) // 向下取整
}
