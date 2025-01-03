package models

import "time"

type Order struct {
	ID         uint        `gorm:"primarykey" json:"id"`
	UserID     uint        `json:"user_id"`
	User       User        `gorm:"foreignKey:UserID" json:"user"`
	TotalPrice Price       `json:"total_price"`
	Status     string      `json:"status"`
	Remark     string      `json:"remark"`
	Items      []OrderItem `gorm:"foreignKey:OrderID" json:"items"`
	CreatedAt  time.Time   `json:"created_at"`
	UpdatedAt  time.Time   `json:"updated_at"`
}
