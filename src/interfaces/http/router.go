package http

import (
	"orderease/application/services"
	"orderease/interfaces/middleware"

	"github.com/gin-gonic/gin"
)

type Router struct {
	orderHandler  *OrderHandler
	productHandler *ProductHandler
	shopHandler   *ShopHandler
}

func NewRouter(services *services.ServiceContainer) *Router {
	return &Router{
		orderHandler:  NewOrderHandler(services.OrderService, services.ShopService),
		productHandler: NewProductHandler(services.ProductService, services.ShopService),
		shopHandler:   NewShopHandler(services.ShopService),
	}
}

func (r *Router) SetupRoutes(app *gin.Engine) {
	api := app.Group("/api")

	r.setupNoAuthRoutes(api)
	r.setupShopOwnerRoutes(api)
	r.setupAdminRoutes(api)
	r.setupFrontendRoutes(api)
}

func (r *Router) setupNoAuthRoutes(api *gin.RouterGroup) {
	noAuth := api.Group("/no-auth")
	{
		noAuth.GET("/shop/info", r.shopHandler.GetShopInfo)
		noAuth.GET("/shop/list", r.shopHandler.GetShopList)
		noAuth.GET("/shop/check-name", r.shopHandler.CheckShopNameExists)
		noAuth.GET("/shop/:shop_id/tags", r.shopHandler.GetShopTags)
		noAuth.GET("/product/list", r.productHandler.GetProducts)
		noAuth.GET("/product/detail", r.productHandler.GetProduct)
		noAuth.GET("/order/list", r.orderHandler.GetOrders)
		noAuth.GET("/order/detail", r.orderHandler.GetOrder)
	}
}

func (r *Router) setupShopOwnerRoutes(api *gin.RouterGroup) {
	shopOwner := api.Group("/shopOwner")
	shopOwner.Use(middleware.AuthMiddleware())
	{
		shopOwner.POST("/shop/create", r.shopHandler.CreateShop)
		shopOwner.PUT("/shop/update", r.shopHandler.UpdateShop)
		shopOwner.DELETE("/shop/delete", r.shopHandler.DeleteShop)
		shopOwner.PUT("/shop/update-order-status-flow", r.shopHandler.UpdateOrderStatusFlow)

		shopOwner.POST("/product/create", r.productHandler.CreateProduct)
		shopOwner.PUT("/product/update", r.productHandler.UpdateProduct)
		shopOwner.DELETE("/product/delete", r.productHandler.DeleteProduct)
		shopOwner.PUT("/product/status", r.productHandler.UpdateProductStatus)

		shopOwner.POST("/order/create", r.orderHandler.CreateOrder)
		shopOwner.PUT("/order/status", r.orderHandler.UpdateOrderStatus)
		shopOwner.DELETE("/order/delete", r.orderHandler.DeleteOrder)
		shopOwner.GET("/order/user-orders", r.orderHandler.GetOrdersByUser)
		shopOwner.GET("/order/unfinished", r.orderHandler.GetUnfinishedOrders)
		shopOwner.POST("/order/search", r.orderHandler.SearchOrders)
	}
}

func (r *Router) setupAdminRoutes(api *gin.RouterGroup) {
	admin := api.Group("/admin")
	admin.Use(middleware.AuthMiddleware(), middleware.AdminMiddleware())
	{
		admin.GET("/shop/list", r.shopHandler.GetShopList)
		admin.GET("/product/list", r.productHandler.GetProducts)
		admin.GET("/order/list", r.orderHandler.GetOrders)
	}
}

func (r *Router) setupFrontendRoutes(api *gin.RouterGroup) {
	frontend := api.Group("/front")
	{
		frontend.GET("/shop/info", r.shopHandler.GetShopInfo)
		frontend.GET("/shop/list", r.shopHandler.GetShopList)
		frontend.GET("/shop/:shop_id/tags", r.shopHandler.GetShopTags)
		frontend.GET("/product/list", r.productHandler.GetProducts)
		frontend.GET("/product/detail", r.productHandler.GetProduct)
		frontend.GET("/order/list", r.orderHandler.GetOrders)
		frontend.GET("/order/detail", r.orderHandler.GetOrder)
	}
}
