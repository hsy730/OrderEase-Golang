package frontend

import (
	"orderease/config"
	"orderease/handlers"
	"orderease/middleware"

	"github.com/gin-gonic/gin"
)

func SetupFrontNoAuthRoutes(r *gin.Engine, h *handlers.Handler) {
	// 获取基础路径
	basePath := config.AppConfig.Server.BasePath

	// 公开路由组 - 不需要认证
	public := r.Group(basePath)
	public.Use(middleware.RateLimitMiddleware())
	public.POST("/user/login", h.FrontendUserLogin)       // 前端用户登录
	public.POST("/user/register", h.FrontendUserRegister) // 前端用户注册
	public.POST("/shop/temp-login", h.TempTokenLogin)

}
