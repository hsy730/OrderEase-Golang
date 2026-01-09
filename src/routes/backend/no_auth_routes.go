package backend

import (
	"orderease/config"
	"orderease/handlers"
	"orderease/middleware"

	"github.com/gin-gonic/gin"
)

// SetupNoAuthRoutes 配置后端不需要认证的路由
func SetupNoAuthRoutes(r *gin.Engine, h *handlers.Handler) {
	basePath := config.AppConfig.Server.BasePath

	// 公开路由组 - 不需要认证
	public := r.Group(basePath)
	public.Use(middleware.RateLimitMiddleware())

	// 认证相关路由
	setupAuthRoutes(public, h)
}

// setupAuthRoutes 配置认证相关路由
func setupAuthRoutes(group *gin.RouterGroup, h *handlers.Handler) {
	// @Summary 通用登录接口
	// @Description 管理员和商家通用登录接口
	// @Tags 认证
	// @Accept json
	// @Produce json
	// @Param loginRequest body UniversalLoginRequest true "登录信息"
	// @Success 200 {object} Response
	// @Router /api/login [post]
	group.POST("/login", h.UniversalLogin)

	// @Summary 刷新管理员令牌
	// @Description 刷新管理员访问令牌
	// @Tags 认证
	// @Accept json
	// @Produce json
	// @Param refreshTokenRequest body RefreshTokenRequest true "刷新令牌信息"
	// @Success 200 {object} Response
	// @Router /api/admin/refresh-token [post]
	group.POST("/admin/refresh-token", h.RefreshAdminToken)

	// @Summary 刷新商家令牌
	// @Description 刷新商家访问令牌
	// @Tags 认证
	// @Accept json
	// @Produce json
	// @Param refreshTokenRequest body RefreshTokenRequest true "刷新令牌信息"
	// @Success 200 {object} Response
	// @Router /api/shop/refresh-token [post]
	group.POST("/shop/refresh-token", h.RefreshShopToken)

	// @Summary 临时令牌登录
	// @Description 使用临时令牌登录接口
	// @Tags 认证
	// @Accept json
	// @Produce json
	// @Param tempTokenRequest body TempTokenRequest true "临时令牌信息"
	// @Success 200 {object} Response
	// @Router /api/shop/temp-login [post]
	group.POST("/shop/temp-login", h.TempTokenLogin)
}
