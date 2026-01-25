package product

import (
	"time"

	"github.com/bwmarrin/snowflake"
	"orderease/models"
)

// ProductStatus 商品状态
type ProductStatus string

const (
	ProductStatusPending ProductStatus = "pending" // 待上架
	ProductStatusOnline  ProductStatus = "online"  // 已上架
	ProductStatusOffline ProductStatus = "offline" // 已下架
)

// Product 商品聚合根
type Product struct {
	id               snowflake.ID
	shopID           uint64
	name             string
	description      string
	price            float64
	stock            int
	imageURL         string
	status           ProductStatus
	optionCategories []models.ProductOptionCategory
	createdAt        time.Time
	updatedAt        time.Time
}

// NewProduct 创建新商品
func NewProduct(shopID uint64, name string, price float64, stock int) *Product {
	return &Product{
		shopID:    shopID,
		name:      name,
		price:     price,
		stock:     stock,
		status:    ProductStatusPending,
		createdAt: time.Now(),
		updatedAt: time.Now(),
	}
}

// NewProductWithDefaults 创建带完整默认值的商品
// 封装商品创建逻辑，避免在 Handler 中多次调用 Setter
func NewProductWithDefaults(shopID uint64, name string, price float64, stock int,
	description string, imageURL string, optionCategories []models.ProductOptionCategory) *Product {

	return &Product{
		shopID:           shopID,
		name:             name,
		price:            price,
		stock:            stock,
		description:      description,
		imageURL:         imageURL,
		status:           ProductStatusPending, // 默认状态
		optionCategories: optionCategories,
		createdAt:        time.Now(),
		updatedAt:        time.Now(),
	}
}

// Getters
func (p *Product) ID() snowflake.ID {
	return p.id
}

func (p *Product) ShopID() uint64 {
	return p.shopID
}

func (p *Product) Name() string {
	return p.name
}

func (p *Product) Description() string {
	return p.description
}

func (p *Product) Price() float64 {
	return p.price
}

func (p *Product) Stock() int {
	return p.stock
}

func (p *Product) ImageURL() string {
	return p.imageURL
}

func (p *Product) Status() ProductStatus {
	return p.status
}

func (p *Product) OptionCategories() []models.ProductOptionCategory {
	return p.optionCategories
}

func (p *Product) CreatedAt() time.Time {
	return p.createdAt
}

func (p *Product) UpdatedAt() time.Time {
	return p.updatedAt
}

// Setters
func (p *Product) SetID(id snowflake.ID) {
	p.id = id
}

func (p *Product) SetName(name string) {
	p.name = name
}

func (p *Product) SetDescription(desc string) {
	p.description = desc
}

func (p *Product) SetPrice(price float64) {
	p.price = price
}

func (p *Product) SetStock(stock int) {
	p.stock = stock
}

func (p *Product) SetImageURL(url string) {
	p.imageURL = url
}

func (p *Product) SetStatus(status ProductStatus) {
	p.status = status
}

func (p *Product) SetOptionCategories(categories []models.ProductOptionCategory) {
	p.optionCategories = categories
}

func (p *Product) SetCreatedAt(t time.Time) {
	p.createdAt = t
}

func (p *Product) SetUpdatedAt(t time.Time) {
	p.updatedAt = t
}

// ==================== 业务方法 ====================

// IsOnline 判断商品是否已上架
func (p *Product) IsOnline() bool {
	return p.status == ProductStatusOnline
}

// IsOffline 判断商品是否已下架
func (p *Product) IsOffline() bool {
	return p.status == ProductStatusOffline
}

// IsPending 判断商品是否待上架
func (p *Product) IsPending() bool {
	return p.status == ProductStatusPending
}

// InStock 判断是否有库存
func (p *Product) InStock() bool {
	return p.stock > 0
}

// HasEnoughStock 判断库存是否足够
func (p *Product) HasEnoughStock(quantity int) bool {
	return p.stock >= quantity
}

// DecreaseStock 减少库存
func (p *Product) DecreaseStock(quantity int) {
	if p.stock >= quantity {
		p.stock -= quantity
	}
}

// IncreaseStock 增加库存
func (p *Product) IncreaseStock(quantity int) {
	p.stock += quantity
}

// ToModel 转换为持久化模型
func (p *Product) ToModel() *models.Product {
	return &models.Product{
		ID:               p.id,
		ShopID:           p.shopID,
		Name:             p.name,
		Description:      p.description,
		Price:            p.price,
		Stock:            p.stock,
		ImageURL:         p.imageURL,
		Status:           string(p.status),
		OptionCategories: p.optionCategories,
		CreatedAt:        p.createdAt,
		UpdatedAt:        p.updatedAt,
	}
}

// ProductFromModel 从持久化模型创建领域实体
func ProductFromModel(model *models.Product) *Product {
	return &Product{
		id:               model.ID,
		shopID:           model.ShopID,
		name:             model.Name,
		description:      model.Description,
		price:            model.Price,
		stock:            model.Stock,
		imageURL:         model.ImageURL,
		status:           ProductStatus(model.Status),
		optionCategories: model.OptionCategories,
		createdAt:        model.CreatedAt,
		updatedAt:        model.UpdatedAt,
	}
}
