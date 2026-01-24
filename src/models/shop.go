package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

// OrderStatusAction 订单状态动作
type OrderStatusAction struct {
	Name            string `json:"name" binding:"required"`
	NextStatus      int    `json:"nextStatus" binding:"required"`
	NextStatusLabel string `json:"nextStatusLabel" binding:"required"`
}

// OrderStatus 订单状态
type OrderStatus struct {
	Value   int                 `json:"value" binding:"required"`
	Label   string              `json:"label" binding:"required"`
	Type    string              `json:"type" binding:"required"`
	IsFinal bool                `json:"isFinal" binding:"required"`
	Actions []OrderStatusAction `json:"actions" binding:"required"`
}

// OrderStatusFlow 订单流转状态配置
type OrderStatusFlow struct {
	Statuses []OrderStatus `json:"statuses"`
}

// Value 实现 driver.Valuer 接口，将 OrderStatusFlow 转换为 JSON 字符串存入数据库
func (osf OrderStatusFlow) Value() (driver.Value, error) {
	return json.Marshal(osf)
}

// Scan 实现 sql.Scanner 接口，将数据库中的 JSON 字符串转换为 OrderStatusFlow
func (osf *OrderStatusFlow) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(b, &osf)
}

type Shop struct {
	ID            uint64 `gorm:"column:id;primarykey" json:"id"`
	Name          string `gorm:"column:name;size:100;not null" json:"name"`                                //店名
	OwnerUsername string `gorm:"column:owner_username;size:50;not null;uniqueIndex" json:"owner_username"` // 店主登录用户
	OwnerPassword string `gorm:"column:owner_password;size:255;not null" json:"-"`                         // 店主登录密码

	ContactPhone string `gorm:"column:contact_phone;size:20" json:"contact_phone"`
	ContactEmail string `gorm:"column:contact_email;size:100" json:"contact_email"`
	Address      string `gorm:"column:address;size:100" json:"address"`
	ImageURL     string `gorm:"column:image_url;size:255" json:"image_url"` // 店铺图片URL

	Description     string          `gorm:"column:description;type:text" json:"description"` // 店铺描述
	CreatedAt       time.Time       `gorm:"column:created_at" json:"created_at"`
	UpdatedAt       time.Time       `gorm:"column:updated_at" json:"updated_at"`
	ValidUntil      time.Time       `gorm:"column:valid_until;index" json:"valid_until"`                 // 有效期
	Settings        json.RawMessage `gorm:"column:settings;type:json" json:"settings"`                   // 店铺设置
	OrderStatusFlow OrderStatusFlow `gorm:"column:order_status_flow;type:json" json:"order_status_flow"` // 订单流转状态配置
	Products        []Product       `gorm:"foreignKey:ShopID" json:"products"`
	Tags            []Tag           `gorm:"foreignKey:ShopID" json:"tags"`
}

// 业务方法已迁移到 domain/shop/shop.go
// CheckPassword 和 IsExpired 现在是领域实体的方法

// BeforeSave 钩子已移除 - 密码哈希现在在 handler 和 domain 层处理

// RemainingDays 方法已移除（未被使用）
