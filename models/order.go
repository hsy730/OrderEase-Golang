package models

import (
	"time"

	"github.com/bwmarrin/snowflake"
)

type Order struct {
	ID         snowflake.ID `gorm:"primarykey" json:"id"`
	UserID     snowflake.ID `json:"user_id"`
	User       User         `gorm:"foreignKey:UserID" json:"user"`
	TotalPrice Price        `json:"total_price"`
	Status     string       `json:"status"`
	Remark     string       `json:"remark"`
	Items      []OrderItem  `gorm:"foreignKey:OrderID" json:"items"`
	CreatedAt  time.Time    `json:"created_at"`
	UpdatedAt  time.Time    `json:"updated_at"`
}

type OrderItem struct {
	ID        snowflake.ID `gorm:"primarykey" json:"id"`
	OrderID   snowflake.ID `json:"order_id"`
	ProductID snowflake.ID `json:"product_id"`
	Quantity  int          `json:"quantity"`
	Price     Price        `json:"price"`
	Product   Product      `gorm:"foreignKey:ProductID" json:"product"`
}

const (
	OrderStatusPending  = "pending"   // 待处理
	OrderStatusAccepted = "accepted"  // 已接单
	OrderStatusRejected = "rejected"  // 已拒绝
	OrderStatusShipped  = "shipped"   // 已发货
	OrderStatusComplete = "completed" // 已完成
	OrderStatusCanceled = "canceled"  // 已取消
)

var OrderStatusTransitions = map[string]string{
	OrderStatusPending:  OrderStatusAccepted, // 待处理 -> 已接单
	OrderStatusAccepted: OrderStatusShipped,  // 已接单 -> 已发货
	OrderStatusShipped:  OrderStatusComplete, // 已发货 -> 已完成
	// 特殊状态转换
	OrderStatusRejected: OrderStatusRejected, // 已拒绝状态不变
	OrderStatusComplete: OrderStatusComplete, // 已完成状态不变
	OrderStatusCanceled: OrderStatusCanceled, // 已取消状态不变
}

// 订单状态变更日志
type OrderStatusLog struct {
	ID          snowflake.ID `gorm:"primarykey" json:"id"`
	OrderID     snowflake.ID `json:"order_id"`
	OldStatus   string       `json:"old_status"`
	NewStatus   string       `json:"new_status"`
	ChangedTime time.Time    `json:"changed_time"`
}
