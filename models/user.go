package models

import "time"

type User struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	Name      string    `json:"name"`
	Phone     string    `json:"phone"`
	Address   string    `json:"address"`
	Type      string    `json:"type"` // delivery: 邮寄, pickup: 自提
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

const (
	UserTypeDelivery = "delivery" // 邮寄
	UserTypePickup   = "pickup"   // 自提
)
