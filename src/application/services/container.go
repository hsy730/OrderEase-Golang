package services

// ServiceContainer 服务容器
// 注意：实例化请使用 wire.InitializeServiceContainer()
type ServiceContainer struct {
	OrderService     *OrderService
	ProductService   *ProductService
	ShopService      *ShopService
	UserService      *UserService
	TempTokenService *TempTokenService
}

// NewServiceContainer 创建服务容器（由 Wire 调用）
func NewServiceContainer(
	orderService *OrderService,
	productService *ProductService,
	shopService *ShopService,
	userService *UserService,
	tempTokenService *TempTokenService,
) *ServiceContainer {
	return &ServiceContainer{
		OrderService:     orderService,
		ProductService:   productService,
		ShopService:      shopService,
		UserService:      userService,
		TempTokenService: tempTokenService,
	}
}
