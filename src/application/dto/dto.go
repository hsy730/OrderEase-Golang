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
	ShopID uint64
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
	ShopID     uint64
	TotalPrice shared.Price
	Status     order.OrderStatus
	Remark     string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type OrderDetailResponse struct {
	ID         shared.ID
	UserID     shared.ID
	ShopID     uint64
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
	ShopID       uint64
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
	ShopID          uint64
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
	ShopID          uint64
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
	Name            string
	OwnerUsername   string
	OwnerPassword   string
	ContactPhone    string
	ContactEmail    string
	Description     string
	ValidUntil      time.Time
	Address         string
	Settings        string
	OrderStatusFlow *order.OrderStatusFlow
}

type UpdateShopRequest struct {
	ID              uint64
	OwnerUsername   string
	OwnerPassword   *string
	Name            string
	ContactPhone    string
	ContactEmail    string
	Description     string
	ValidUntil      time.Time
	Address         string
	Settings        string
	OrderStatusFlow *order.OrderStatusFlow
}

type ShopResponse struct {
	ID            uint64
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
	ShopID      uint64
	Name        string
	Description string
}

type TagResponse struct {
	ID          int
	ShopID      uint64
	Name        string
	Description string
	CreatedAt   time.Time
	UpdatedAt   time.Time
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
