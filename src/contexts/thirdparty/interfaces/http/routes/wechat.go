package routes

import (
	thirdpartyHandlers "orderease/contexts/thirdparty/application/handlers"

	"github.com/gin-gonic/gin"
)

// SetupWeChatRoutes 设置微信路由
func SetupWeChatRoutes(r *gin.Engine, h *thirdpartyHandlers.WeChatHandler) {
	if h == nil {
		return
	}

	// 第三方平台 API 路由组
	thirdpartyAPI := r.Group("/api/order-ease/v1/thirdparty/wechat")
	{
		// 获取授权 URL（无需认证）
		thirdpartyAPI.GET("/authorize", h.Authorize)

		// 微信授权回调（无需认证）
		thirdpartyAPI.GET("/callback", h.Callback)

		// 获取微信配置（无需认证，用于前端判断是否显示微信登录按钮）
		thirdpartyAPI.GET("/config", h.GetConfig)

		// 通过 OpenID 直接登录（用于个人公众号，无需认证）
		thirdpartyAPI.POST("/login-by-openid", h.LoginByOpenID)
	}
}
