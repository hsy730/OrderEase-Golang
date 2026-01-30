package order

import (
	"fmt"

	"github.com/bwmarrin/snowflake"
)

// CreateOrderRequest 创建订单请求 DTO
type CreateOrderRequest struct {
	ID     snowflake.ID             `json:"id"`
	UserID snowflake.ID             `json:"user_id"`
	ShopID snowflake.ID             `json:"shop_id"`
	Items  []CreateOrderItemRequest `json:"items"`
	Remark string                   `json:"remark"`
	Status int                      `json:"status"`
}

// Validate 验证创建订单请求
func (r *CreateOrderRequest) Validate() error {
	if r.UserID == 0 {
		return fmt.Errorf("用户ID不能为空")
	}
	if r.ShopID == 0 {
		return fmt.Errorf("店铺ID不能为空")
	}
	if len(r.Items) == 0 {
		return fmt.Errorf("订单项不能为空")
	}
	return nil
}

// CreateOrderItemRequest 创建订单项请求 DTO
type CreateOrderItemRequest struct {
	ProductID snowflake.ID            `json:"product_id"`
	Quantity  int                     `json:"quantity"`
	Price     float64                 `json:"price"`
	Options   []CreateOrderItemOption `json:"options"`
}

// Validate 验证订单项
func (r *CreateOrderItemRequest) Validate() error {
	if r.ProductID == 0 {
		return fmt.Errorf("商品ID不能为空")
	}
	if r.Quantity <= 0 {
		return fmt.Errorf("商品数量必须大于0")
	}
	if r.Price < 0 {
		return fmt.Errorf("商品价格不能为负数")
	}
	return nil
}

// CreateOrderItemOption 创建订单项选项请求 DTO
type CreateOrderItemOption struct {
	CategoryID snowflake.ID `json:"category_id"`
	OptionID   snowflake.ID `json:"option_id"`
}

// AdvanceSearchOrderRequest 高级查询订单请求 DTO
type AdvanceSearchOrderRequest struct {
	Page      int          `json:"page"`
	PageSize  int          `json:"pageSize"`
	UserID    string       `json:"user_id"`
	Status    []int        `json:"status"`
	StartTime string       `json:"start_time"`
	EndTime   string       `json:"end_time"`
	ShopID    snowflake.ID `json:"shop_id"`
}

// Validate 验证高级查询请求
func (r *AdvanceSearchOrderRequest) Validate() error {
	if r.Page < 1 {
		r.Page = 1
	}
	if r.PageSize < 1 || r.PageSize > 100 {
		r.PageSize = 10
	}
	if r.ShopID == 0 {
		return fmt.Errorf("店铺ID不能为空")
	}
	return nil
}

// ToggleOrderStatusRequest 切换订单状态请求 DTO
type ToggleOrderStatusRequest struct {
	ID         snowflake.ID `json:"id" binding:"required"`
	ShopID     snowflake.ID `json:"shop_id" binding:"required"`
	NextStatus int          `json:"next_status" binding:"required"`
}

// Validate 验证切换订单状态请求
func (r *ToggleOrderStatusRequest) Validate() error {
	if r.ID == 0 {
		return fmt.Errorf("订单ID不能为空")
	}
	if r.ShopID == 0 {
		return fmt.Errorf("店铺ID不能为空")
	}
	return nil
}
