package handlers

import (
	"errors"
	"orderease/models"
	"orderease/utils"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"net/http"
	"strconv"
)

// GetShopTags 获取店铺标签列表
func (h *Handler) GetShopTags(c *gin.Context) {
	shopID := c.Param("shopId")

	// 转换店铺ID为数字
	shopIDInt, err := strconv.Atoi(shopID)
	if err != nil || shopIDInt <= 0 {
		errorResponse(c, http.StatusBadRequest, "无效的店铺ID")
		return
	}

	var tags []models.Tag
	if err := h.DB.Where("shop_id = ?", shopIDInt).Find(&tags).Error; err != nil {
		h.logger.Printf("查询店铺标签失败，店铺ID: %d，错误: %v", shopIDInt, err)
		errorResponse(c, http.StatusInternalServerError, "获取标签失败")
		return
	}

	// 如果查询结果为空返回空数组
	if len(tags) == 0 {
		tags = make([]models.Tag, 0)
	}

	successResponse(c, gin.H{
		"total": len(tags),
		"tags":  tags,
	})
}

// GetShopInfo 获取店铺详细信息
func (h *Handler) GetShopInfo(c *gin.Context) {
	shopID := c.Query("shopId")

	// 转换店铺ID为数字
	shopIDInt, err := strconv.Atoi(shopID)
	if err != nil || shopIDInt <= 0 {
		errorResponse(c, http.StatusBadRequest, "无效的店铺ID")
		return
	}

	var shop models.Shop
	if err := h.DB.Preload("Tags").First(&shop, shopIDInt).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			errorResponse(c, http.StatusNotFound, "店铺不存在")
			return
		}
		h.logger.Printf("查询店铺失败，ID: %d，错误: %v", shopIDInt, err)
		errorResponse(c, http.StatusInternalServerError, "查询失败")
		return
	}

	successResponse(c, gin.H{
		"id":          shop.ID,
		"name":        shop.Name,
		"description": shop.Description,
		"valid_until": shop.ValidUntil.Format(time.RFC3339),
		"settings":    shop.Settings,
		"created_at":  shop.CreatedAt.Format(time.RFC3339),
	})
}

// ShopOwnerLogin 店铺店主登录
func (h *Handler) ShopOwnerLogin(c *gin.Context) {
	var loginData struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&loginData); err != nil {
		h.logger.Printf("无效的登录数据: %v", err)
		errorResponse(c, http.StatusBadRequest, "无效的登录数据")
		return
	}

	var shop models.Shop
	if err := h.DB.Where("owner_username = ?", loginData.Username).First(&shop).Error; err != nil {
		h.logger.Printf("店主登录失败, 用户名: %s, 错误: %v", loginData.Username, err)
		errorResponse(c, http.StatusUnauthorized, "用户名或密码错误")
		return
	}

	// 验证密码
	if err := shop.CheckPassword(loginData.Password); err != nil {
		h.logger.Printf("密码验证失败, 用户名: %s", loginData.Username)
		errorResponse(c, http.StatusUnauthorized, "用户名或密码错误")
		return
	}

	// 生成JWT token
	token, expiredAt, err := utils.GenerateToken(shop.ID, shop.OwnerUsername)
	if err != nil {
		h.logger.Printf("生成token失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, "登录失败")
		return
	}

	successResponse(c, gin.H{
		"message": "登录成功",
		"shop_info": gin.H{
			"id":       shop.ID,
			"name":     shop.Name,
			"username": shop.OwnerUsername,
		},
		"token":     token,
		"expiredAt": expiredAt.Unix(),
	})
}
