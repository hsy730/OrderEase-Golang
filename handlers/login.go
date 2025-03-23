package handlers

import (
	"net/http"
	"orderease/models"
	"orderease/utils"
	"orderease/utils/log2"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// 管理员登录
func (h *Handler) UniversalLogin(c *gin.Context) {
	log2.Debugf("开始处理统一登录请求")

	var loginData struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&loginData); err != nil {
		log2.Errorf("无效的登录数据: %s, %v", loginData.Username, err)
		errorResponse(c, http.StatusBadRequest, "无效的登录数据")
		return
	}

	// 尝试管理员登录
	var admin models.Admin
	if err := h.DB.Where("username = ?", loginData.Username).First(&admin).Error; err == nil {
		if !admin.CheckPassword(loginData.Password) {
			log2.Errorf("管理员密码验证失败, 用户名: %s", loginData.Username)
			errorResponse(c, http.StatusUnauthorized, "用户名或密码错误")
			return
		}

		token, expiredAt, err := utils.GenerateToken(admin.ID, admin.Username)
		if err != nil {
			log2.Errorf("生成token失败: %v", err)
			errorResponse(c, http.StatusInternalServerError, "登录失败")
			return
		}

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
	if err := h.DB.Where("owner_username = ?", loginData.Username).First(&shop).Error; err != nil {
		log2.Errorf("登录失败，用户名: %s, 错误: %v", loginData.Username, err)
		errorResponse(c, http.StatusUnauthorized, "用户名或密码错误")
		return
	}

	if shop.IsExpired() {
		errorResponse(c, http.StatusUnauthorized, "店铺已到期")
		return
	}

	if err := shop.CheckPassword(loginData.Password); err != nil {
		log2.Errorf("店主密码验证失败, 用户名: %s", loginData.Username)
		errorResponse(c, http.StatusUnauthorized, "用户名或密码错误")
		return
	}

	token, expiredAt, err := utils.GenerateToken(shop.ID, "shop_"+shop.OwnerUsername)
	if err != nil {
		log2.Errorf("生成token失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, "登录失败")
		return
	}

	successResponse(c, gin.H{
		"role":      "shop",
		"user_info": gin.H{"id": shop.ID, "shop_name": shop.Name, "username": shop.OwnerUsername},
		"token":     token,
		"expiredAt": expiredAt.Unix(),
	})
}

// 修改管理员密码
func (h *Handler) ChangeAdminPassword(c *gin.Context) {
	log2.Debugf("开始处理管理员修改密码请求")

	var passwordData struct {
		OldPassword string `json:"old_password"`
		NewPassword string `json:"new_password"`
	}

	if err := c.ShouldBindJSON(&passwordData); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求数据")
		return
	}

	// 获取唯一的管理员账户
	var admin models.Admin
	if err := h.DB.First(&admin).Error; err != nil {
		log2.Errorf("查找管理员失败: %v", err)
		errorResponse(c, http.StatusNotFound, "管理员账户不存在")
		return
	}

	// 验证旧密码
	if !admin.CheckPassword(passwordData.OldPassword) {
		errorResponse(c, http.StatusUnauthorized, "旧密码错误")
		return
	}

	// 验证新密码强度
	if err := utils.ValidatePassword(passwordData.NewPassword); err != nil {
		errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	// 更新密码
	admin.Password = passwordData.NewPassword
	if err := admin.HashPassword(); err != nil {
		log2.Errorf("密码加密失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, "修改密码失败")
		return
	}

	if err := h.DB.Save(&admin).Error; err != nil {
		log2.Errorf("保存新密码失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, "修改密码失败")
		return
	}

	successResponse(c, gin.H{"message": "密码修改成功"})
}

func (h *Handler) ChangeShopPassword(c *gin.Context) {
	log2.Debugf("开始处理店主密码修改请求")

	// 从上下文中获取店主ID
	shopID := c.MustGet("userID").(uint)

	var passwordData struct {
		OldPassword string `json:"old_password"`
		NewPassword string `json:"new_password"`
	}

	if err := c.ShouldBindJSON(&passwordData); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求数据")
		return
	}

	// 获取当前店主账户
	var shop models.Shop
	if err := h.DB.First(&shop, shopID).Error; err != nil {
		log2.Errorf("查找店主失败: %v", err)
		errorResponse(c, http.StatusNotFound, "店铺账户不存在")
		return
	}

	if shop.IsExpired() {
		errorResponse(c, http.StatusForbidden, "店铺服务已到期")
		return
	}

	// 验证旧密码
	if err := shop.CheckPassword(passwordData.OldPassword); err != nil {
		errorResponse(c, http.StatusUnauthorized, "旧密码错误")
		return
	}

	// 验证新密码强度（复用相同规则）
	if err := utils.ValidatePassword(passwordData.NewPassword); err != nil {
		errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	// 更新密码
	shop.OwnerPassword = passwordData.NewPassword
	if err := shop.HashPassword(); err != nil {
		log2.Errorf("密码加密失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, "修改密码失败")
		return
	}

	if err := h.DB.Save(&shop).Error; err != nil {
		log2.Errorf("保存新密码失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, "修改密码失败")
		return
	}

	successResponse(c, gin.H{"message": "密码修改成功"})
}
func (h *Handler) RefreshAdminToken(c *gin.Context) {
	h.RefreshToken(c, false)
}

func (h *Handler) RefreshShopToken(c *gin.Context) {
	h.RefreshToken(c, true)
}

// RefreshToken 刷新token
func (h *Handler) RefreshToken(c *gin.Context, isShopOwner bool) {
	// 从请求头获取旧token
	oldToken := c.GetHeader("Authorization")
	if oldToken == "" {
		errorResponse(c, http.StatusBadRequest, "缺少token")
		return
	}

	// 去掉Bearer前缀
	oldToken = strings.TrimPrefix(oldToken, "Bearer ")

	// 验证旧token
	claims, err := utils.ParseToken(oldToken)
	if err != nil {
		errorResponse(c, http.StatusUnauthorized, "无效的token")
		return
	}

	if isShopOwner {
		// 去掉用户名前缀
		rawUsername := strings.TrimPrefix(claims.Username, "shop_")

		// 查询当前店铺状态
		var shop models.Shop
		if err := h.DB.Where("owner_username = ?", rawUsername).First(&shop).Error; err != nil {
			log2.Errorf("店铺查询失败: %s, 错误: %v", rawUsername, err)
			errorResponse(c, http.StatusUnauthorized, "店铺账户不存在")
			return
		}

		if shop.IsExpired() {
			log2.Warnf("店铺服务已到期: %s (ID: %d)", shop.Name, shop.ID)
			errorResponse(c, http.StatusForbidden, "店铺服务已到期")
			return
		}
	}

	// 生成新token
	newToken, expiredAt, err := utils.GenerateToken(claims.UserID, claims.Username)
	if err != nil {
		log2.Errorf("生成新token失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, "刷新token失败")
		return
	}

	successResponse(c, gin.H{
		"message":   "token刷新成功",
		"token":     newToken,
		"expiredAt": expiredAt.Unix(),
	})
}

// Logout 管理员登出
func (h *Handler) Logout(c *gin.Context) {
	// 从请求头获取token
	token := c.GetHeader("Authorization")
	if token == "" {
		errorResponse(c, http.StatusBadRequest, "缺少token")
		return
	}

	// 去掉Bearer前缀
	token = strings.TrimPrefix(token, "Bearer ")

	// 验证token
	claims, err := utils.ParseToken(token)
	if err != nil {
		errorResponse(c, http.StatusUnauthorized, "无效的token")
		return
	}

	// 将token加入黑名单
	blacklistedToken := models.BlacklistedToken{
		Token:     token,
		ExpiredAt: time.Unix(claims.ExpiresAt.Unix(), 0),
		CreatedAt: time.Now(),
	}

	if err := h.DB.Create(&blacklistedToken).Error; err != nil {
		log2.Errorf("添加token到黑名单失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, "登出失败")
		return
	}

	successResponse(c, gin.H{
		"message": "登出成功",
	})
}
