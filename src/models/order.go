package models

import (
	"time"

	"github.com/bwmarrin/snowflake"
)

type Order struct {
	ID         snowflake.ID `gorm:"primarykey;column:id;type:bigint unsigned" json:"id,omitempty"`
	UserID     snowflake.ID `gorm:"column:user_id;index:idx_user_shop_status,priority:1;not null" json:"user_id"`
	ShopID     snowflake.ID `gorm:"column:shop_id;type:bigint unsigned;index:idx_shop;index:idx_user_shop_status,priority:2;index:idx_shop_status,priority:1;not null" json:"shop_id"`
	TotalPrice Price        `gorm:"column:total_price;type:double" json:"total_price"`
	Status     int          `gorm:"column:status;index:idx_user_shop_status,priority:3;index:idx_shop_status,priority:2" json:"status"`
	Remark     string       `gorm:"column:remark" json:"remark"`
	CreatedAt  time.Time    `gorm:"column:created_at;index" json:"created_at"`
	UpdatedAt  time.Time    `gorm:"column:updated_at" json:"updated_at"`
	Items      []OrderItem  `gorm:"foreignKey:OrderID;constraint:OnDelete:CASCADE" json:"items"`
	User       User         `gorm:"foreignKey:UserID" json:"user"`
}

type OrderItem struct {
	ID         snowflake.ID `gorm:"column:id;primarykey;type:bigint unsigned" json:"id,omitempty"`
	OrderID    snowflake.ID `gorm:"column:order_id;type:bigint unsigned" json:"order_id"`
	ProductID  snowflake.ID `gorm:"column:product_id" json:"product_id"`
	Quantity   int          `gorm:"column:quantity" json:"quantity"`
	Price      Price        `gorm:"column:price;type:double" json:"price"`
	TotalPrice Price        `gorm:"column:total_price;type:double" json:"total_price"`
	// 添加商品快照字段
	ProductName        string `gorm:"column:product_name;size:255" json:"product_name"`           // 商品名称
	ProductDescription string `gorm:"column:product_description" json:"product_description"`      // 商品描述
	ProductImageURL    string `gorm:"column:product_image_url;size:255" json:"product_image_url"` // 商品图片URL
	// 删除Product关联字段，避免混淆和不必要的关联查询

	// 添加参数选项关联
	Options []OrderItemOption `gorm:"foreignKey:OrderItemID;constraint:OnDelete:CASCADE" json:"options"`
}

// OrderItemOption 订单项选择的商品参数选项
type OrderItemOption struct {
	ID              snowflake.ID `gorm:"column:id;primarykey;type:bigint unsigned" json:"id"`
	OrderItemID     snowflake.ID `gorm:"column:order_item_id;index;not null;type:bigint unsigned" json:"order_item_id"`
	CategoryID      snowflake.ID `gorm:"column:category_id;index;not null;type:bigint unsigned" json:"category_id"` // 类别ID快照
	OptionID        snowflake.ID `gorm:"column:option_id" json:"option_id"`
	OptionName      string       `gorm:"column:option_name;size:100" json:"option_name"`     // 选项名称快照
	CategoryName    string       `gorm:"column:category_name;size:100" json:"category_name"` // 类别名称快照
	PriceAdjustment float64      `gorm:"column:price_adjustment" json:"price_adjustment"`    // 价格调整快照
	CreatedAt       time.Time    `gorm:"column:created_at" json:"created_at"`
	UpdatedAt       time.Time    `gorm:"column:updated_at" json:"updated_at"`
}

const (
	OrderStatusPending  = 1  // 待处理
	OrderStatusAccepted = 2  // 已接单
	OrderStatusRejected = 3  // 已拒绝
	OrderStatusShipped  = 4  // 已发货
	OrderStatusComplete = 10 // 已完成，保留冗余的状态码
	OrderStatusCanceled = -1 // 已取消
)

var OrderStatusTransitions = map[int]int{
	OrderStatusPending:  OrderStatusAccepted, // 待处理 -> 已接单
	OrderStatusAccepted: OrderStatusShipped,  // 已接单 -> 已发货
	OrderStatusShipped:  OrderStatusComplete, // 已发货 -> 已完成
	// 特殊状态转换
	OrderStatusRejected: OrderStatusRejected, // 已拒绝状态不变
	OrderStatusComplete: OrderStatusComplete, // 已完成状态不变
	OrderStatusCanceled: OrderStatusCanceled, // 已取消状态不变
}

// DefaultOrderStatusFlow 默认订单流转状态配置
var DefaultOrderStatusFlow = `{
  "statuses": [
    {
      "value": 0,
      "label": "待处理",
      "type": "warning",
      "isFinal": false,
      "actions": [
        {
          "name": "接单",
          "nextStatus": 1,
          "nextStatusLabel": "已接单"
        },
        {
          "name": "取消",
          "nextStatus": 10,
          "nextStatusLabel": "已取消"
        }
      ]
    },
    {
      "value": 1,
      "label": "已接单",
      "type": "primary",
      "isFinal": false,
      "actions": [
        {
          "name": "完成",
          "nextStatus": 9,
          "nextStatusLabel": "已完成"
        },
        {
          "name": "取消",
          "nextStatus": 10,
          "nextStatusLabel": "已取消"
        }
      ]
    },
    {
      "value": 9,
      "label": "已完成",
      "type": "success",
      "isFinal": true,
      "actions": []
    },
    {
      "value": 10,
      "label": "已取消",
      "type": "info",
      "isFinal": true,
      "actions": []
    }
  ]
}`

// 订单状态变更日志
type OrderStatusLog struct {
	ID          snowflake.ID `gorm:"primarykey" json:"id"`
	OrderID     snowflake.ID `gorm:"type:bigint unsigned" json:"order_id"`
	OldStatus   int          `json:"old_status"`
	NewStatus   int          `json:"new_status"`
	ChangedTime time.Time    `json:"changed_time"`
}

type OrderElement struct {
	ID         snowflake.ID `gorm:"primarykey;column:id;type:bigint unsigned" json:"id,omitempty"`
	UserID     snowflake.ID `gorm:"column:user_id" json:"user_id"`
	ShopID     snowflake.ID `gorm:"column:shop_id;type:bigint unsigned;index;not null" json:"shop_id"`
	TotalPrice Price        `gorm:"column:total_price;type:double" json:"total_price"`
	Status     int          `gorm:"column:status" json:"status"`
	Remark     string       `gorm:"column:remark" json:"remark"`
	CreatedAt  time.Time    `gorm:"column:created_at" json:"created_at"`
	UpdatedAt  time.Time    `gorm:"column:updated_at" json:"updated_at"`
}
