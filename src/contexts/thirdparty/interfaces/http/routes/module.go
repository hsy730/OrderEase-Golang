package routes

import (
	thirdpartyHandlers "orderease/contexts/thirdparty/application/handlers"

	"github.com/gin-gonic/gin"
)

// SetupRoutes 设置第三方平台所有路由
func SetupRoutes(r *gin.Engine, h *thirdpartyHandlers.Handler) {
	if h == nil {
		return
	}

	// 设置微信路由
	if h.WeChat != nil {
		SetupWeChatRoutes(r, h.WeChat)
	}

	// 未来添加:
	// if h.Alipay != nil {
	//     SetupAlipayRoutes(r, h.Alipay)
	// }
}
