package backend

import (
	"orderease/config"
	ordercontextHandlers "orderease/contexts/ordercontext/application/handlers"
	"orderease/middleware"

	"github.com/gin-gonic/gin"
)

func SetupNoAuthRoutes(r *gin.Engine, h *ordercontextHandlers.Handler) {
	basePath := config.AppConfig.Server.BasePath

	// 公开路由组 - 不需要认证
	public := r.Group(basePath)
	public.Use(middleware.RateLimitMiddleware())

	{
		public.POST("/login", h.UniversalLogin) // 合并后的登录接口
		public.POST("/admin/refresh-token", h.RefreshAdminToken)
		public.POST("/shop/refresh-token", h.RefreshShopToken)
		public.POST("/shop/temp-login", h.TempTokenLogin)
	}
}
