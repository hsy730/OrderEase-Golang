// @Deprecated
// 本文件中的路由已被废弃，请使用 interfaces/http/router.go 中的新路由
// 新路由遵循 DDD 架构，提供更好的可维护性和扩展性
// 旧路由将在下个版本 (v2.0) 中完全移除
//
// 迁移指南:
// 1. handlers/ 的业务逻辑已迁移至 application/services/
// 2. 新的路由定义在 interfaces/http/router.go
// 3. 新的 Handler 在 interfaces/http/*_handler.go
//
package routes

import (
	"orderease/handlers"
	"orderease/routes/backend"
	"orderease/routes/frontend"

	"github.com/gin-gonic/gin"
)

// SetupRoutes 设置旧路由
// @Deprecated 请使用 interfaces/http/router.SetupRoutes() 代替
func SetupRoutes(r *gin.Engine, h *handlers.Handler) {
	backend.SetupAdminRoutes(r, h)
	backend.SetupNoAuthRoutes(r, h)
	backend.SetupShopRoutes(r, h)

	frontend.SetupFrontRoutes(r, h)
	frontend.SetupFrontNoAuthRoutes(r, h)
}
