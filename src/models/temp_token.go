package models

import (
	"time"

	"github.com/bwmarrin/snowflake"
)

// TempToken 临时令牌模型
type TempToken struct {
	ID        snowflake.ID `gorm:"primarykey;autoIncrement:false" json:"id"`
	ShopID    snowflake.ID `gorm:"index;type:bigint unsigned" json:"shop_id"`
	UserID    uint64       `gorm:"index;not null" json:"user_id"` // 关联系统用户
	Token     string       `gorm:"size:6;not null" json:"token"`  // 6位数字令牌
	ExpiresAt time.Time    `gorm:"index;not null" json:"expires_at"`
	CreatedAt time.Time    `gorm:"column:created_at" json:"created_at"`
	UpdatedAt time.Time    `gorm:"column:updated_at" json:"updated_at"`
}

// TableName 指定表名
func (TempToken) TableName() string {
	return "temp_tokens"
}
