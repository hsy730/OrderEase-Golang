package handlers

import (
	"errors"
	"orderease/models"
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

// GetShopList 获取店铺列表（分页+搜索）
func (h *Handler) GetShopList(c *gin.Context) {
	// 获取分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	search := c.Query("search")

	// 验证分页参数
	if page < 1 || pageSize < 1 || pageSize > 100 {
		errorResponse(c, http.StatusBadRequest, "无效的分页参数")
		return
	}

	query := h.DB.Model(&models.Shop{}).Preload("Tags")

	// 添加搜索条件
	if search != "" {
		search = "%" + search + "%"
		query = query.Where("name LIKE ? OR owner_username LIKE ?", search, search)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		h.logger.Printf("查询店铺总数失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, "查询失败")
		return
	}

	// 执行分页查询
	var shops []models.Shop
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("id DESC").Find(&shops).Error; err != nil {
		h.logger.Printf("查询店铺列表失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, "查询失败")
		return
	}

	// 构建响应数据
	responseData := make([]gin.H, 0, len(shops))
	for _, shop := range shops {
		responseData = append(responseData, gin.H{
			"id":             shop.ID,
			"name":           shop.Name,
			"owner_username": shop.OwnerUsername,
			"contact_phone":  shop.ContactPhone,
			"valid_until":    shop.ValidUntil.Format(time.RFC3339),
			"tags_count":     len(shop.Tags),
		})
	}

	successResponse(c, gin.H{
		"total": total,
		"page":  page,
		"data":  responseData,
	})
}
