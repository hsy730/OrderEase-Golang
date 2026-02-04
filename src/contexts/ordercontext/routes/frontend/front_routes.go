package frontend

import (
	"orderease/config"
	ordercontextHandlers "orderease/contexts/ordercontext/application/handlers"
	"orderease/middleware"

	"github.com/gin-gonic/gin"
)

func SetupFrontRoutes(r *gin.Engine, h *ordercontextHandlers.Handler) {
	// 获取基础路径
	basePath := config.AppConfig.Server.BasePath

	// 需要认证的路由组
	protected := r.Group(basePath)
	protected.Use(middleware.RateLimitMiddleware())
	protected.Use(middleware.FrontendAuthMiddleware())

	{
		protected.GET("/product/image", h.GetProductImage)
		protected.GET("/product/list", h.GetProducts)
		protected.GET("/product/detail", h.GetProduct)
		protected.POST("/order/create", h.CreateOrder)
		protected.GET("/order/user/list", h.GetOrdersByUser)
		protected.GET("/order/detail", h.GetOrder)
		protected.DELETE("/order/delete", h.DeleteOrder)
		// protected.POST("/order/pay", h.PayOrder)
		protected.GET("/tag/list", h.GetTagsForFront)
		protected.GET("/tag/detail", h.GetTag)
		protected.GET("/tag/bound-products", h.GetTagBoundProducts) // 获取标签已绑定的商品列表
		protected.GET("/shop/:shopId/tags", h.GetShopTags)
		protected.GET("/shop/detail", h.GetShopInfo)
		protected.GET("/shop/image", h.GetShopImage)
	}
}
