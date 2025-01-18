package handlers

import (
	"fmt"
	"net/http"
	"orderease/models"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// CreateTag 创建标签
func (h *Handler) CreateTag(c *gin.Context) {
	var tag models.Tag
	if err := c.ShouldBindJSON(&tag); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的标签数据")
		return
	}

	if err := h.DB.Create(&tag).Error; err != nil {
		h.logger.Printf("创建标签失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, "创建标签失败")
		return
	}

	successResponse(c, tag)
}

// GetTagOnlineProducts 获取标签关联的已上架商品
func (h *Handler) GetTagOnlineProducts(c *gin.Context) {
	tagID := c.Query("tag_id")
	if tagID == "" {
		errorResponse(c, http.StatusBadRequest, "缺少标签ID")
		return
	}

	var products []models.Product
	err := h.DB.Joins("JOIN product_tags ON product_tags.product_id = products.id").
		Where("product_tags.tag_id = ? AND products.status = ?", tagID, models.ProductStatusOnline).
		Find(&products).Error

	if err != nil {
		h.logger.Printf("查询标签关联商品失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, "查询失败")
		return
	}

	successResponse(c, gin.H{
		"tag_id":   tagID,
		"products": products,
	})
}

// GetBoundTags 获取商品已绑定的标签
func (h *Handler) GetBoundTags(c *gin.Context) {
	productID := c.Query("product_id")
	if productID == "" {
		errorResponse(c, http.StatusBadRequest, "缺少商品ID")
		return
	}

	var tags []models.Tag
	err := h.DB.Raw(`
		SELECT tags.* FROM tags
		JOIN product_tags ON product_tags.tag_id = tags.id
		WHERE product_tags.product_id = ?`, productID).Scan(&tags).Error

	if err != nil {
		h.logger.Printf("查询已绑定标签失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, "查询失败")
		return
	}

	successResponse(c, gin.H{
		"product_id": productID,
		"tags":       tags,
	})
}

// GetUnboundTags 获取商品未绑定的标签
func (h *Handler) GetUnboundTags(c *gin.Context) {
	productID := c.Query("product_id")
	if productID == "" {
		errorResponse(c, http.StatusBadRequest, "缺少商品ID")
		return
	}

	var tags []models.Tag
	err := h.DB.Raw(`
		SELECT * FROM tags 
		WHERE id NOT IN (
			SELECT tag_id FROM product_tags 
			WHERE product_id = ?
		)`, productID).Scan(&tags).Error

	if err != nil {
		h.logger.Printf("查询未绑定标签失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, "查询失败")
		return
	}

	successResponse(c, gin.H{
		"product_id": productID,
		"tags":       tags,
	})
}

// BatchTagProducts 批量打标签
func (h *Handler) BatchTagProducts(c *gin.Context) {
	type request struct {
		ProductIDs []uint `json:"product_ids" binding:"required"`
		TagID      uint   `json:"tag_id" binding:"required"`
	}

	var req request
	if err := c.ShouldBindJSON(&req); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求数据")
		return
	}

	// 检查标签是否存在
	var tag models.Tag
	if err := h.DB.First(&tag, req.TagID).Error; err != nil {
		h.logger.Printf("标签不存在, ID: %d", req.TagID)
		errorResponse(c, http.StatusNotFound, "标签不存在")
		return
	}

	// 批量创建关联
	var productTags []models.ProductTag
	for _, productID := range req.ProductIDs {
		productTags = append(productTags, models.ProductTag{
			ProductID: productID,
			TagID:     req.TagID,
		})
	}

	if err := h.DB.Create(&productTags).Error; err != nil {
		h.logger.Printf("批量打标签失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, "批量打标签失败")
		return
	}

	successResponse(c, gin.H{
		"message":    "批量打标签成功",
		"total":      len(req.ProductIDs),
		"successful": len(productTags),
	})
}

// UpdateTag 更新标签
func (h *Handler) UpdateTag(c *gin.Context) {
	var tag models.Tag
	if err := c.ShouldBindJSON(&tag); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的标签数据")
		return
	}

	if err := h.DB.Save(&tag).Error; err != nil {
		h.logger.Printf("更新标签失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, "更新标签失败")
		return
	}

	successResponse(c, tag)
}

// DeleteTag 删除标签
func (h *Handler) DeleteTag(c *gin.Context) {
	id := c.Query("id")
	if id == "" {
		errorResponse(c, http.StatusBadRequest, "缺少标签ID")
		return
	}

	// 检查标签是否存在
	var tag models.Tag
	if err := h.DB.First(&tag, id).Error; err != nil {
		h.logger.Printf("标签不存在, ID: %s", id)
		errorResponse(c, http.StatusNotFound, "标签不存在")
		return
	}

	// 检查是否有关联的商品
	var count int64
	if err := h.DB.Model(&models.ProductTag{}).Where("tag_id = ?", id).Count(&count).Error; err != nil {
		h.logger.Printf("检查标签关联商品失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, "删除标签失败")
		return
	}

	// 如果有关联商品，不允许删除
	if count > 0 {
		errorResponse(c, http.StatusBadRequest,
			fmt.Sprintf("该标签已关联 %d 个商品，请先解除关联后再删除", count))
		return
	}

	// 删除标签
	if err := h.DB.Delete(&tag).Error; err != nil {
		h.logger.Printf("删除标签失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, "删除标签失败")
		return
	}

	successResponse(c, gin.H{"message": "标签删除成功"})
}

// GetTags 获取标签列表
func (h *Handler) GetTags(c *gin.Context) {
	var tags []models.Tag
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))

	offset := (page - 1) * pageSize

	var total int64
	h.DB.Model(&models.Tag{}).Count(&total)

	if err := h.DB.Offset(offset).Limit(pageSize).Find(&tags).Error; err != nil {
		h.logger.Printf("获取标签列表失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, "获取标签列表失败")
		return
	}

	successResponse(c, gin.H{
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
		"tags":     tags,
	})
}

// GetTag 获取标签详情
func (h *Handler) GetTag(c *gin.Context) {
	id := c.Query("id")
	if id == "" {
		errorResponse(c, http.StatusBadRequest, "缺少标签ID")
		return
	}

	var tag models.Tag
	if err := h.DB.Preload("Products").First(&tag, id).Error; err != nil {
		h.logger.Printf("获取标签详情失败: %v", err)
		errorResponse(c, http.StatusNotFound, "标签不存在")
		return
	}

	successResponse(c, tag)
}

// BatchTagProduct 批量设置商品标签
func (h *Handler) BatchTagProduct(c *gin.Context) {
	type request struct {
		ProductID uint   `json:"product_id" binding:"required"`
		TagIDs    []uint `json:"tag_ids" binding:"required"`
	}

	var req request
	if err := c.ShouldBindJSON(&req); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求数据")
		return
	}

	// 获取当前标签
	var currentTags []models.Tag
	if err := h.DB.Joins("JOIN product_tags ON product_tags.tag_id = tags.id").
		Where("product_tags.product_id = ?", req.ProductID).
		Find(&currentTags).Error; err != nil {
		h.logger.Printf("获取当前标签失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, "获取当前标签失败")
		return
	}

	// 计算需要添加和删除的标签
	currentTagMap := make(map[uint]bool)
	for _, tag := range currentTags {
		currentTagMap[tag.ID] = true
	}

	newTagMap := make(map[uint]bool)
	for _, tagID := range req.TagIDs {
		newTagMap[tagID] = true
	}

	// 需要添加的标签
	var tagsToAdd []models.ProductTag
	for _, tagID := range req.TagIDs {
		if !currentTagMap[tagID] {
			tagsToAdd = append(tagsToAdd, models.ProductTag{
				ProductID: req.ProductID,
				TagID:     tagID,
			})
		}
	}

	// 需要删除的标签
	var tagsToDelete []uint
	for _, tag := range currentTags {
		if !newTagMap[tag.ID] {
			tagsToDelete = append(tagsToDelete, tag.ID)
		}
	}

	// 在事务中执行操作
	err := h.DB.Transaction(func(tx *gorm.DB) error {
		// 添加新标签
		if len(tagsToAdd) > 0 {
			if err := tx.Create(&tagsToAdd).Error; err != nil {
				return err
			}
		}

		// 删除旧标签
		if len(tagsToDelete) > 0 {
			if err := tx.Where("product_id = ? AND tag_id IN (?)", req.ProductID, tagsToDelete).
				Delete(&models.ProductTag{}).Error; err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		h.logger.Printf("批量更新标签失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, "批量更新标签失败")
		return
	}

	successResponse(c, gin.H{
		"message":       "批量更新标签成功",
		"added_count":   len(tagsToAdd),
		"deleted_count": len(tagsToDelete),
	})
}
