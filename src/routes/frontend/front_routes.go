package frontend

import (
	"orderease/config"
	"orderease/handlers"
	"orderease/middleware"

	"github.com/gin-gonic/gin"
)

func SetupFrontRoutes(r *gin.Engine, h *handlers.Handler) {
	// 获取基础路径
	basePath := config.AppConfig.Server.BasePath

	// 公开路由组 - 不需要认证
	public := r.Group(basePath)
	public.Use(middleware.RateLimitMiddleware())

	{
		public.GET("/product/image", h.GetProductImage)
		public.GET("/product/list", h.GetProducts)
		public.GET("/product/detail", h.GetProduct)
		public.POST("/order/create", h.CreateOrder)
		public.GET("/order/user/list", h.GetOrdersByUser)
		public.GET("/order/detail", h.GetOrder)
		public.DELETE("/order/delete", h.DeleteOrder)
		// public.POST("/order/pay", h.PayOrder)
		public.GET("/tag/list", h.GetTagsForFront)
		public.GET("/tag/detail", h.GetTag)
		public.GET("/tag/bound-products", h.GetTagBoundProducts) // 获取标签已绑定的商品列表
		public.GET("/shop/:shopId/tags", h.GetShopTags)
		public.GET("/shop/detail", h.GetShopInfo)
		public.GET("/shop/image", h.GetShopImage)
		public.GET("/user/check-username", h.CheckUsernameExists)
	}
}
