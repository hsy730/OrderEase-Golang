package services

// ServiceContainer 服务容器
// 注意：实例化请使用 wire.InitializeServiceContainer()
type ServiceContainer struct {
	OrderService   *OrderService
	ProductService *ProductService
	ShopService    *ShopService
	UserService    *UserService
}
