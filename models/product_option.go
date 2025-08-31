package models

import (
	"time"

	"github.com/bwmarrin/snowflake"
)

// ProductOptionCategory 商品参数类别（如：大小、甜度、颜色等）
type ProductOptionCategory struct {
	ID           snowflake.ID `gorm:"primarykey" json:"id"`
	ProductID    snowflake.ID `gorm:"index;not null" json:"product_id"`
	Name         string       `gorm:"size:100" json:"name"`           // 类别名称，如"大小"、"甜度"
	IsRequired   bool         `json:"is_required"`                    // 是否必填
	IsMultiple   bool         `json:"is_multiple"`                    // 是否允许多选
	DisplayOrder int          `gorm:"default:0" json:"display_order"` // 显示顺序
	CreatedAt    time.Time    `json:"created_at"`
	UpdatedAt    time.Time    `json:"updated_at"`

	// 在ProductOptionCategory结构体中
	// 一对多关联：一个类别有多个选项
	Options []ProductOption `gorm:"foreignKey:CategoryID;constraint:OnDelete:CASCADE" json:"options,omitempty"`
}

// ProductOption 商品参数选项（如：小杯、中杯、无糖、白色等）
type ProductOption struct {
	ID              snowflake.ID `gorm:"primarykey" json:"id"`
	CategoryID      snowflake.ID `gorm:"index;not null" json:"category_id"`
	Name            string       `gorm:"size:100" json:"name"`            // 选项名称，如"小杯"、"无糖"
	PriceAdjustment float64      `json:"price_adjustment"`                // 价格调整值，可以是正或负
	DisplayOrder    int          `gorm:"default:0" json:"display_order"`  // 显示顺序
	IsDefault       bool         `gorm:"default:false" json:"is_default"` // 是否为默认选项
	CreatedAt       time.Time    `json:"created_at"`
	UpdatedAt       time.Time    `json:"updated_at"`

	// 多对一关联：多个选项属于一个类别，不指定foreignKey
	Category *ProductOptionCategory `json:"category,omitempty"`
}

// OrderItemOption 订单项选择的商品参数选项
type OrderItemOption struct {
	ID              snowflake.ID `gorm:"primarykey" json:"id"`
	OrderItemID     snowflake.ID `gorm:"index;not null" json:"order_item_id"`
	OptionID        snowflake.ID `json:"option_id"`
	OptionName      string       `gorm:"size:100" json:"option_name"`   // 选项名称快照
	CategoryName    string       `gorm:"size:100" json:"category_name"` // 类别名称快照
	PriceAdjustment float64      `json:"price_adjustment"`              // 价格调整快照
	CreatedAt       time.Time    `json:"created_at"`
	UpdatedAt       time.Time    `json:"updated_at"`

	// 关联
	OrderItem *OrderItem `gorm:"foreignKey:OrderItemID" json:"order_item,omitempty"`
}
