package handlers

import (
	"orderease/models"

	"github.com/gin-gonic/gin"

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
