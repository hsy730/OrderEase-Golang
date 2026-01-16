package dto

import (
	"orderease/domain/order"
	"orderease/domain/product"
	"orderease/domain/shared"
	"orderease/domain/user"
	"time"
)

type CreateOrderRequest struct {
	UserID shared.ID                `json:"user_id"`
	ShopID shared.ID                `json:"shop_id"`
	Items  []CreateOrderItemRequest `json:"items"`
	Remark string                   `json:"remark"`
}

type CreateOrderItemRequest struct {
	ProductID shared.ID               `json:"product_id"`
	Quantity  int                     `json:"quantity"`
	Price     float64                 `json:"price"`
	Options   []CreateOrderItemOption `json:"options"`
}

type CreateOrderItemOption struct {
	CategoryID shared.ID `json:"category_id"`
	OptionID   shared.ID `json:"option_id"`
}

type UpdateOrderRequest struct {
	ID     shared.ID                `json:"id" binding:"required"`
	UserID shared.ID                `json:"user_id"`
	ShopID shared.ID                `json:"shop_id" binding:"required"`
	Items  []CreateOrderItemRequest `json:"items" binding:"required"`
	Remark string                   `json:"remark"`
	Status order.OrderStatus        `json:"status"`
}

type OrderResponse struct {
	ID         shared.ID         `json:"id"`
	UserID     shared.ID         `json:"user_id"`
	ShopID     shared.ID         `json:"shop_id"`
	TotalPrice shared.Price      `json:"total_price"`
	Status     order.OrderStatus `json:"status"`
	Remark     string            `json:"remark"`
	CreatedAt  time.Time         `json:"created_at"`
	UpdatedAt  time.Time         `json:"updated_at"`
}

type OrderDetailResponse struct {
	ID         shared.ID           `json:"id"`
	UserID     shared.ID           `json:"user_id"`
	ShopID     shared.ID           `json:"shop_id"`
	TotalPrice shared.Price        `json:"total_price"`
	Status     order.OrderStatus   `json:"status"`
	Remark     string              `json:"remark"`
	CreatedAt  time.Time           `json:"created_at"`
	UpdatedAt  time.Time           `json:"updated_at"`
	Items      []OrderItemResponse `json:"items"`
}

type OrderItemResponse struct {
	ID                 shared.ID                 `json:"id"`
	ProductID          shared.ID                 `json:"product_id"`
	Quantity           int                       `json:"quantity"`
	Price              shared.Price              `json:"price"`
	TotalPrice         shared.Price              `json:"total_price"`
	ProductName        string                    `json:"product_name"`
	ProductDescription string                    `json:"product_description"`
	ProductImageURL    string                    `json:"product_image_url"`
	Options            []OrderItemOptionResponse `json:"options"`
}

type OrderItemOptionResponse struct {
	ID              shared.ID `json:"id"`
	CategoryID      shared.ID `json:"category_id"`
	OptionID        shared.ID `json:"option_id"`
	OptionName      string    `json:"option_name"`
	CategoryName    string    `json:"category_name"`
	PriceAdjustment float64   `json:"price_adjustment"`
}

type OrderListResponse struct {
	Total    int64           `json:"total"`
	Page     int             `json:"page"`
	PageSize int             `json:"page_size"`
	Data     []OrderResponse `json:"data"`
}

type SearchOrdersRequest struct {
	ShopID       shared.ID           `json:"shop_id"`
	UserID       string              `json:"user_id"`
	Statuses     []order.OrderStatus `json:"statuses"`
	StartTime    time.Time           `json:"start_time"`
	EndTime      time.Time           `json:"end_time"`
	StartTimeStr string              `json:"start_time_str"`
	EndTimeStr   string              `json:"end_time_str"`
	Page         int                 `json:"page"`
	PageSize     int                 `json:"page_size"`
}

type AdvanceSearchOrderRequest struct {
	ShopID    shared.ID           `json:"shop_id"`
	UserID    string              `json:"user_id"`
	Status    []int               `json:"status"`
	StartTime string              `json:"start_time"`
	EndTime   string              `json:"end_time"`
	Page      int                 `json:"page"`
	PageSize  int                 `json:"page_size"`
}

type CreateProductRequest struct {
	ShopID           shared.ID                            `json:"shop_id"`
	Name             string                               `json:"name"`
	Description      string                               `json:"description"`
	Price            float64                              `json:"price"`
	Stock            int                                  `json:"stock"`
	ImageURL         string                               `json:"image_url"`
	OptionCategories []CreateProductOptionCategoryRequest `json:"option_categories"`
}

type CreateProductOptionCategoryRequest struct {
	Name         string                       `json:"name"`
	IsRequired   bool                         `json:"is_required"`
	IsMultiple   bool                         `json:"is_multiple"`
	DisplayOrder int                          `json:"display_order"`
	Options      []CreateProductOptionRequest `json:"options"`
}

type CreateProductOptionRequest struct {
	Name            string  `json:"name"`
	PriceAdjustment float64 `json:"price_adjustment"`
	IsDefault       bool    `json:"is_default"`
	DisplayOrder    int     `json:"display_order"`
}

type ProductResponse struct {
	ID               shared.ID                       `json:"id"`
	ShopID           shared.ID                       `json:"shop_id"`
	Name             string                          `json:"name"`
	Description      string                          `json:"description"`
	Price            shared.Price                    `json:"price"`
	Stock            int                             `json:"stock"`
	ImageURL         string                          `json:"image_url"`
	Status           product.ProductStatus           `json:"status"`
	CreatedAt        time.Time                       `json:"created_at"`
	UpdatedAt        time.Time                       `json:"updated_at"`
	OptionCategories []ProductOptionCategoryResponse `json:"option_categories"`
}

type ProductOptionCategoryResponse struct {
	ID           shared.ID               `json:"id"`
	ProductID    shared.ID               `json:"product_id"`
	Name         string                  `json:"name"`
	IsRequired   bool                    `json:"is_required"`
	IsMultiple   bool                    `json:"is_multiple"`
	DisplayOrder int                     `json:"display_order"`
	Options      []ProductOptionResponse `json:"options"`
}

type ProductOptionResponse struct {
	ID              shared.ID `json:"id"`
	CategoryID      shared.ID `json:"category_id"`
	Name            string    `json:"name"`
	PriceAdjustment float64   `json:"price_adjustment"`
	DisplayOrder    int       `json:"display_order"`
	IsDefault       bool      `json:"is_default"`
}

type ProductListResponse struct {
	Total    int64             `json:"total"`
	Page     int               `json:"page"`
	PageSize int               `json:"page_size"`
	Data     []ProductResponse `json:"data"`
}

type UpdateProductStatusRequest struct {
	ID     shared.ID             `json:"id"`
	Status product.ProductStatus `json:"status"`
}

type CreateShopRequest struct {
	Name            string                 `json:"name"`
	OwnerUsername   string                 `json:"owner_username"`
	OwnerPassword   string                 `json:"owner_password"`
	ContactPhone    string                 `json:"contact_phone"`
	ContactEmail    string                 `json:"contact_email"`
	Description     string                 `json:"description"`
	ValidUntil      time.Time              `json:"valid_until"`
	Address         string                 `json:"address"`
	Settings        string                 `json:"settings"`
	OrderStatusFlow *order.OrderStatusFlow `json:"order_status_flow"`
}

type UpdateShopRequest struct {
	ID              shared.ID              `json:"id"`
	OwnerUsername   string                 `json:"owner_username"`
	OwnerPassword   *string                `json:"owner_password"`
	Name            string                 `json:"name"`
	ContactPhone    string                 `json:"contact_phone"`
	ContactEmail    string                 `json:"contact_email"`
	Description     string                 `json:"description"`
	ValidUntil      time.Time              `json:"valid_until"`
	Address         string                 `json:"address"`
	Settings        string                 `json:"settings"`
	OrderStatusFlow *order.OrderStatusFlow `json:"order_status_flow"`
}

type ShopResponse struct {
	ID              shared.ID             `json:"id"`
	Name            string                `json:"name"`
	OwnerUsername   string                `json:"owner_username"`
	ContactPhone    string                `json:"contact_phone"`
	ContactEmail    string                `json:"contact_email"`
	Address         string                `json:"address"`
	Description     string                `json:"description"`
	CreatedAt       time.Time             `json:"created_at"`
	UpdatedAt       time.Time             `json:"updated_at"`
	ValidUntil      time.Time             `json:"valid_until"`
	Settings        string                `json:"settings"`
	ImageURL        string                `json:"image_url"`
	OrderStatusFlow order.OrderStatusFlow `json:"order_status_flow"`
}

type ShopListResponse struct {
	Total    int64          `json:"total"`
	Page     int            `json:"page"`
	PageSize int            `json:"page_size"`
	Data     []ShopResponse `json:"data"`
}

type CreateTagRequest struct {
	ShopID      shared.ID `json:"shop_id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
}

type TagResponse struct {
	ID          int       `json:"id"`
	ShopID      shared.ID `json:"shop_id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type TagListResponse struct {
	Total int64         `json:"total"`
	Tags  []TagResponse `json:"tags"`
}

type CreateUserRequest struct {
	Name     string        `json:"name"`
	Role     user.UserRole `json:"role"`
	Type     user.UserType `json:"type"`
	Password string        `json:"password"`
	Phone    string        `json:"phone"`
	Address  string        `json:"address"`
}

type UpdateUserRequest struct {
	ID       shared.ID `json:"id"`
	Name     string    `json:"name"`
	Role     string    `json:"role"`
	Type     string    `json:"type"`
	Password *string   `json:"password"`
	Phone    string    `json:"phone"`
	Address  string    `json:"address"`
}

type UserResponse struct {
	ID        shared.ID     `json:"id"`
	Name      string        `json:"name"`
	Role      user.UserRole `json:"role"`
	Type      user.UserType `json:"type"`
	Phone     string        `json:"phone"`
	Address   string        `json:"address"`
	CreatedAt time.Time     `json:"created_at"`
	UpdatedAt time.Time     `json:"updated_at"`
}

type UserListResponse struct {
	Total    int64          `json:"total"`
	Page     int            `json:"page"`
	PageSize int            `json:"page_size"`
	Data     []UserResponse `json:"data"`
}
