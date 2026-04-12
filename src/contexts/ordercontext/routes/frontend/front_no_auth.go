package frontend

import (
	"orderease/config"
	ordercontextHandlers "orderease/contexts/ordercontext/application/handlers"
	"orderease/middleware"

	"github.com/gin-gonic/gin"
)

func SetupFrontNoAuthRoutes(r *gin.Engine, h *ordercontextHandlers.Handler) {
	// 获取基础路径
	basePath := config.AppConfig.Server.BasePath

	// 公开路由组 - 不需要认证
	public := r.Group(basePath)
	public.Use(middleware.RateLimitMiddleware())
	public.POST("/user/login", h.FrontendUserLogin)       // 前端用户登录
	public.POST("/user/register", h.FrontendUserRegister) // 前端用户注册
	public.GET("/user/check-username", h.CheckUsernameExists)

	// 新增：微信小程序登录接口
	miniProgramHandler := h.GetMiniProgramAuthHandler()
	if miniProgramHandler != nil {
		public.POST("/user/wechat-login", miniProgramHandler.WeChatMiniProgramLogin)
	}

	// 图片接口公开访问，便于 CDN 缓存和浏览器缓存
	public.GET("/product/image", h.GetProductImage)
	public.GET("/shop/image", h.GetShopImage)
	public.GET("/user/avatar", h.GetUserAvatar)

}
