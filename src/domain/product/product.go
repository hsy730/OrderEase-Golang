package product

import (
	"errors"
	"time"

	"orderease/domain/shared"
)

type ProductStatus string

const (
	ProductStatusPending ProductStatus = "pending"
	ProductStatusOnline  ProductStatus = "online"
	ProductStatusOffline ProductStatus = "offline"
)

func (s ProductStatus) String() string {
	return string(s)
}

func (s ProductStatus) IsValid() bool {
	return s == ProductStatusPending || s == ProductStatusOnline || s == ProductStatusOffline
}

func (s ProductStatus) CanTransitionTo(newStatus ProductStatus) bool {
	transitions := map[ProductStatus][]ProductStatus{
		ProductStatusPending: {ProductStatusOnline},
		ProductStatusOnline:  {ProductStatusOffline},
		ProductStatusOffline: {ProductStatusOnline},
	}

	allowed, exists := transitions[s]
	if !exists {
		return false
	}

	for _, status := range allowed {
		if status == newStatus {
			return true
		}
	}

	return false
}

type Product struct {
	ID              shared.ID
	ShopID          uint64
	Name            string
	Description     string
	Price           shared.Price
	Stock           int
	ImageURL        string
	Status          ProductStatus
	CreatedAt       time.Time
	UpdatedAt       time.Time
	OptionCategories []ProductOptionCategory
}

type ProductOptionCategory struct {
	ID           shared.ID
	ProductID    shared.ID
	Name         string
	IsRequired   bool
	IsMultiple   bool
	DisplayOrder int
	CreatedAt    time.Time
	UpdatedAt    time.Time
	Options      []ProductOption
}

type ProductOption struct {
	ID              shared.ID
	CategoryID      shared.ID
	Name            string
	PriceAdjustment float64
	DisplayOrder    int
	IsDefault       bool
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

func NewProduct(shopID uint64, name, description string, price shared.Price, stock int) (*Product, error) {
	if shopID == 0 {
		return nil, errors.New("店铺ID不能为空")
	}

	if name == "" {
		return nil, errors.New("商品名称不能为空")
	}

	if price.IsZero() {
		return nil, errors.New("商品价格不能为零")
	}

	if stock < 0 {
		return nil, errors.New("商品库存不能为负数")
	}

	now := time.Now()

	return &Product{
		ID:          shared.ID(0),
		ShopID:      shopID,
		Name:        name,
		Description: description,
		Price:       price,
		Stock:       stock,
		Status:      ProductStatusPending,
		CreatedAt:   now,
		UpdatedAt:   now,
	}, nil
}

func (p *Product) UpdateStock(quantity int) error {
	if quantity < 0 {
		return errors.New("库存不能为负数")
	}
	p.Stock = quantity
	p.UpdatedAt = time.Now()
	return nil
}

func (p *Product) DecreaseStock(quantity int) error {
	if p.Stock < quantity {
		return errors.New("库存不足")
	}
	p.Stock -= quantity
	p.UpdatedAt = time.Now()
	return nil
}

func (p *Product) IncreaseStock(quantity int) error {
	if quantity < 0 {
		return errors.New("增加的库存数量不能为负数")
	}
	p.Stock += quantity
	p.UpdatedAt = time.Now()
	return nil
}

func (p *Product) ChangeStatus(newStatus ProductStatus) error {
	if !newStatus.IsValid() {
		return errors.New("无效的商品状态")
	}

	if !p.Status.CanTransitionTo(newStatus) {
		return errors.New("不允许的状态转换")
	}

	p.Status = newStatus
	p.UpdatedAt = time.Now()
	return nil
}

func (p *Product) IsAvailable() bool {
	return p.Status == ProductStatusOnline || p.Status == ProductStatusPending
}

func (p *Product) HasStock(quantity int) bool {
	return p.Stock >= quantity
}

func NewProductOptionCategory(productID shared.ID, name string, isRequired, isMultiple bool, displayOrder int) (*ProductOptionCategory, error) {
	if productID.IsZero() {
		return nil, errors.New("商品ID不能为空")
	}

	if name == "" {
		return nil, errors.New("类别名称不能为空")
	}

	now := time.Now()

	return &ProductOptionCategory{
		ID:           shared.ID(0),
		ProductID:    productID,
		Name:         name,
		IsRequired:   isRequired,
		IsMultiple:   isMultiple,
		DisplayOrder: displayOrder,
		CreatedAt:    now,
		UpdatedAt:    now,
	}, nil
}

func NewProductOption(categoryID shared.ID, name string, priceAdjustment float64, isDefault bool, displayOrder int) (*ProductOption, error) {
	if categoryID.IsZero() {
		return nil, errors.New("类别ID不能为空")
	}

	if name == "" {
		return nil, errors.New("选项名称不能为空")
	}

	now := time.Now()

	return &ProductOption{
		ID:              shared.ID(0),
		CategoryID:      categoryID,
		Name:            name,
		PriceAdjustment: priceAdjustment,
		IsDefault:       isDefault,
		DisplayOrder:    displayOrder,
		CreatedAt:       now,
		UpdatedAt:       now,
	}, nil
}
