package models

import (
	"time"

	"github.com/bwmarrin/snowflake"
)

// ProductOptionCategory 商品参数类别（如：大小、甜度、颜色等）
type ProductOptionCategory struct {
	ID           snowflake.ID `gorm:"column:id;primarykey;autoIncrement:false;type:bigint unsigned" json:"id"` // 主键
	ProductID    snowflake.ID `gorm:"column:product_id;index;not null;type:bigint unsigned" json:"product_id"`
	Name         string       `gorm:"column:name;size:100" json:"name"`                    // 类别名称，如"大小"、"甜度"
	IsRequired   bool         `gorm:"column:is_required" json:"is_required"`               // 是否必填
	IsMultiple   bool         `gorm:"column:is_multiple" json:"is_multiple"`               // 是否允许多选
	DisplayOrder int          `gorm:"column:display_order;default:0" json:"display_order"` // 显示顺序
	CreatedAt    time.Time    `gorm:"column:created_at" json:"created_at"`
	UpdatedAt    time.Time    `gorm:"column:updated_at" json:"updated_at"`

	// 在ProductOptionCategory结构体中
	// 一对多关联：一个类别有多个选项
	Options []ProductOption `gorm:"foreignKey:CategoryID;constraint:OnDelete:CASCADE" json:"options,omitempty"`
}

// ProductOption 商品参数选项（如：小杯、中杯、无糖、白色等）
type ProductOption struct {
	ID              snowflake.ID `gorm:"column:id;primarykey;autoIncrement:false;type:bigint unsigned" json:"id"`
	CategoryID      snowflake.ID `gorm:"column:category_id;index;not null;type:bigint unsigned" json:"category_id"`
	Name            string       `gorm:"column:name;size:100" json:"name"`                    // 选项名称，如"小杯"、"无糖"
	PriceAdjustment float64      `gorm:"column:price_adjustment" json:"price_adjustment"`     // 价格调整值，可以是正或负
	DisplayOrder    int          `gorm:"column:display_order;default:0" json:"display_order"` // 显示顺序
	IsDefault       bool         `gorm:"column:is_default;default:false" json:"is_default"`   // 是否为默认选项
	CreatedAt       time.Time    `gorm:"column:created_at" json:"created_at"`
	UpdatedAt       time.Time    `gorm:"column:updated_at" json:"updated_at"`

	// 多对一关联：多个选项属于一个类别，不指定foreignKey
	Category *ProductOptionCategory `json:"category,omitempty"`
}
