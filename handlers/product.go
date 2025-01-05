package handlers

import (
	"errors"
	"net/http"
	"orderease/models"
	"orderease/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ToggleProductStatus 更新商品状态
func (h *Handler) ToggleProductStatus(c *gin.Context) {
	// 解析请求参数
	var req struct {
		ID     uint   `json:"id" binding:"required"`
		Status string `json:"status" binding:"required,oneof=pending online offline"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Logger.Printf("解析请求参数失败: %v", err)
		errorResponse(c, http.StatusBadRequest, "无效的请求参数")
		return
	}

	// 获取当前商品信息
	var product models.Product
	if err := h.DB.First(&product, req.ID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			errorResponse(c, http.StatusNotFound, "商品不存在")
			return
		}
		utils.Logger.Printf("查询商品失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, "服务器内部错误")
		return
	}

	if product.Status == "" {
		product.Status = models.ProductStatusPending
	}

	utils.Logger.Printf("更新商品前状态: %s", product.Status)
	// 验证状态流转
	if !isValidProductStatusTransition(product.Status, req.Status) {
		errorResponse(c, http.StatusBadRequest, "无效的状态变更")
		return
	}

	// 更新商品状态
	if err := h.DB.Model(&product).Update("status", req.Status).Error; err != nil {
		utils.Logger.Printf("更新商品状态失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, "更新商品状态失败")
		return
	}

	// 返回成功响应
	successResponse(c, gin.H{
		"message": "商品状态更新成功",
		"product": gin.H{
			"id":         product.ID,
			"status":     req.Status,
			"updated_at": product.UpdatedAt,
		},
	})
}

// isValidProductStatusTransition 验证商品状态流转是否合法
func isValidProductStatusTransition(currentStatus, newStatus string) bool {
	// 定义状态流转规则
	transitions := map[string][]string{
		models.ProductStatusPending: {models.ProductStatusOnline},
		models.ProductStatusOnline:  {models.ProductStatusOffline},
		models.ProductStatusOffline: {}, // 已下架状态不能转换到其他状态
	}

	// 检查是否是允许的状态转换
	allowedStatus, exists := transitions[currentStatus]
	if !exists {
		return false
	}

	// 检查新状态是否在允许的转换列表中
	for _, status := range allowedStatus {
		if status == newStatus {
			return true
		}
	}

	return false
}
