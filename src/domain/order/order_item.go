package order

import (
	"github.com/bwmarrin/snowflake"
	"orderease/models"
)

// OrderItem 订单项值对象
type OrderItem struct {
	id                snowflake.ID
	productID         snowflake.ID
	quantity          int
	price             models.Price
	totalPrice        models.Price
	productName       string
	productDescription string
	productImageURL   string
	options           []OrderItemOption
}

// OrderItemOption 订单项选项
type OrderItemOption struct {
	ID              snowflake.ID
	OrderItemID     snowflake.ID
	CategoryID      snowflake.ID
	OptionID        snowflake.ID
	OptionName      string
	CategoryName    string
	PriceAdjustment float64
}

// NewOrderItem 创建订单项
func NewOrderItem(productID snowflake.ID, quantity int, price models.Price) *OrderItem {
	return &OrderItem{
		id:         snowflake.ID(0), // 将在持久化时生成
		productID:  productID,
		quantity:   quantity,
		price:      price,
		totalPrice: models.Price(float64(quantity) * float64(price)),
	}
}

// Getters
func (oi *OrderItem) ID() snowflake.ID {
	return oi.id
}

func (oi *OrderItem) ProductID() snowflake.ID {
	return oi.productID
}

func (oi *OrderItem) Quantity() int {
	return oi.quantity
}

func (oi *OrderItem) Price() models.Price {
	return oi.price
}

func (oi *OrderItem) TotalPrice() models.Price {
	return oi.totalPrice
}

func (oi *OrderItem) ProductName() string {
	return oi.productName
}

func (oi *OrderItem) ProductDescription() string {
	return oi.productDescription
}

func (oi *OrderItem) ProductImageURL() string {
	return oi.productImageURL
}

func (oi *OrderItem) Options() []OrderItemOption {
	return oi.options
}

// Setters
func (oi *OrderItem) SetProductName(name string) {
	oi.productName = name
}

func (oi *OrderItem) SetProductDescription(desc string) {
	oi.productDescription = desc
}

func (oi *OrderItem) SetProductImageURL(url string) {
	oi.productImageURL = url
}

func (oi *OrderItem) SetTotalPrice(total models.Price) {
	oi.totalPrice = total
}

func (oi *OrderItem) SetPrice(price models.Price) {
	oi.price = price
}

func (oi *OrderItem) AddOption(option OrderItemOption) {
	oi.options = append(oi.options, option)
}

// ToModel 转换为持久化模型
func (oi *OrderItem) ToModel(orderID snowflake.ID) *models.OrderItem {
	modelOptions := make([]models.OrderItemOption, len(oi.options))
	for i, opt := range oi.options {
		modelOptions[i] = models.OrderItemOption{
			ID:              opt.ID,
			OrderItemID:     opt.OrderItemID,
			CategoryID:      opt.CategoryID,
			OptionID:        opt.OptionID,
			OptionName:      opt.OptionName,
			CategoryName:    opt.CategoryName,
			PriceAdjustment: opt.PriceAdjustment,
		}
	}

	return &models.OrderItem{
		ID:                 oi.id,
		OrderID:            orderID,
		ProductID:          oi.productID,
		Quantity:           oi.quantity,
		Price:              oi.price,
		TotalPrice:         oi.totalPrice,
		ProductName:        oi.productName,
		ProductDescription: oi.productDescription,
		ProductImageURL:    oi.productImageURL,
		Options:            modelOptions,
	}
}

// OrderItemFromModel 从持久化模型创建领域实体
func OrderItemFromModel(model *models.OrderItem) *OrderItem {
	options := make([]OrderItemOption, len(model.Options))
	for i, opt := range model.Options {
		options[i] = OrderItemOption{
			ID:              opt.ID,
			OrderItemID:     opt.OrderItemID,
			CategoryID:      opt.CategoryID,
			OptionID:        opt.OptionID,
			OptionName:      opt.OptionName,
			CategoryName:    opt.CategoryName,
			PriceAdjustment: opt.PriceAdjustment,
		}
	}

	return &OrderItem{
		id:                model.ID,
		productID:         model.ProductID,
		quantity:          model.Quantity,
		price:             model.Price,
		totalPrice:        model.TotalPrice,
		productName:       model.ProductName,
		productDescription: model.ProductDescription,
		productImageURL:   model.ProductImageURL,
		options:           options,
	}
}
