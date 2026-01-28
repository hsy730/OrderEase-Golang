package handlers

import (
	"fmt"
	"net/http"
	"orderease/domain/tag"
	"orderease/models"
	"orderease/utils/log2"
	"strconv"

	"github.com/bwmarrin/snowflake"
	"github.com/gin-gonic/gin"
)

// CreateTag 创建标签
func (h *Handler) CreateTag(c *gin.Context) {
	var tag models.Tag
	if err := c.ShouldBindJSON(&tag); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的标签数据")
		return
	}

	validShopID, err := h.validAndReturnShopID(c, tag.ShopID)
	if err != nil {
		errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	tag.ShopID = validShopID // 将shopID设置为请求的店铺ID

	if err := h.tagRepo.Create(&tag); err != nil {
		h.logger.Errorf("创建标签失败: %v", err)
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

	tagIDInt, err := strconv.Atoi(tagID)
	if err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的标签ID")
		return
	}

	requestShopID, err := strconv.ParseUint(c.Query("shop_id"), 10, 64)
	if err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的店铺ID")
		return
	}

	validShopID, err := h.validAndReturnShopID(c, requestShopID)
	if err != nil {
		errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	// 使用 Repository 获取标签关联的在线商品
	products, err := h.tagRepo.GetOnlineProductsByTag(tagIDInt, validShopID)
	if err != nil {
		h.logger.Errorf("查询标签关联商品失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, err.Error())
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

	requestShopID, err := strconv.ParseUint(c.Query("shop_id"), 10, 64)
	if err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的店铺ID")
		return
	}

	validShopID, err := h.validAndReturnShopID(c, requestShopID)
	if err != nil {
		errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	// 将productID转换为snowflake.ID
	productIDSnowflake, err := snowflake.ParseString(productID)
	if err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的商品ID")
		return
	}

	// 使用 repository 获取已绑定标签
	tags, err := h.productRepo.GetCurrentProductTags(productIDSnowflake, validShopID)
	if err != nil {
		h.logger.Errorf("查询已绑定标签失败: %v", err)
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

	requestShopID, err := strconv.ParseUint(c.Query("shop_id"), 10, 64)
	if err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的店铺ID")
		return
	}

	validShopID, err := h.validAndReturnShopID(c, requestShopID)
	if err != nil {
		errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	// 将productID转换为snowflake.ID
	productIDSnowflake, err := snowflake.ParseString(productID)
	if err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的商品ID")
		return
	}

	// 使用 repository 获取未绑定标签
	tags, err := h.productRepo.GetUnboundTags(productIDSnowflake, validShopID)
	if err != nil {
		h.logger.Errorf("查询未绑定标签失败: %v", err)
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
		ProductIDs []snowflake.ID `json:"product_ids" binding:"required"`
		TagID      int            `json:"tag_id" binding:"required"`
		ShopID     uint64         `json:"shop_id" binding:"required"`
	}

	var req request
	if err := c.ShouldBindJSON(&req); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求数据")
		return
	}

	validShopID, err := h.validAndReturnShopID(c, req.ShopID)
	if err != nil {
		errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	req.ShopID = validShopID // 将shopID设置为请求的店铺ID

	// 检查标签是否存在
	tag, err := h.tagRepo.GetByIDAndShopID(req.TagID, req.ShopID)
	if err != nil {
		h.logger.Errorf("标签不存在, ID: %d", req.TagID)
		errorResponse(c, http.StatusNotFound, "标签不存在")
		return
	}

	// 使用 Repository 批量打标签
	result, err := h.tagRepo.BatchTagProducts(req.ProductIDs, req.TagID, tag.ShopID)
	if err != nil {
		h.logger.Errorf("批量打标签失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	successResponse(c, gin.H{
		"message":    "批量打标签成功",
		"total":      result.Total,
		"successful": result.Successful,
	})
}

// UpdateTag 更新标签
func (h *Handler) UpdateTag(c *gin.Context) {
	var tag models.Tag
	if err := c.ShouldBindJSON(&tag); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的标签数据")
		return
	}

	validShopID, err := h.validAndReturnShopID(c, tag.ShopID)
	if err != nil {
		errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	tag.ShopID = validShopID // 将shopID设置为请求的店铺ID

	if err := h.tagRepo.Update(&tag); err != nil {
		h.logger.Errorf("更新标签失败: %v", err)
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

	requestShopID, err := strconv.ParseUint(c.Query("shop_id"), 10, 64)
	if err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的店铺ID")
		return
	}

	validShopID, err := h.validAndReturnShopID(c, requestShopID)
	if err != nil {
		errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	// 检查标签是否存在
	tagIDInt, _ := strconv.Atoi(id)
	tag, err := h.tagRepo.GetByIDAndShopID(tagIDInt, validShopID)
	if err != nil {
		h.logger.Errorf("标签不存在, ID: %s", id)
		errorResponse(c, http.StatusNotFound, "标签不存在")
		return
	}

	// 检查是否有关联的商品
	var count int64
	if err := h.DB.Model(&models.ProductTag{}).Where("tag_id = ?", id).Count(&count).Error; err != nil {
		h.logger.Errorf("检查标签关联商品失败: %v", err)
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
	if err := h.tagRepo.Delete(tag); err != nil {
		h.logger.Errorf("删除标签失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, "删除标签失败")
		return
	}

	successResponse(c, gin.H{"message": "标签删除成功"})
}

func (h *Handler) GetTagsForFront(c *gin.Context) {
	h.GetTags(c, true)
}

func (h *Handler) GetTagsForBackend(c *gin.Context) {
	h.GetTags(c, false)
}

// GetTags 获取标签列表, 全部
func (h *Handler) GetTags(c *gin.Context, isFront bool) {
	var tags []models.Tag
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))

	requestShopID, err := strconv.ParseUint(c.Query("shop_id"), 10, 64)
	if err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的店铺ID")
		return
	}

	validShopID, err := h.validAndReturnShopID(c, requestShopID)
	if err != nil {
		errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	// 查询所有标签
	tags, err = h.tagRepo.GetListByShopID(validShopID)
	if err != nil {
		h.logger.Errorf("获取标签列表失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, "获取标签列表失败")
		return
	}

	// 检测是否存在未绑定标签的商品
	if isFront {
		log2.Debugf("GetTags isFront: %v ", isFront)
		unbindCount, err := h.tagRepo.GetUnboundProductsCount(validShopID)
		if err != nil {
			h.logger.Errorf("检查未绑定商品失败: %v", err)
		}

		// 如果存在未绑定商品，添加虚拟标签
		if unbindCount > 0 {
			tags = append(tags, models.Tag{
				ID:   -1,
				Name: "其他",
			})
		}
	}

	// 获取总数时包含虚拟标签
	total := int64(len(tags))

	successResponse(c, gin.H{
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
		"tags":     tags,
	})
}

// GetUnboundProductsForTag 获取标签未绑定的商品列表
func (h *Handler) GetUnboundProductsForTag(c *gin.Context) {
	tagID := c.Query("tag_id")
	if tagID == "" {
		errorResponse(c, http.StatusBadRequest, "缺少标签ID")
		return
	}

	requestShopID, err := strconv.ParseUint(c.Query("shop_id"), 10, 64)
	if err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的店铺ID")
		return
	}

	validShopID, err := h.validAndReturnShopID(c, requestShopID)
	if err != nil {
		errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))

	tagIDInt, _ := strconv.Atoi(tagID)

	// 使用 Repository 方法查询未绑定该标签的商品
	products, total, err := h.tagRepo.GetUnboundProductsForTag(tagIDInt, validShopID, page, pageSize)
	if err != nil {
		h.logger.Errorf("查询未绑定商品失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	successResponse(c, gin.H{
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
		"data":     products,
	})
}

// GetUnboundTagsList 获取没有绑定商品的标签列表
func (h *Handler) GetUnboundTagsList(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))

	requestShopID, err := strconv.ParseUint(c.Query("shop_id"), 10, 64)
	if err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的店铺ID")
		return
	}

	validShopID, err := h.validAndReturnShopID(c, requestShopID)
	if err != nil {
		errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	// 使用 Repository 方法查询未绑定商品的标签
	tags, total, err := h.tagRepo.GetUnboundTagsList(validShopID, page, pageSize)
	if err != nil {
		h.logger.Errorf("查询未绑定商品标签失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	successResponse(c, gin.H{
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
		"tags":     tags,
	})
}

// GetTagBoundProducts 获取标签已绑定的商品列表（分页）
func (h *Handler) GetTagBoundProducts(c *gin.Context) {
	tagID := c.Query("tag_id")
	if tagID == "" {
		errorResponse(c, http.StatusBadRequest, "缺少标签ID")
		return
	}

	requestShopID, err := strconv.ParseUint(c.Query("shop_id"), 10, 64)
	if err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的店铺ID")
		return
	}

	validShopID, err := h.validAndReturnShopID(c, requestShopID)
	if err != nil {
		errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))

	// 处理未绑定商品查询
	if tagID == "-1" {
		// 使用 Repository 获取未绑定商品
		result, err := h.tagRepo.GetUnboundProductsWithPagination(validShopID, page, pageSize)
		if err != nil {
			h.logger.Errorf("查询未绑定商品失败: %v", err)
			errorResponse(c, http.StatusInternalServerError, "查询失败")
			return
		}
		successResponse(c, gin.H{
			"total":    result.Total,
			"page":     page,
			"pageSize": pageSize,
			"data":     result.Products,
		})
		return
	}

	tagIDInt, _ := strconv.Atoi(tagID)

	// 使用 Repository 获取标签绑定的商品（分页）
	result, err := h.tagRepo.GetBoundProductsWithPagination(tagIDInt, validShopID, page, pageSize)
	if err != nil {
		h.logger.Errorf("查询绑定商品失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	successResponse(c, gin.H{
		"total":    result.Total,
		"page":     page,
		"pageSize": pageSize,
		"data":     result.Products,
	})
}

// 新增独立方法处理未绑定商品查询（已移至 Repository，此方法保留以防其他地方调用）
func (h *Handler) getUnboundProducts(shopID uint64, page int, pageSize int) ([]models.Product, int64, error) {
	result, err := h.tagRepo.GetUnboundProductsWithPagination(shopID, page, pageSize)
	if err != nil {
		return nil, 0, err
	}
	return result.Products, result.Total, nil
}

// GetTag 获取标签详情
func (h *Handler) GetTag(c *gin.Context) {
	id := c.Query("id")
	if id == "" {
		errorResponse(c, http.StatusBadRequest, "缺少标签ID")
		return
	}

	requestShopID, err := strconv.ParseUint(c.Query("shop_id"), 10, 64)
	if err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的店铺ID")
		return
	}

	validShopID, err := h.validAndReturnShopID(c, requestShopID)
	if err != nil {
		errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	tagIDInt, _ := strconv.Atoi(id)
	tag, err := h.tagRepo.GetByIDAndShopID(tagIDInt, validShopID)
	if err != nil {
		h.logger.Errorf("获取标签详情失败: %v", err)
		errorResponse(c, http.StatusNotFound, "标签不存在")
		return
	}
	successResponse(c, tag)
}

// BatchUntagProducts 批量解绑商品标签
func (h *Handler) BatchUntagProducts(c *gin.Context) {
	type request struct {
		ProductIDs []snowflake.ID `json:"product_ids" binding:"required"`
		TagID      uint           `json:"tag_id" binding:"required"`
		ShopID     uint64         `json:"shop_id" binding:"required"`
	}

	var req request
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Errorf("批量解绑标签, 数据绑定错误: %v", err)
		errorResponse(c, http.StatusBadRequest, "无效的请求数据")
		return
	}

	validShopID, err := h.validAndReturnShopID(c, req.ShopID)
	if err != nil {
		errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}
	req.ShopID = validShopID // 将shopID设置为请求的店铺ID

	// 检查标签是否存在
	tag, err := h.tagRepo.GetByIDAndShopID(int(req.TagID), validShopID)
	if err != nil {
		h.logger.Errorf("标签不存在, ID: %d", req.TagID)
		errorResponse(c, http.StatusNotFound, "标签不存在")
		return
	}

	// 使用 Repository 批量解绑标签
	result, err := h.tagRepo.BatchUntagProducts(req.ProductIDs, req.TagID, tag.ShopID)
	if err != nil {
		h.logger.Errorf("批量解绑标签失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	successResponse(c, gin.H{
		"message":    "批量解绑标签成功",
		"total":      result.Total,
		"successful": result.Successful,
	})
}

// BatchTagProduct 批量设置商品标签
func (h *Handler) BatchTagProduct(c *gin.Context) {
	type request struct {
		ProductID snowflake.ID `json:"product_id" binding:"required"`
		TagIDs    []int        `json:"tag_ids" binding:"required"`
		ShopID    uint64       `json:"shop_id" binding:"required"`
	}

	var req request
	if err := c.ShouldBindJSON(&req); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求数据")
		return
	}

	validShopID, err := h.validAndReturnShopID(c, req.ShopID)
	if err != nil {
		errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}
	req.ShopID = validShopID // 将shopID设置为请求的店铺ID

	// 替换原有查询代码
	currentTags, err := h.productRepo.GetCurrentProductTags(req.ProductID, req.ShopID)
	if err != nil {
		errorResponse(c, http.StatusInternalServerError, "获取当前标签失败")
		return
	}

	// 使用 Tag Domain Service 更新标签关联
	result, err := h.tagService.UpdateProductTags(tag.UpdateProductTagsDTO{
		CurrentTags: currentTags,
		NewTagIDs:   req.TagIDs,
		ProductID:   req.ProductID,
		ShopID:      req.ShopID,
	})
	if err != nil {
		h.logger.Errorf("批量更新标签失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, "批量更新标签失败")
		return
	}

	successResponse(c, gin.H{
		"message":       "批量更新标签成功",
		"added_count":   result.AddedCount,
		"deleted_count": result.DeletedCount,
	})
}

// 注意：updateProductTags 方法已迁移到 domain/tag/service.go (Step 45)
