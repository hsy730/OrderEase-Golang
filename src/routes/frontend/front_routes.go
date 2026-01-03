package frontend

import (
	"orderease/config"
	"orderease/handlers"
	"orderease/middleware"

	"github.com/gin-gonic/gin"
)

// SetupFrontRoutes 配置前端需要认证的路由
func SetupFrontRoutes(r *gin.Engine, h *handlers.Handler) {
	// 获取基础路径
	basePath := config.AppConfig.Server.BasePath

	// 需要认证的路由组
	protected := r.Group(basePath)
	protected.Use(middleware.RateLimitMiddleware())
	protected.Use(middleware.FrontendAuthMiddleware())

	// 产品相关路由
	setupProductRoutes(protected, h)

	// 订单相关路由
	setupOrderRoutes(protected, h)

	// 标签相关路由
	setupTagRoutes(protected, h)

	// 店铺相关路由
	setupShopRoutes(protected, h)
}

// setupProductRoutes 配置产品相关路由
func setupProductRoutes(group *gin.RouterGroup, h *handlers.Handler) {
	// @Summary 获取产品图片
	// @Description 获取指定产品的图片
	// @Tags 产品
	// @Accept json
	// @Produce json
	// @Param productId query string true "产品ID"
	// @Success 200 {object} Response
	// @Router /api/product/image [get]
	group.GET("/product/image", h.GetProductImage)

	// @Summary 获取产品列表
	// @Description 获取产品列表，支持分页和筛选
	// @Tags 产品
	// @Accept json
	// @Produce json
	// @Param page query int false "页码"
	// @Param pageSize query int false "每页数量"
	// @Param categoryId query string false "分类ID"
	// @Param tagId query string false "标签ID"
	// @Success 200 {object} Response
	// @Router /api/product/list [get]
	group.GET("/product/list", h.GetProducts)

	// @Summary 获取产品详情
	// @Description 获取指定产品的详细信息
	// @Tags 产品
	// @Accept json
	// @Produce json
	// @Param productId query string true "产品ID"
	// @Success 200 {object} Response
	// @Router /api/product/detail [get]
	group.GET("/product/detail", h.GetProduct)
}

// setupOrderRoutes 配置订单相关路由
func setupOrderRoutes(group *gin.RouterGroup, h *handlers.Handler) {
	// @Summary 创建订单
	// @Description 创建新订单
	// @Tags 订单
	// @Accept json
	// @Produce json
	// @Param order body CreateOrderRequest true "订单信息"
	// @Success 200 {object} Response
	// @Router /api/order/create [post]
	group.POST("/order/create", h.CreateOrder)

	// @Summary 获取用户订单列表
	// @Description 获取当前用户的订单列表
	// @Tags 订单
	// @Accept json
	// @Produce json
	// @Param page query int false "页码"
	// @Param pageSize query int false "每页数量"
	// @Param status query string false "订单状态"
	// @Success 200 {object} Response
	// @Router /api/order/user/list [get]
	group.GET("/order/user/list", h.GetOrdersByUser)

	// @Summary 获取订单详情
	// @Description 获取指定订单的详细信息
	// @Tags 订单
	// @Accept json
	// @Produce json
	// @Param orderId query string true "订单ID"
	// @Success 200 {object} Response
	// @Router /api/order/detail [get]
	group.GET("/order/detail", h.GetOrder)

	// @Summary 删除订单
	// @Description 删除指定的订单
	// @Tags 订单
	// @Accept json
	// @Produce json
	// @Param orderId query string true "订单ID"
	// @Success 200 {object} Response
	// @Router /api/order/delete [delete]
	group.DELETE("/order/delete", h.DeleteOrder)
}

// setupTagRoutes 配置标签相关路由
func setupTagRoutes(group *gin.RouterGroup, h *handlers.Handler) {
	// @Summary 获取标签列表
	// @Description 获取标签列表
	// @Tags 标签
	// @Accept json
	// @Produce json
	// @Success 200 {object} Response
	// @Router /api/tag/list [get]
	group.GET("/tag/list", h.GetTagsForFront)

	// @Summary 获取标签详情
	// @Description 获取指定标签的详细信息
	// @Tags 标签
	// @Accept json
	// @Produce json
	// @Param tagId query string true "标签ID"
	// @Success 200 {object} Response
	// @Router /api/tag/detail [get]
	group.GET("/tag/detail", h.GetTag)

	// @Summary 获取标签已绑定的商品
	// @Description 获取指定标签已绑定的商品列表
	// @Tags 标签
	// @Accept json
	// @Produce json
	// @Param tagId query string true "标签ID"
	// @Success 200 {object} Response
	// @Router /api/tag/bound-products [get]
	group.GET("/tag/bound-products", h.GetTagBoundProducts)

	// @Summary 获取店铺标签列表
	// @Description 获取指定店铺的标签列表
	// @Tags 标签
	// @Accept json
	// @Produce json
	// @Param shopId path string true "店铺ID"
	// @Success 200 {object} Response
	// @Router /api/shop/{shopId}/tags [get]
	group.GET("/shop/:shopId/tags", h.GetShopTags)
}

// setupShopRoutes 配置店铺相关路由
func setupShopRoutes(group *gin.RouterGroup, h *handlers.Handler) {
	// @Summary 获取店铺详情
	// @Description 获取指定店铺的详细信息
	// @Tags 店铺
	// @Accept json
	// @Produce json
	// @Param shopId query string true "店铺ID"
	// @Success 200 {object} Response
	// @Router /api/shop/detail [get]
	group.GET("/shop/detail", h.GetShopInfo)

	// @Summary 获取店铺图片
	// @Description 获取指定店铺的图片
	// @Tags 店铺
	// @Accept json
	// @Produce json
	// @Param shopId query string true "店铺ID"
	// @Success 200 {object} Response
	// @Router /api/shop/image [get]
	group.GET("/shop/image", h.GetShopImage)
}

