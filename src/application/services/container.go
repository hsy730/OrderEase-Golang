package services

import (
	"gorm.io/gorm"
	"orderease/infrastructure/repositories"
)

type ServiceContainer struct {
	OrderService  *OrderService
	ProductService *ProductService
	ShopService   *ShopService
	UserService   *UserService
}

func NewServiceContainer(db *gorm.DB) *ServiceContainer {
	orderRepo := repositories.NewOrderRepository(db)
	orderItemRepo := repositories.NewOrderItemRepository(db)
	orderItemOptionRepo := repositories.NewOrderItemOptionRepository(db)
	orderStatusLogRepo := repositories.NewOrderStatusLogRepository(db)

	productRepo := repositories.NewProductRepository(db)
	productCategoryRepo := repositories.NewProductOptionCategoryRepository(db)
	productOptionRepo := repositories.NewProductOptionRepository(db)
	productTagRepo := repositories.NewProductTagRepository(db)

	shopRepo := repositories.NewShopRepository(db)
	tagRepo := repositories.NewTagRepository(db)

	userRepo := repositories.NewUserRepository(db)

	orderService := NewOrderService(
		orderRepo,
		orderItemRepo,
		orderItemOptionRepo,
		orderStatusLogRepo,
		productRepo,
		productOptionRepo,
		productCategoryRepo,
		db,
	)

	productService := NewProductService(
		productRepo,
		productCategoryRepo,
		productOptionRepo,
		productTagRepo,
		productRepo,
		db,
	)

	shopService := NewShopService(
		shopRepo,
		tagRepo,
		productRepo,
		db,
	)

	userService := NewUserService(
		userRepo,
		db,
	)

	return &ServiceContainer{
		OrderService:  orderService,
		ProductService: productService,
		ShopService:   shopService,
		UserService:   userService,
	}
}
