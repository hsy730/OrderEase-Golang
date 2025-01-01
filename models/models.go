package models

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"
)

type Product struct {
	ID          uint      `gorm:"primarykey" json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Price       float64   `json:"price"`
	Stock       int       `json:"stock"`
	ImageURL    string    `json:"image_url"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Price float64

func (p *Price) UnmarshalJSON(data []byte) error {
	var value interface{}
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}

	switch v := value.(type) {
	case float64:
		*p = Price(v)
	case string:
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			*p = Price(f)
		} else {
			return fmt.Errorf("invalid price format: %s", v)
		}
	default:
		return fmt.Errorf("invalid price type: %T", value)
	}

	return nil
}

type Order struct {
	ID         uint        `gorm:"primarykey" json:"id"`
	UserID     uint        `json:"user_id"`
	TotalPrice Price       `json:"total_price"`
	Status     string      `json:"status"`
	Remark     string      `json:"remark"`
	Items      []OrderItem `gorm:"foreignKey:OrderID" json:"items"`
	CreatedAt  time.Time   `json:"created_at"`
	UpdatedAt  time.Time   `json:"updated_at"`
}

type OrderItem struct {
	ID        uint    `gorm:"primarykey" json:"id"`
	OrderID   uint    `json:"order_id"`
	ProductID uint    `json:"product_id"`
	Quantity  int     `json:"quantity"`
	Price     Price   `json:"price"`
	Product   Product `gorm:"foreignKey:ProductID" json:"product"`
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
	ID          uint      `gorm:"primarykey" json:"id"`
	OrderID     uint      `json:"order_id"`
	OldStatus   string    `json:"old_status"`
	NewStatus   string    `json:"new_status"`
	ChangedTime time.Time `json:"changed_time"`
}
