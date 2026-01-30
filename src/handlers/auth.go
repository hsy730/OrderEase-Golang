package handlers

import (
	"net/http"
	"orderease/domain/shop"
	"orderease/domain/shared/value_objects"
	"orderease/models"
	"orderease/utils"
	"orderease/utils/log2"
	"strings"
	"time"

	"github.com/bwmarrin/snowflake"
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
	admin, err := h.adminRepo.GetAdminByUsername(loginData.Username)
	if err == nil {
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
	shopModel, err := h.shopRepo.GetByUsername(loginData.Username)
	if err != nil {
		log2.Errorf("登录失败，用户名: %s, 错误: %v", loginData.Username, err)
		errorResponse(c, http.StatusUnauthorized, "用户名或密码错误")
		return
	}

	// 检查店铺是否过期
	if err := h.checkShopExpiration(shopModel); err != nil {
		errorResponse(c, http.StatusUnauthorized, err.Error())
		return
	}

	// 转换为领域实体验证密码
	shopDomain := shop.ShopFromModel(shopModel)
	if err := shopDomain.CheckPassword(loginData.Password); err != nil {
		log2.Errorf("店主密码验证失败, 用户名: %s", loginData.Username)
		errorResponse(c, http.StatusUnauthorized, "用户名或密码错误")
		return
	}

	token, expiredAt, err := utils.GenerateToken(uint64(shopModel.ID), "shop_"+shopModel.OwnerUsername)
	if err != nil {
		log2.Errorf("生成token失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, "登录失败")
		return
	}

	successResponse(c, gin.H{
		"role":      "shop",
		"user_info": gin.H{"id": shopModel.ID, "shop_name": shopModel.Name, "username": shopModel.OwnerUsername},
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
	admin, err := h.adminRepo.GetFirstAdmin()
	if err != nil {
		log2.Errorf("查找管理员失败: %v", err)
		errorResponse(c, http.StatusNotFound, "管理员账户不存在")
		return
	}

	// 验证旧密码
	if !admin.CheckPassword(passwordData.OldPassword) {
		errorResponse(c, http.StatusUnauthorized, "旧密码错误")
		return
	}

	// 验证新密码强度（使用 Domain 值对象）
	if _, err := value_objects.NewStrictPassword(passwordData.NewPassword); err != nil {
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

	// 使用 Repository 更新管理员
	if err := h.adminRepo.Update(admin); err != nil {
		log2.Errorf("保存新密码失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, "修改密码失败")
		return
	}

	successResponse(c, gin.H{"message": "密码修改成功"})
}

func (h *Handler) ChangeShopPassword(c *gin.Context) {
	log2.Debugf("开始处理店主密码修改请求")

	// 从上下文中获取店主ID
	// 从上下文中获取用户信息
	userInfo := c.MustGet("userInfo").(models.UserInfo)
	shopID := userInfo.UserID

	var passwordData struct {
		OldPassword string `json:"old_password"`
		NewPassword string `json:"new_password"`
	}

	if err := c.ShouldBindJSON(&passwordData); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求数据")
		return
	}

	// 获取当前店主账户
	shopModel, err := h.shopRepo.GetShopByID(snowflake.ID(shopID))
	if err != nil {
		log2.Errorf("查找店主失败: %v", err)
		if err.Error() == "店铺不存在" {
			errorResponse(c, http.StatusNotFound, "店铺账户不存在")
		} else {
			errorResponse(c, http.StatusInternalServerError, "查询店铺失败")
		}
		return
	}

	// 检查店铺是否过期
	if err := h.checkShopExpiration(shopModel); err != nil {
		errorResponse(c, http.StatusForbidden, err.Error())
		return
	}

	// 转换为领域实体验证密码
	shopDomain := shop.ShopFromModel(shopModel)
	if err := shopDomain.CheckPassword(passwordData.OldPassword); err != nil {
		errorResponse(c, http.StatusUnauthorized, "旧密码错误")
		return
	}

	// 验证新密码强度（使用 Domain 值对象）
	if _, err := value_objects.NewStrictPassword(passwordData.NewPassword); err != nil {
		errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	// 更新密码（使用 Domain 实体自动哈希）
	shopDomain.SetOwnerPassword(passwordData.NewPassword)
	updatedShop := shopDomain.ToModel()

	// 使用 Repository 更新店铺
	if err := h.shopRepo.Update(updatedShop); err != nil {
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
		log2.Debugf("invalid token: %s, error: %v", oldToken, err)
		errorResponse(c, http.StatusUnauthorized, "无效的token")
		return
	}

	if isShopOwner {
		// 去掉用户名前缀
		rawUsername := strings.TrimPrefix(claims.Username, "shop_")

		// 使用 Repository 查询当前店铺状态
		shopModel, err := h.shopRepo.GetByUsername(rawUsername)
		if err != nil {
			log2.Errorf("店铺查询失败: %s, 错误: %v", rawUsername, err)
			if err.Error() == "店铺不存在" {
				errorResponse(c, http.StatusUnauthorized, "店铺账户不存在")
			} else {
				errorResponse(c, http.StatusUnauthorized, "查询店铺失败")
			}
			return
		}

		// 检查店铺是否过期
		if err := h.checkShopExpiration(shopModel); err != nil {
			log2.Warnf("店铺服务已到期: %s (ID: %d)", shopModel.Name, shopModel.ID)
			errorResponse(c, http.StatusForbidden, err.Error())
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
		log2.Debugf("invalid token: %s, error: %v", token, err)
		errorResponse(c, http.StatusUnauthorized, "无效的token")
		return
	}

	// 将token加入黑名单
	blacklistedToken := models.BlacklistedToken{
		Token:     token,
		ExpiredAt: time.Unix(claims.ExpiresAt.Unix(), 0),
		CreatedAt: time.Now(),
	}

	// 使用 Repository 添加 token 到黑名单
	if err := h.tokenRepo.CreateBlacklistedToken(&blacklistedToken); err != nil {
		log2.Errorf("添加token到黑名单失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, "登出失败")
		return
	}

	successResponse(c, gin.H{
		"message": "登出成功",
	})
}

type UserInfo struct {
	UserID   uint
	Username string
	IsAdmin  bool
}

// TempTokenLogin 使用临时令牌登录
func (h *Handler) TempTokenLogin(c *gin.Context) {
	var loginData struct {
		ShopID snowflake.ID `json:"shop_id" binding:"required"`
		Token  string       `json:"token" binding:"required,len=6"`
	}

	if err := c.ShouldBindJSON(&loginData); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的登录数据")
		return
	}

	// 验证临时令牌
	valid, user, err := h.tempTokenService.ValidateTempToken(loginData.ShopID, loginData.Token)
	if err != nil || !valid {
		errorResponse(c, http.StatusUnauthorized, "无效的临时令牌")
		return
	}

	// 生成JWT令牌
	token, expiredAt, err := utils.GenerateToken(uint64(user.ID), user.Name)
	if err != nil {
		errorResponse(c, http.StatusInternalServerError, "生成令牌失败")
		return
	}

	// 使用 Repository 获取店铺信息
	shop, err := h.shopRepo.GetShopByID(loginData.ShopID)
	if err != nil {
		errorResponse(c, http.StatusInternalServerError, "获取店铺信息失败")
		return
	}

	successResponse(c, gin.H{
		"role":      "user",
		"user_info": gin.H{"id": user.ID, "name": user.Name, "shop_id": loginData.ShopID, "shop_name": shop.Name},
		"token":     token,
		"expiredAt": expiredAt.Unix(),
	})
}
