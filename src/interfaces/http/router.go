package http

import (
	"orderease/application/services"
	"orderease/config"
	imiddleware "orderease/interfaces/middleware"
	"orderease/middleware"
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
	exportHandler  *ExportHandler
	importHandler  *ImportHandler
}

func NewRouter(db *gorm.DB, services *services.ServiceContainer) *Router {
	return &Router{
		orderHandler:   NewOrderHandler(services.OrderService, services.ShopService),
		productHandler: NewProductHandler(services.ProductService, services.ShopService),
		shopHandler:    NewShopHandler(services.ShopService),
		userHandler:    NewUserHandler(services.UserService),
		authHandler:    NewAuthHandler(db, services.ShopService, services.UserService, services.TempTokenService),
		exportHandler:  NewExportHandler(db),
		importHandler:  NewImportHandler(db),
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
		noAuth.GET("/product/image", r.productHandler.GetProductImage)
		noAuth.GET("/order/list", r.orderHandler.GetOrders)
		noAuth.GET("/order/detail", r.orderHandler.GetOrder)
		noAuth.GET("/order/user/list", r.orderHandler.GetOrdersByUser)
		noAuth.GET("/tag/list", r.shopHandler.GetShopTags)
	}

	// 用户认证相关公开路由
	user := api.Group("/user")
	{
		user.POST("/login", r.authHandler.FrontendUserLogin)
		user.POST("/register", r.authHandler.FrontendUserRegister)
		user.GET("/check-username", r.userHandler.CheckFrontendUsernameExists)
	}
}

// 店主路由（需要认证）
func (r *Router) setupShopOwnerRoutes(api *gin.RouterGroup) {
	shopOwner := api.Group("/shopOwner")
	shopOwner.Use(imiddleware.AuthMiddleware())
	{
		// 认证管理
		shopOwner.POST("/logout", r.authHandler.Logout)
		shopOwner.POST("/change-password", r.authHandler.ChangePassword)

		// 店铺管理
		// shopOwner.POST("/shop/create", r.shopHandler.CreateShop)
		shopOwner.PUT("/shop/update", r.shopHandler.UpdateShop)
		shopOwner.DELETE("/shop/delete", r.shopHandler.DeleteShop)
		shopOwner.GET("/shop/detail", r.shopHandler.GetShopInfo)
		shopOwner.PUT("/shop/update-order-status-flow", r.shopHandler.UpdateOrderStatusFlow)
		shopOwner.GET("/shop/temp-token", r.authHandler.GetShopTempToken)
		shopOwner.GET("/shop/image", r.shopHandler.GetShopImage)
		shopOwner.POST("/shop/upload-image", r.shopHandler.UploadShopImage)

		// 商品管理
		shopOwner.POST("/product/create", r.productHandler.CreateProduct)
		shopOwner.PUT("/product/update", r.productHandler.UpdateProduct)
		shopOwner.DELETE("/product/delete", r.productHandler.DeleteProduct)
		shopOwner.PUT("/product/status", r.productHandler.UpdateProductStatus)
		shopOwner.PUT("/product/toggle-status", r.productHandler.ToggleProductStatus)
		shopOwner.GET("/product/detail", r.productHandler.GetProduct)
		shopOwner.GET("/product/list", r.productHandler.GetProducts)
		shopOwner.GET("/product/image", r.productHandler.GetProductImage)
		shopOwner.POST("/product/upload-image", r.productHandler.UploadProductImage)

		// 订单管理
		shopOwner.POST("/order/create", r.orderHandler.CreateOrder)
		shopOwner.PUT("/order/update", r.orderHandler.UpdateOrder)
		shopOwner.PUT("/order/status", r.orderHandler.UpdateOrderStatus)
		shopOwner.PUT("/order/toggle-status", r.orderHandler.ToggleOrderStatus)
		shopOwner.DELETE("/order/delete", r.orderHandler.DeleteOrder)
		shopOwner.GET("/order/detail", r.orderHandler.GetOrder)
		shopOwner.GET("/order/list", r.orderHandler.GetOrders)
		shopOwner.GET("/order/user-orders", r.orderHandler.GetOrdersByUser)
		shopOwner.GET("/order/unfinished", r.orderHandler.GetUnfinishedOrders)
		shopOwner.POST("/order/search", r.orderHandler.SearchOrders)
		shopOwner.POST("/order/advance-search", r.orderHandler.GetAdvanceSearchOrders)
		shopOwner.GET("/order/status-flow", r.orderHandler.GetOrderStatusFlow)
		shopOwner.GET("/order/user/list", r.orderHandler.GetOrdersByUser)

		// 标签管理
		shopOwner.POST("/tag/create", r.shopHandler.CreateTag)
		shopOwner.PUT("/tag/update", r.shopHandler.UpdateTag)
		shopOwner.DELETE("/tag/delete", r.shopHandler.DeleteTag)
		shopOwner.GET("/tag/list", r.shopHandler.GetShopTags)
		shopOwner.GET("/tag/detail", r.shopHandler.GetTag)
		shopOwner.GET("/tag/bound-tags", r.shopHandler.GetBoundTags)
		shopOwner.GET("/tag/unbound-tags", r.shopHandler.GetUnboundTags)
		shopOwner.POST("/tag/batch-tag", r.shopHandler.BatchTagProducts)
		shopOwner.DELETE("/tag/batch-untag", r.shopHandler.BatchUntagProducts)
		shopOwner.POST("/tag/batch-tag-product", r.shopHandler.BatchTagProduct)
		shopOwner.GET("/tag/bound-products", r.shopHandler.GetTagBoundProducts)
		shopOwner.GET("/tag/unbound-products", r.shopHandler.GetUnboundProductsForTag)
		shopOwner.GET("/tag/unbound-list", r.shopHandler.GetUnboundTagsList)
		shopOwner.GET("/tag/online-products", r.shopHandler.GetTagOnlineProducts)

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
	admin.Use(imiddleware.AuthMiddleware(), imiddleware.AdminMiddleware())
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
		admin.GET("/shop/image", r.shopHandler.GetShopImage)
		admin.POST("/shop/upload-image", r.shopHandler.UploadShopImage)

		// 商品管理
		admin.POST("/product/create", r.productHandler.CreateProduct)
		admin.PUT("/product/update", r.productHandler.UpdateProduct)
		admin.DELETE("/product/delete", r.productHandler.DeleteProduct)
		admin.PUT("/product/status", r.productHandler.UpdateProductStatus)
		admin.PUT("/product/toggle-status", r.productHandler.ToggleProductStatus)
		admin.GET("/product/list", r.productHandler.GetProducts)
		admin.GET("/product/detail", r.productHandler.GetProduct)
		admin.GET("/product/image", r.productHandler.GetProductImage)
		admin.POST("/product/upload-image", r.productHandler.UploadProductImage)

		// 订单管理
		admin.POST("/order/create", r.orderHandler.CreateOrder)
		admin.PUT("/order/update", r.orderHandler.UpdateOrder)
		admin.PUT("/order/status", r.orderHandler.UpdateOrderStatus)
		admin.PUT("/order/toggle-status", r.orderHandler.ToggleOrderStatus)
		admin.DELETE("/order/delete", r.orderHandler.DeleteOrder)
		admin.GET("/order/list", r.orderHandler.GetOrders)
		admin.GET("/order/detail", r.orderHandler.GetOrder)
		admin.POST("/order/search", r.orderHandler.SearchOrders)
		admin.POST("/order/advance-search", r.orderHandler.GetAdvanceSearchOrders)
		admin.GET("/order/status-flow", r.orderHandler.GetOrderStatusFlow)
		admin.GET("/order/user/list", r.orderHandler.GetOrdersByUser)

		// 标签管理
		admin.POST("/tag/create", r.shopHandler.CreateTag)
		admin.PUT("/tag/update", r.shopHandler.UpdateTag)
		admin.DELETE("/tag/delete", r.shopHandler.DeleteTag)
		admin.GET("/tag/list", r.shopHandler.GetShopTags)
		admin.GET("/tag/detail", r.shopHandler.GetTag)
		admin.GET("/tag/bound-tags", r.shopHandler.GetBoundTags)
		admin.GET("/tag/unbound-tags", r.shopHandler.GetUnboundTags)
		admin.POST("/tag/batch-tag", r.shopHandler.BatchTagProducts)
		admin.DELETE("/tag/batch-untag", r.shopHandler.BatchUntagProducts)
		admin.POST("/tag/batch-tag-product", r.shopHandler.BatchTagProduct)
		admin.GET("/tag/bound-products", r.shopHandler.GetTagBoundProducts)
		admin.GET("/tag/unbound-products", r.shopHandler.GetUnboundProductsForTag)
		admin.GET("/tag/unbound-list", r.shopHandler.GetUnboundTagsList)
		admin.GET("/tag/online-products", r.shopHandler.GetTagOnlineProducts)

		// 数据管理
		admin.GET("/data/export", r.exportHandler.ExportData)
		admin.POST("/data/import", r.importHandler.ImportData)

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
func (r *Router) setupFrontendRoutes(frontend *gin.RouterGroup) {
	// 前端路由需要认证和限流
	frontend.Use(middleware.RateLimitMiddleware())
	frontend.Use(middleware.FrontendAuthMiddleware())

	{
		// 店铺管理
		frontend.GET("/shop/detail", r.shopHandler.GetShopInfo)
		frontend.GET("/shop/list", r.shopHandler.GetShopList)
		frontend.GET("/shop/image", r.shopHandler.GetShopImage)
		frontend.GET("/shop/:shop_id/tags", r.shopHandler.GetShopTags)

		// 商品管理
		frontend.GET("/product/list", r.productHandler.GetProducts)
		frontend.GET("/product/detail", r.productHandler.GetProduct)
		frontend.GET("/product/image", r.productHandler.GetProductImage)

		// 订单管理
		frontend.POST("/order/create", r.orderHandler.CreateOrder)
		frontend.GET("/order/list", r.orderHandler.GetOrders)
		frontend.GET("/order/detail", r.orderHandler.GetOrder)
		frontend.DELETE("/order/delete", r.orderHandler.DeleteOrder)
		frontend.GET("/order/user/list", r.orderHandler.GetOrdersByUser)

		// 标签管理
		frontend.GET("/tag/list", r.shopHandler.GetShopTags)
		frontend.GET("/tag/detail", r.shopHandler.GetTag)
		frontend.GET("/tag/bound-products", r.shopHandler.GetTagBoundProducts)
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
