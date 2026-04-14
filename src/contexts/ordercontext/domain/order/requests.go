package order

import (
	"fmt"

	"github.com/bwmarrin/snowflake"
	"orderease/models"
)

// CreateOrderRequest 创建订单请求 DTO
type CreateOrderRequest struct {
	ID     models.SnowflakeString   `json:"id"`
	UserID models.SnowflakeString   `json:"user_id"`
	ShopID models.SnowflakeString   `json:"shop_id"`
	Items  []CreateOrderItemRequest `json:"items"`
	Remark string                   `json:"remark"`
	Status int                      `json:"status"`
}

// GetUserID 获取用户ID作为 snowflake.ID
func (r *CreateOrderRequest) GetUserID() snowflake.ID {
	return r.UserID.ToSnowflakeID()
}

// GetShopID 获取店铺ID作为 snowflake.ID
func (r *CreateOrderRequest) GetShopID() snowflake.ID {
	return r.ShopID.ToSnowflakeID()
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
	ProductID models.SnowflakeString  `json:"product_id"`
	Quantity  int                     `json:"quantity"`
	Price     float64                 `json:"price"`
	Options   []CreateOrderItemOption `json:"options"`
}

// GetProductID 获取商品ID作为 snowflake.ID
func (r *CreateOrderItemRequest) GetProductID() snowflake.ID {
	return r.ProductID.ToSnowflakeID()
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
	CategoryID models.SnowflakeString `json:"category_id"`
	OptionID   models.SnowflakeString `json:"option_id"`
}

// GetCategoryID 获取类别ID作为 snowflake.ID
func (o *CreateOrderItemOption) GetCategoryID() snowflake.ID {
	return o.CategoryID.ToSnowflakeID()
}

// GetOptionID 获取选项ID作为 snowflake.ID
func (o *CreateOrderItemOption) GetOptionID() snowflake.ID {
	return o.OptionID.ToSnowflakeID()
}

// AdvanceSearchOrderRequest 高级查询订单请求 DTO
type AdvanceSearchOrderRequest struct {
	Page      int                  `json:"page"`
	PageSize  int                  `json:"pageSize"`
	UserID    string               `json:"user_id"`
	Status    []int                `json:"status"`
	StartTime string               `json:"start_time"`
	EndTime   string               `json:"end_time"`
	ShopID    models.SnowflakeString `json:"shop_id"`
}

// GetShopID 获取店铺ID作为 snowflake.ID
func (r *AdvanceSearchOrderRequest) GetShopID() snowflake.ID {
	return r.ShopID.ToSnowflakeID()
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
	ID         models.SnowflakeString `json:"id" binding:"required"`
	ShopID     models.SnowflakeString `json:"shop_id" binding:"required"`
	NextStatus int                    `json:"next_status" binding:"required"`
}

// GetID 获取订单ID作为 snowflake.ID
func (r *ToggleOrderStatusRequest) GetID() snowflake.ID {
	return r.ID.ToSnowflakeID()
}

// GetShopID 获取店铺ID作为 snowflake.ID
func (r *ToggleOrderStatusRequest) GetShopID() snowflake.ID {
	return r.ShopID.ToSnowflakeID()
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
