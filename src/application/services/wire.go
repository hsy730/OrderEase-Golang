//go:build wireinject
// +build wireinject

package services

import (
	"github.com/google/wire"
	"gorm.io/gorm"
	"orderease/domain/order"
	"orderease/domain/product"
	"orderease/domain/shop"
	"orderease/domain/user"
	"orderease/infrastructure/repositories"
)

// InitializeServiceContainer 初始化服务容器（Wire 生成）
func InitializeServiceContainer(db *gorm.DB) (*ServiceContainer, error) {
	wire.Build(
		// Repository 层
		repositories.NewOrderRepository,
		repositories.NewOrderItemRepository,
		repositories.NewOrderItemOptionRepository,
		repositories.NewOrderStatusLogRepository,
		repositories.NewProductRepository,
		repositories.NewProductOptionCategoryRepository,
		repositories.NewProductOptionRepository,
		repositories.NewProductTagRepository,
		repositories.NewShopRepository,
		repositories.NewTagRepository,
		repositories.NewUserRepository,

		// Service 层
		NewOrderService,
		NewProductService,
		NewShopService,
		NewUserService,
		NewTempTokenService,

		// Container
		NewServiceContainer,
	)
	return &ServiceContainer{}, nil
}

// Provider Sets - 按 Provider Sets 组织依赖

// RepositoryProviderSet 所有仓储的 Provider Set
var RepositoryProviderSet = wire.NewSet(
	// Order 仓储
	wire.Bind(new(order.OrderRepository), new(*repositories.OrderRepository)),
	repositories.NewOrderRepository,

	wire.Bind(new(order.OrderItemRepository), new(*repositories.OrderItemRepository)),
	repositories.NewOrderItemRepository,

	wire.Bind(new(order.OrderItemOptionRepository), new(*repositories.OrderItemOptionRepository)),
	repositories.NewOrderItemOptionRepository,

	wire.Bind(new(order.OrderStatusLogRepository), new(*repositories.OrderStatusLogRepository)),
	repositories.NewOrderStatusLogRepository,

	// Product 仓储
	wire.Bind(new(product.ProductRepository), new(*repositories.ProductRepository)),
	repositories.NewProductRepository,

	wire.Bind(new(product.ProductOptionCategoryRepository), new(*repositories.ProductOptionCategoryRepository)),
	repositories.NewProductOptionCategoryRepository,

	wire.Bind(new(product.ProductOptionRepository), new(*repositories.ProductOptionRepository)),
	repositories.NewProductOptionRepository,

	wire.Bind(new(product.ProductTagRepository), new(*repositories.ProductTagRepository)),
	repositories.NewProductTagRepository,

	// Shop 仓储
	wire.Bind(new(shop.ShopRepository), new(*repositories.ShopRepository)),
	repositories.NewShopRepository,

	wire.Bind(new(shop.TagRepository), new(*repositories.TagRepository)),
	repositories.NewTagRepository,

	// User 仓储
	wire.Bind(new(user.UserRepository), new(*repositories.UserRepository)),
	repositories.NewUserRepository,
)

// ServiceProviderSet 所有服务的 Provider Set
var ServiceProviderSet = wire.NewSet(
	NewOrderService,
	NewProductService,
	NewShopService,
	NewUserService,
	NewTempTokenService,
)
