package http

import (
	"orderease/application/services"
	"orderease/config"
	"orderease/interfaces/middleware"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Router struct {
	orderHandler   *OrderHandler
	productHandler *ProductHandler
	shopHandler    *ShopHandler
	userHandler    *UserHandler
	authHandler    *AuthHandler
}

func NewRouter(db *gorm.DB, services *services.ServiceContainer) *Router {
	return &Router{
		orderHandler:   NewOrderHandler(services.OrderService, services.ShopService),
		productHandler: NewProductHandler(services.ProductService, services.ShopService),
		shopHandler:    NewShopHandler(services.ShopService),
		userHandler:    NewUserHandler(services.UserService),
		authHandler:    NewAuthHandler(db, services.ShopService, services.UserService),
	}
}

func (r *Router) SetupRoutes(app *gin.Engine) {
	basePath := config.AppConfig.Server.BasePath
	api := app.Group(basePath)

	r.setupAuthRoutes(api)
	r.setupNoAuthRoutes(api)
	r.setupShopOwnerRoutes(api)
	r.setupAdminRoutes(api)
	r.setupFrontendRoutes(api)
}

// 认证路由（登录、登出、刷新令牌）
func (r *Router) setupAuthRoutes(api *gin.RouterGroup) {
	// 登录相关
	api.POST("/login", r.authHandler.Login)
	api.POST("/shop/refresh-token", r.authHandler.RefreshShopToken)
	api.POST("/admin/refresh-token", r.authHandler.RefreshAdminToken)
	api.POST("/shop/temp-login", r.authHandler.TempTokenLogin)
}

// 无认证路由（公开查询）
func (r *Router) setupNoAuthRoutes(api *gin.RouterGroup) {
	// 公开查询接口
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
		noAuth.GET("/tag/list", r.shopHandler.GetShopTags)
	}

	// 用户认证相关公开路由
	user := api.Group("/user")
	{
		user.POST("/login", r.authHandler.Login)
		user.POST("/register", r.authHandler.Register)
		user.GET("/check-username", r.userHandler.CheckUsernameExists)
	}
}

// 店主路由（需要认证）
func (r *Router) setupShopOwnerRoutes(api *gin.RouterGroup) {
	shopOwner := api.Group("/shopOwner")
	shopOwner.Use(middleware.AuthMiddleware())
	{
		// 认证管理
		shopOwner.POST("/logout", r.authHandler.Logout)
		shopOwner.POST("/change-password", r.authHandler.ChangePassword)

		// 店铺管理
		shopOwner.POST("/shop/create", r.shopHandler.CreateShop)
		shopOwner.PUT("/shop/update", r.shopHandler.UpdateShop)
		shopOwner.DELETE("/shop/delete", r.shopHandler.DeleteShop)
		shopOwner.GET("/shop/detail", r.shopHandler.GetShopInfo)
		shopOwner.PUT("/shop/update-order-status-flow", r.shopHandler.UpdateOrderStatusFlow)
		shopOwner.GET("/shop/temp-token", r.authHandler.GetShopTempToken)

		// 商品管理
		shopOwner.POST("/product/create", r.productHandler.CreateProduct)
		shopOwner.PUT("/product/update", r.productHandler.UpdateProduct)
		shopOwner.DELETE("/product/delete", r.productHandler.DeleteProduct)
		shopOwner.PUT("/product/status", r.productHandler.UpdateProductStatus)
		shopOwner.GET("/product/detail", r.productHandler.GetProduct)
		shopOwner.GET("/product/list", r.productHandler.GetProducts)

		// 订单管理
		shopOwner.POST("/order/create", r.orderHandler.CreateOrder)
		shopOwner.PUT("/order/status", r.orderHandler.UpdateOrderStatus)
		shopOwner.DELETE("/order/delete", r.orderHandler.DeleteOrder)
		shopOwner.GET("/order/detail", r.orderHandler.GetOrder)
		shopOwner.GET("/order/list", r.orderHandler.GetOrders)
		shopOwner.GET("/order/user-orders", r.orderHandler.GetOrdersByUser)
		shopOwner.GET("/order/unfinished", r.orderHandler.GetUnfinishedOrders)
		shopOwner.POST("/order/search", r.orderHandler.SearchOrders)

		// 标签管理
		shopOwner.POST("/tag/create", r.shopHandler.CreateTag)
		shopOwner.PUT("/tag/update", r.shopHandler.UpdateTag)
		shopOwner.DELETE("/tag/delete", r.shopHandler.DeleteTag)
		shopOwner.GET("/tag/list", r.shopHandler.GetShopTags)
		shopOwner.GET("/tag/detail", r.shopHandler.GetTag)

		// 用户管理
		shopOwner.POST("/user/create", r.userHandler.CreateUser)
		shopOwner.PUT("/user/update", r.userHandler.UpdateUser)
		shopOwner.DELETE("/user/delete", r.userHandler.DeleteUser)
		shopOwner.GET("/user/detail", r.userHandler.GetUser)
		shopOwner.GET("/user/list", r.userHandler.GetUsers)
		shopOwner.GET("/user/simple-list", r.userHandler.GetUserSimpleList)
	}
}

// 管理员路由（需要管理员权限）
func (r *Router) setupAdminRoutes(api *gin.RouterGroup) {
	admin := api.Group("/admin")
	admin.Use(middleware.AuthMiddleware(), middleware.AdminMiddleware())
	{
		// 认证管理
		admin.POST("/logout", r.authHandler.Logout)
		admin.POST("/change-password", r.authHandler.ChangePassword)

		// 店铺管理
		admin.POST("/shop/create", r.shopHandler.CreateShop)
		admin.PUT("/shop/update", r.shopHandler.UpdateShop)
		admin.DELETE("/shop/delete", r.shopHandler.DeleteShop)
		admin.GET("/shop/list", r.shopHandler.GetShopList)
		admin.GET("/shop/detail", r.shopHandler.GetShopInfo)
		admin.PUT("/shop/update-order-status-flow", r.shopHandler.UpdateOrderStatusFlow)
		admin.GET("/shop/check-name", r.shopHandler.CheckShopNameExists)

		// 商品管理
		admin.POST("/product/create", r.productHandler.CreateProduct)
		admin.PUT("/product/update", r.productHandler.UpdateProduct)
		admin.DELETE("/product/delete", r.productHandler.DeleteProduct)
		admin.PUT("/product/status", r.productHandler.UpdateProductStatus)
		admin.GET("/product/list", r.productHandler.GetProducts)
		admin.GET("/product/detail", r.productHandler.GetProduct)

		// 订单管理
		admin.POST("/order/create", r.orderHandler.CreateOrder)
		admin.PUT("/order/status", r.orderHandler.UpdateOrderStatus)
		admin.DELETE("/order/delete", r.orderHandler.DeleteOrder)
		admin.GET("/order/list", r.orderHandler.GetOrders)
		admin.GET("/order/detail", r.orderHandler.GetOrder)
		admin.POST("/order/search", r.orderHandler.SearchOrders)

		// 标签管理
		admin.POST("/tag/create", r.shopHandler.CreateTag)
		admin.PUT("/tag/update", r.shopHandler.UpdateTag)
		admin.DELETE("/tag/delete", r.shopHandler.DeleteTag)
		admin.GET("/tag/list", r.shopHandler.GetShopTags)
		admin.GET("/tag/detail", r.shopHandler.GetTag)

		// 用户管理
		admin.POST("/user/create", r.userHandler.CreateUser)
		admin.PUT("/user/update", r.userHandler.UpdateUser)
		admin.DELETE("/user/delete", r.userHandler.DeleteUser)
		admin.GET("/user/list", r.userHandler.GetUsers)
		admin.GET("/user/simple-list", r.userHandler.GetUserSimpleList)
		admin.GET("/user/detail", r.userHandler.GetUser)
	}
}

// 前端路由（面向最终用户）
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
		frontend.GET("/tag/list", r.shopHandler.GetShopTags)
	}
}

// Helper function for path parameters
func getShopIDParam(c *gin.Context) (string, bool) {
	shopID := c.Param("shop_id")
	if shopID == "" {
		shopID = c.Query("shop_id")
	}
	return shopID, shopID != ""
}

// Helper function for int path parameters
func getIntParam(c *gin.Context, key string) (int, error) {
	return strconv.Atoi(c.Param(key))
}
