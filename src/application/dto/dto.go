package dto

import (
	"orderease/domain/order"
	"orderease/domain/product"
	"orderease/domain/shared"
	"orderease/domain/user"
	"time"
)

type CreateOrderRequest struct {
	UserID shared.ID
	ShopID shared.ID
	Items  []CreateOrderItemRequest
	Remark string
}

type CreateOrderItemRequest struct {
	ProductID shared.ID
	Quantity  int
	Price     float64
	Options   []CreateOrderItemOption
}

type CreateOrderItemOption struct {
	CategoryID shared.ID
	OptionID   shared.ID
}

type OrderResponse struct {
	ID         shared.ID
	UserID     shared.ID
	ShopID     shared.ID
	TotalPrice shared.Price
	Status     order.OrderStatus
	Remark     string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type OrderDetailResponse struct {
	ID         shared.ID
	UserID     shared.ID
	ShopID     shared.ID
	TotalPrice shared.Price
	Status     order.OrderStatus
	Remark     string
	CreatedAt  time.Time
	UpdatedAt  time.Time
	Items      []OrderItemResponse
}

type OrderItemResponse struct {
	ID                shared.ID
	ProductID         shared.ID
	Quantity          int
	Price             shared.Price
	TotalPrice        shared.Price
	ProductName       string
	ProductDescription string
	ProductImageURL   string
	Options           []OrderItemOptionResponse
}

type OrderItemOptionResponse struct {
	ID              shared.ID
	CategoryID      shared.ID
	OptionID        shared.ID
	OptionName      string
	CategoryName    string
	PriceAdjustment float64
}

type OrderListResponse struct {
	Total    int64
	Page     int
	PageSize int
	Data     []OrderResponse
}

type SearchOrdersRequest struct {
	ShopID       shared.ID
	UserID       string
	Statuses     []order.OrderStatus
	StartTime    time.Time
	EndTime      time.Time
	StartTimeStr string
	EndTimeStr   string
	Page         int
	PageSize      int
}

type CreateProductRequest struct {
	ShopID          shared.ID
	Name            string
	Description     string
	Price           float64
	Stock           int
	ImageURL        string
	OptionCategories []CreateProductOptionCategoryRequest
}

type CreateProductOptionCategoryRequest struct {
	Name         string
	IsRequired   bool
	IsMultiple   bool
	DisplayOrder int
	Options      []CreateProductOptionRequest
}

type CreateProductOptionRequest struct {
	Name            string
	PriceAdjustment float64
	IsDefault       bool
	DisplayOrder    int
}

type ProductResponse struct {
	ID              shared.ID
	ShopID          shared.ID
	Name            string
	Description     string
	Price           shared.Price
	Stock           int
	ImageURL        string
	Status          product.ProductStatus
	CreatedAt       time.Time
	UpdatedAt       time.Time
	OptionCategories []ProductOptionCategoryResponse
}

type ProductOptionCategoryResponse struct {
	ID           shared.ID
	ProductID    shared.ID
	Name         string
	IsRequired   bool
	IsMultiple   bool
	DisplayOrder int
	Options      []ProductOptionResponse
}

type ProductOptionResponse struct {
	ID              shared.ID
	CategoryID      shared.ID
	Name            string
	PriceAdjustment float64
	DisplayOrder    int
	IsDefault       bool
}

type ProductListResponse struct {
	Total    int64
	Page     int
	PageSize int
	Data     []ProductResponse
}

type UpdateProductStatusRequest struct {
	ID     shared.ID
	Status product.ProductStatus
}

type CreateShopRequest struct {
	Name            string    `json:"name"`
	OwnerUsername   string    `json:"owner_username"`
	OwnerPassword   string    `json:"owner_password"`
	ContactPhone    string    `json:"contact_phone"`
	ContactEmail    string    `json:"contact_email"`
	Description     string    `json:"description"`
	ValidUntil      time.Time `json:"valid_until"`
	Address         string    `json:"address"`
	Settings        string    `json:"settings"`
	OrderStatusFlow *order.OrderStatusFlow `json:"order_status_flow"`
}

type UpdateShopRequest struct {
	ID              shared.ID  `json:"id"`
	OwnerUsername   string    `json:"owner_username"`
	OwnerPassword   *string   `json:"owner_password"`
	Name            string    `json:"name"`
	ContactPhone    string    `json:"contact_phone"`
	ContactEmail    string    `json:"contact_email"`
	Description     string    `json:"description"`
	ValidUntil      time.Time `json:"valid_until"`
	Address         string    `json:"address"`
	Settings        string    `json:"settings"`
	OrderStatusFlow *order.OrderStatusFlow `json:"order_status_flow"`
}

type ShopResponse struct {
	ID            shared.ID
	Name          string
	OwnerUsername string
	ContactPhone  string
	ContactEmail  string
	Address       string
	Description   string
	CreatedAt     time.Time
	UpdatedAt     time.Time
	ValidUntil    time.Time
	Settings      string
	ImageURL      string
	OrderStatusFlow order.OrderStatusFlow
}

type ShopListResponse struct {
	Total    int64
	Page     int
	PageSize int
	Data     []ShopResponse
}

type CreateTagRequest struct {
	ShopID      shared.ID  `json:"shop_id"`
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
	Total int64
	Tags  []TagResponse
}

type CreateUserRequest struct {
	Name     string
	Role     user.UserRole
	Type     user.UserType
	Password string
	Phone    string
	Address  string
}

type UpdateUserRequest struct {
	ID       shared.ID
	Name     string
	Role     string
	Type     string
	Password *string
	Phone    string
	Address  string
}

type UserResponse struct {
	ID        shared.ID
	Name      string
	Role      user.UserRole
	Type      user.UserType
	Phone     string
	Address   string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type UserListResponse struct {
	Total    int64
	Page     int
	PageSize int
	Data     []UserResponse
}
