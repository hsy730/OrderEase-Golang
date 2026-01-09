package frontend

import (
	"orderease/config"
	"orderease/handlers"
	"orderease/middleware"

	"github.com/gin-gonic/gin"
)

// SetupFrontNoAuthRoutes 配置前端不需要认证的路由
func SetupFrontNoAuthRoutes(r *gin.Engine, h *handlers.Handler) {
	// 获取基础路径
	basePath := config.AppConfig.Server.BasePath

	// 公开路由组 - 不需要认证
	public := r.Group(basePath)
	public.Use(middleware.RateLimitMiddleware())

	// 用户认证相关路由
	setupAuthRoutes(public, h)
}

// setupAuthRoutes 配置用户认证相关路由
func setupAuthRoutes(group *gin.RouterGroup, h *handlers.Handler) {
	// @Summary 前端用户登录
	// @Description 前端用户登录接口
	// @Tags 用户认证
	// @Accept json
	// @Produce json
	// @Param loginRequest body LoginRequest true "登录信息"
	// @Success 200 {object} Response
	// @Router /api/user/login [post]
	group.POST("/user/login", h.FrontendUserLogin)

	// @Summary 前端用户注册
	// @Description 前端用户注册接口
	// @Tags 用户认证
	// @Accept json
	// @Produce json
	// @Param registerRequest body RegisterRequest true "注册信息"
	// @Success 200 {object} Response
	// @Router /api/user/register [post]
	group.POST("/user/register", h.FrontendUserRegister)

	// @Summary 检查用户名是否存在
	// @Description 检查用户名是否已被注册
	// @Tags 用户
	// @Accept json
	// @Produce json
	// @Param username query string true "用户名"
	// @Success 200 {object} Response
	// @Router /api/user/check-username [get]
	group.GET("/user/check-username", h.CheckUsernameExists)
}
