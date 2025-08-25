package backend

import (
	"orderease/config"
	"orderease/handlers"
	"orderease/middleware"

	"github.com/gin-gonic/gin"
)

func SetupNoAuthRoutes(r *gin.Engine, h *handlers.Handler) {
	basePath := config.AppConfig.Server.BasePath

	// 公开路由组 - 不需要认证
	public := r.Group(basePath)
	public.Use(middleware.RateLimitMiddleware())

	{
		public.POST("/login", h.UniversalLogin) // 合并后的登录接口
		public.POST("/admin/refresh-token", h.RefreshAdminToken)
		public.POST("/shop/refresh-token", h.RefreshShopToken)
		public.GET("/admin/product/image", h.GetProductImage)
		public.GET("/admin/shop/image", h.GetShopImage)
	}
}
