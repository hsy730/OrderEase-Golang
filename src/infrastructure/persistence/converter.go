package persistence

import (
	"orderease/domain/order"
	"orderease/domain/product"
	"orderease/domain/shared"
	"orderease/domain/shop"
	"orderease/domain/user"
	"orderease/models"

	"github.com/bwmarrin/snowflake"
)

func OrderToDomain(m models.Order) *order.Order {
	items := make([]order.OrderItem, len(m.Items))
	for i, item := range m.Items {
		items[i] = *OrderItemToDomain(item)
	}

	return &order.Order{
		ID:         shared.ID(m.ID),
		UserID:     shared.ID(m.UserID),
		ShopID:     uint64(m.ShopID),
		TotalPrice: shared.Price(m.TotalPrice),
		Status:     order.OrderStatus(m.Status),
		Remark:     m.Remark,
		CreatedAt:  m.CreatedAt,
		UpdatedAt:  m.UpdatedAt,
		Items:      items,
	}
}

func OrderToModel(d *order.Order) *models.Order {
	items := make([]models.OrderItem, len(d.Items))
	for i, item := range d.Items {
		items[i] = OrderItemToModel(item)
	}

	return &models.Order{
		ID:         d.ID.Value(),
		UserID:     d.UserID.Value(),
		ShopID:     snowflake.ID(d.ShopID),
		TotalPrice: models.Price(d.TotalPrice),
		Status:     int(d.Status),
		Remark:     d.Remark,
		CreatedAt:  d.CreatedAt,
		UpdatedAt:  d.UpdatedAt,
		Items:      items,
	}
}

func OrderItemToDomain(m models.OrderItem) *order.OrderItem {
	options := make([]order.OrderItemOption, len(m.Options))
	for i, opt := range m.Options {
		options[i] = *OrderItemOptionToDomain(opt)
	}

	return &order.OrderItem{
		ID:                 shared.ID(m.ID),
		OrderID:            shared.ID(m.OrderID),
		ProductID:          shared.ID(m.ProductID),
		Quantity:           m.Quantity,
		Price:              shared.Price(m.Price),
		TotalPrice:         shared.Price(m.TotalPrice),
		ProductName:        m.ProductName,
		ProductDescription: m.ProductDescription,
		ProductImageURL:    m.ProductImageURL,
		Options:            options,
	}
}

func OrderItemToModel(d order.OrderItem) models.OrderItem {
	options := make([]models.OrderItemOption, len(d.Options))
	for i, opt := range d.Options {
		options[i] = OrderItemOptionToModel(opt)
	}

	return models.OrderItem{
		ID:                 d.ID.Value(),
		OrderID:            d.OrderID.Value(),
		ProductID:          d.ProductID.Value(),
		Quantity:           d.Quantity,
		Price:              models.Price(d.Price),
		TotalPrice:         models.Price(d.TotalPrice),
		ProductName:        d.ProductName,
		ProductDescription: d.ProductDescription,
		ProductImageURL:    d.ProductImageURL,
		Options:            options,
	}
}

func OrderItemOptionToDomain(m models.OrderItemOption) *order.OrderItemOption {
	return &order.OrderItemOption{
		ID:              shared.ID(m.ID),
		OrderItemID:     shared.ID(m.OrderItemID),
		CategoryID:      shared.ID(m.CategoryID),
		OptionID:        shared.ID(m.OptionID),
		OptionName:      m.OptionName,
		CategoryName:    m.CategoryName,
		PriceAdjustment: m.PriceAdjustment,
		CreatedAt:       m.CreatedAt,
		UpdatedAt:       m.UpdatedAt,
	}
}

func OrderItemOptionToModel(d order.OrderItemOption) models.OrderItemOption {
	return models.OrderItemOption{
		ID:              d.ID.Value(),
		OrderItemID:     d.OrderItemID.Value(),
		CategoryID:      d.CategoryID.Value(),
		OptionID:        d.OptionID.Value(),
		OptionName:      d.OptionName,
		CategoryName:    d.CategoryName,
		PriceAdjustment: d.PriceAdjustment,
		CreatedAt:       d.CreatedAt,
		UpdatedAt:       d.UpdatedAt,
	}
}

func OrderStatusLogToDomain(m models.OrderStatusLog) *order.OrderStatusLog {
	return &order.OrderStatusLog{
		ID:          shared.ID(m.ID),
		OrderID:     shared.ID(m.OrderID),
		OldStatus:   order.OrderStatus(m.OldStatus),
		NewStatus:   order.OrderStatus(m.NewStatus),
		ChangedTime: m.ChangedTime,
	}
}

func OrderStatusLogToModel(d *order.OrderStatusLog) *models.OrderStatusLog {
	return &models.OrderStatusLog{
		ID:          d.ID.Value(),
		OrderID:     d.OrderID.Value(),
		OldStatus:   int(d.OldStatus),
		NewStatus:   int(d.NewStatus),
		ChangedTime: d.ChangedTime,
	}
}

func ProductToDomain(m models.Product) *product.Product {
	categories := make([]product.ProductOptionCategory, len(m.OptionCategories))
	for i, cat := range m.OptionCategories {
		categories[i] = *ProductOptionCategoryToDomain(cat)
	}

	return &product.Product{
		ID:               shared.ID(m.ID),
		ShopID:           uint64(m.ShopID),
		Name:             m.Name,
		Description:      m.Description,
		Price:            shared.Price(m.Price),
		Stock:            m.Stock,
		ImageURL:         m.ImageURL,
		Status:           product.ProductStatus(m.Status),
		CreatedAt:        m.CreatedAt,
		UpdatedAt:        m.UpdatedAt,
		OptionCategories: categories,
	}
}

func ProductToModel(d *product.Product) *models.Product {
	categories := make([]models.ProductOptionCategory, len(d.OptionCategories))
	for i, cat := range d.OptionCategories {
		categories[i] = ProductOptionCategoryToModel(cat)
	}

	return &models.Product{
		ID:               d.ID.Value(),
		ShopID:           snowflake.ID(d.ShopID),
		Name:             d.Name,
		Description:      d.Description,
		Price:            float64(d.Price),
		Stock:            d.Stock,
		ImageURL:         d.ImageURL,
		Status:           string(d.Status),
		CreatedAt:        d.CreatedAt,
		UpdatedAt:        d.UpdatedAt,
		OptionCategories: categories,
	}
}

func ProductOptionCategoryToDomain(m models.ProductOptionCategory) *product.ProductOptionCategory {
	options := make([]product.ProductOption, len(m.Options))
	for i, opt := range m.Options {
		options[i] = *ProductOptionToDomain(opt)
	}

	return &product.ProductOptionCategory{
		ID:           shared.ID(m.ID),
		ProductID:    shared.ID(m.ProductID),
		Name:         m.Name,
		IsRequired:   m.IsRequired,
		IsMultiple:   m.IsMultiple,
		DisplayOrder: m.DisplayOrder,
		CreatedAt:    m.CreatedAt,
		UpdatedAt:    m.UpdatedAt,
		Options:      options,
	}
}

func ProductOptionCategoryToModel(d product.ProductOptionCategory) models.ProductOptionCategory {
	options := make([]models.ProductOption, len(d.Options))
	for i, opt := range d.Options {
		options[i] = ProductOptionToModel(opt)
	}

	return models.ProductOptionCategory{
		ID:           d.ID.Value(),
		ProductID:    d.ProductID.Value(),
		Name:         d.Name,
		IsRequired:   d.IsRequired,
		IsMultiple:   d.IsMultiple,
		DisplayOrder: d.DisplayOrder,
		CreatedAt:    d.CreatedAt,
		UpdatedAt:    d.UpdatedAt,
		Options:      options,
	}
}

func ProductOptionToDomain(m models.ProductOption) *product.ProductOption {
	return &product.ProductOption{
		ID:              shared.ID(m.ID),
		CategoryID:      shared.ID(m.CategoryID),
		Name:            m.Name,
		PriceAdjustment: m.PriceAdjustment,
		DisplayOrder:    m.DisplayOrder,
		IsDefault:       m.IsDefault,
		CreatedAt:       m.CreatedAt,
		UpdatedAt:       m.UpdatedAt,
	}
}

func ProductOptionToModel(d product.ProductOption) models.ProductOption {
	return models.ProductOption{
		ID:              d.ID.Value(),
		CategoryID:      d.CategoryID.Value(),
		Name:            d.Name,
		PriceAdjustment: d.PriceAdjustment,
		DisplayOrder:    d.DisplayOrder,
		IsDefault:       d.IsDefault,
		CreatedAt:       d.CreatedAt,
		UpdatedAt:       d.UpdatedAt,
	}
}

func ShopToDomain(m models.Shop) *shop.Shop {
	return &shop.Shop{
		ID:            shared.ID(m.ID),
		Name:          m.Name,
		OwnerUsername: m.OwnerUsername,
		OwnerPassword: m.OwnerPassword,
		ContactPhone:  m.ContactPhone,
		ContactEmail:  m.ContactEmail,
		Address:       m.Address,
		ImageURL:      m.ImageURL,
		Description:   m.Description,
		CreatedAt:     m.CreatedAt,
		UpdatedAt:     m.UpdatedAt,
		ValidUntil:    m.ValidUntil,
		Settings:      string(m.Settings),
		OrderStatusFlow: order.OrderStatusFlow{
			Statuses: convertOrderStatuses(m.OrderStatusFlow.Statuses),
		},
	}
}

func ShopToModel(d *shop.Shop) *models.Shop {
	return &models.Shop{
		ID:            snowflake.ID(d.ID),
		Name:          d.Name,
		OwnerUsername: d.OwnerUsername,
		OwnerPassword: d.OwnerPassword,
		ContactPhone:  d.ContactPhone,
		ContactEmail:  d.ContactEmail,
		Address:       d.Address,
		ImageURL:      d.ImageURL,
		Description:   d.Description,
		CreatedAt:     d.CreatedAt,
		UpdatedAt:     d.UpdatedAt,
		ValidUntil:    d.ValidUntil,
		Settings:      []byte(d.Settings),
		OrderStatusFlow: models.OrderStatusFlow{
			Statuses: convertOrderStatusesToModel(d.OrderStatusFlow.Statuses),
		},
	}
}

func convertOrderStatuses(statuses []models.OrderStatus) []order.OrderStatusConfig {
	result := make([]order.OrderStatusConfig, len(statuses))
	for i, s := range statuses {
		actions := make([]order.OrderStatusTransition, len(s.Actions))
		for j, a := range s.Actions {
			actions[j] = order.OrderStatusTransition{
				Name:            a.Name,
				NextStatus:      order.OrderStatus(a.NextStatus),
				NextStatusLabel: a.NextStatusLabel,
			}
		}
		result[i] = order.OrderStatusConfig{
			Value:   order.OrderStatus(s.Value),
			Label:   s.Label,
			Type:    s.Type,
			IsFinal: s.IsFinal,
			Actions: actions,
		}
	}
	return result
}

func convertOrderStatusesToModel(statuses []order.OrderStatusConfig) []models.OrderStatus {
	result := make([]models.OrderStatus, len(statuses))
	for i, s := range statuses {
		actions := make([]models.OrderStatusAction, len(s.Actions))
		for j, a := range s.Actions {
			actions[j] = models.OrderStatusAction{
				Name:            a.Name,
				NextStatus:      int(a.NextStatus),
				NextStatusLabel: a.NextStatusLabel,
			}
		}
		result[i] = models.OrderStatus{
			Value:   int(s.Value),
			Label:   s.Label,
			Type:    s.Type,
			IsFinal: s.IsFinal,
			Actions: actions,
		}
	}
	return result
}

func TagToDomain(m models.Tag) *shop.Tag {
	return &shop.Tag{
		ID:          m.ID,
		ShopID:      shared.ID(m.ShopID),
		Name:        m.Name,
		Description: m.Description,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}
}

func TagToModel(d *shop.Tag) *models.Tag {
	return &models.Tag{
		ID:          d.ID,
		ShopID:      snowflake.ID(d.ShopID),
		Name:        d.Name,
		Description: d.Description,
		CreatedAt:   d.CreatedAt,
		UpdatedAt:   d.UpdatedAt,
	}
}

func UserToDomain(m models.User) *user.User {
	return &user.User{
		ID:        shared.ID(m.ID),
		Name:      m.Name,
		Role:      user.UserRole(m.Role),
		Password:  m.Password,
		Phone:     m.Phone,
		Address:   m.Address,
		Type:      user.UserType(m.Type),
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}
}

func UserToModel(d *user.User) *models.User {
	return &models.User{
		ID:        d.ID.Value(),
		Name:      d.Name,
		Role:      string(d.Role),
		Password:  d.Password,
		Phone:     d.Phone,
		Address:   d.Address,
		Type:      string(d.Type),
		CreatedAt: d.CreatedAt,
		UpdatedAt: d.UpdatedAt,
	}
}
