package handlers

import (
	"net/http"
	"orderease/models"
	"orderease/utils"
	"strings"

	"github.com/gin-gonic/gin"
)

// 管理员登录
func (h *Handler) AdminLogin(c *gin.Context) {
	utils.Logger.Printf("开始处理管理员登录请求")

	var loginData struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := c.ShouldBindJSON(&loginData); err != nil {
		utils.Logger.Printf("无效的登录数据: %s, %v", loginData.Username, err)
		errorResponse(c, http.StatusBadRequest, "无效的登录数据")
		return
	}

	var admin models.Admin
	if err := h.DB.Where("username = ?", loginData.Username).First(&admin).Error; err != nil {
		utils.Logger.Printf("管理员登录失败, 用户名: %s, 错误: %v", loginData.Username, err)
		errorResponse(c, http.StatusUnauthorized, "用户名或密码错误")
		return
	}

	if !admin.CheckPassword(loginData.Password) {
		utils.Logger.Printf("管理员密码验证失败, 用户名: %s", loginData.Username)
		errorResponse(c, http.StatusUnauthorized, "用户名或密码错误")
		return
	}

	// 生成JWT token
	token, expiredAt, err := utils.GenerateToken(admin.ID, admin.Username)
	if err != nil {
		utils.Logger.Printf("生成token失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, "登录失败")
		return
	}

	successResponse(c, gin.H{
		"message": "登录成功",
		"admin": gin.H{
			"id":       admin.ID,
			"username": admin.Username,
		},
		"token":     token,
		"expiredAt": expiredAt.Unix(),
	})
}

// 修改管理员密码
func (h *Handler) ChangeAdminPassword(c *gin.Context) {
	utils.Logger.Printf("开始处理管理员修改密码请求")

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
		utils.Logger.Printf("查找管理员失败: %v", err)
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
		utils.Logger.Printf("密码加密失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, "修改密码失败")
		return
	}

	if err := h.DB.Save(&admin).Error; err != nil {
		utils.Logger.Printf("保存新密码失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, "修改密码失败")
		return
	}

	successResponse(c, gin.H{"message": "密码修改成功"})
}

// RefreshToken 刷新token
func (h *Handler) RefreshToken(c *gin.Context) {
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

	// 生成新token
	newToken, expiredAt, err := utils.GenerateToken(claims.UserID, claims.Username)
	if err != nil {
		utils.Logger.Printf("生成新token失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, "刷新token失败")
		return
	}

	successResponse(c, gin.H{
		"message":   "token刷新成功",
		"token":     newToken,
		"expiredAt": expiredAt.Unix(),
	})
}
