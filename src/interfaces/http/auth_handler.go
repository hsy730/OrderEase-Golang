package http

import (
	"net/http"
	"orderease/application/services"
	"orderease/domain/shared"
	"orderease/models"
	"orderease/utils"
	"orderease/utils/log2"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type AuthHandler struct {
	db          *gorm.DB
	shopService *services.ShopService
	userService *services.UserService
}

func NewAuthHandler(db *gorm.DB, shopService *services.ShopService, userService *services.UserService) *AuthHandler {
	return &AuthHandler{
		db:          db,
		shopService: shopService,
		userService: userService,
	}
}

// Login 通用登录接口（管理员和店主）
func (h *AuthHandler) Login(c *gin.Context) {
	type LoginRequest struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
		IsAdmin  bool   `json:"is_admin"`
	}

	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log2.Errorf("无效的登录数据: %s, %v", req.Username, err)
		errorResponse(c, http.StatusBadRequest, "无效的登录数据")
		return
	}

	log2.Debugf("开始处理登录请求, 用户名: %s", req.Username)

	// 尝试管理员登录
	var admin models.Admin
	if err := h.db.Where("username = ?", req.Username).First(&admin).Error; err == nil {
		if !admin.CheckPassword(req.Password) {
			log2.Errorf("管理员密码验证失败, 用户名: %s", req.Username)
			errorResponse(c, http.StatusUnauthorized, "用户名或密码错误")
			return
		}

		token, expiredAt, err := utils.GenerateToken(admin.ID, admin.Username, true)
		if err != nil {
			log2.Errorf("生成token失败: %v", err)
			errorResponse(c, http.StatusInternalServerError, "登录失败")
			return
		}

		log2.Infof("管理员登录成功: %s", admin.Username)
		successResponse(c, gin.H{
			"role":      "admin",
			"user_info": gin.H{"id": admin.ID, "username": admin.Username},
			"token":     token,
			"expiredAt": expiredAt.Unix(),
		})
		return
	}

	// 管理员登录失败，尝试店主登录
	var shop models.Shop
	if err := h.db.Where("owner_username = ?", req.Username).First(&shop).Error; err != nil {
		log2.Errorf("登录失败，用户名: %s, 错误: %v", req.Username, err)
		errorResponse(c, http.StatusUnauthorized, "用户名或密码错误")
		return
	}

	if shop.IsExpired() {
		log2.Errorf("店铺已到期: %s", shop.Name)
		errorResponse(c, http.StatusUnauthorized, "店铺已到期")
		return
	}

	if err := shop.CheckPassword(req.Password); err != nil {
		log2.Errorf("店主密码验证失败, 用户名: %s", req.Username)
		errorResponse(c, http.StatusUnauthorized, "用户名或密码错误")
		return
	}

	token, expiredAt, err := utils.GenerateToken(uint64(shop.ID), shop.OwnerUsername, false)
	if err != nil {
		log2.Errorf("生成token失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, "登录失败")
		return
	}

	log2.Infof("店主登录成功: %s", shop.OwnerUsername)
	successResponse(c, gin.H{
		"role":      "shop",
		"user_info": gin.H{"id": shop.ID, "shop_name": shop.Name, "username": shop.OwnerUsername},
		"token":     token,
		"expiredAt": expiredAt.Unix(),
	})
}

// Register 前端用户注册接口
func (h *AuthHandler) Register(c *gin.Context) {
	type RegisterRequest struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
		Phone     string `json:"phone"`
		Address   string `json:"address"`
	}

	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求数据")
		return
	}

	// TODO: 实现注册逻辑
	// 1. 检查用户名是否已存在
	// 2. 创建新用户
	// 3. 返回成功信息

	successResponse(c, gin.H{
		"code":    200,
		"message": "注册成功",
	})
}

// RefreshShopToken 刷新店主令牌
func (h *AuthHandler) RefreshShopToken(c *gin.Context) {
	type RefreshTokenRequest struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}

	var req RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求数据")
		return
	}

	// TODO: 实现 token 刷新逻辑
	successResponse(c, gin.H{
		"code":    200,
		"message": "令牌刷新成功",
		"token":   "new_token",
	})
}

// RefreshAdminToken 刷新管理员令牌
func (h *AuthHandler) RefreshAdminToken(c *gin.Context) {
	type RefreshTokenRequest struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}

	var req RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求数据")
		return
	}

	// TODO: 实现 token 刷新逻辑
	successResponse(c, gin.H{
		"code":    200,
		"message": "令牌刷新成功",
		"token":   "new_token",
	})
}

// TempTokenLogin 临时令牌登录
func (h *AuthHandler) TempTokenLogin(c *gin.Context) {
	type TempTokenRequest struct {
		TempToken string `json:"temp_token" binding:"required"`
		ShopID    string `json:"shop_id" binding:"required"`
	}

	var req TempTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求数据")
		return
	}

	// TODO: 实现临时令牌登录逻辑
	successResponse(c, gin.H{
		"code":    200,
		"message": "临时令牌登录成功",
		"token":   "new_token",
	})
}

// Logout 登出
func (h *AuthHandler) Logout(c *gin.Context) {
	// TODO: 将 token 加入黑名单
	successResponse(c, gin.H{
		"code":    200,
		"message": "登出成功",
	})
}

// ChangePassword 修改密码
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	type ChangePasswordRequest struct {
		OldPassword string `json:"old_password" binding:"required"`
		NewPassword string `json:"new_password" binding:"required"`
	}

	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求数据")
		return
	}

	// TODO: 实现修改密码逻辑
	successResponse(c, gin.H{
		"code":    200,
		"message": "密码修改成功",
	})
}

// GetShopTempToken 获取店铺临时令牌
func (h *AuthHandler) GetShopTempToken(c *gin.Context) {
	shopIDStr := c.Query("shop_id")
	if shopIDStr == "" {
		errorResponse(c, http.StatusBadRequest, "缺少店铺ID")
		return
	}

	shopID, err := shared.ParseIDFromString(shopIDStr)
	if err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的店铺ID")
		return
	}

	// 生成临时令牌
	tempToken := utils.GenerateTempToken()

	successResponse(c, gin.H{
		"code":       200,
		"temp_token": tempToken,
		"shop_id":    shopID.String(),
		"expires_in": 3600,
	})
}

// GenerateTempToken 生成临时令牌（辅助函数）
func utilsGenerateTempToken(shopID shared.ID) string {
	// TODO: 实现临时令牌生成逻辑
	return "temp_token_" + shopID.String()
}
