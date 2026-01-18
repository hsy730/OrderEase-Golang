package handlers

import (
	"fmt"
	"net/http"
	"orderease/models"
	"orderease/utils/log2"
	"strconv"

	"github.com/bwmarrin/snowflake"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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

	var products []models.Product
	err = h.DB.Joins("JOIN product_tags ON product_tags.product_id = products.id").
		Where("product_tags.tag_id = ? AND products.status = ? AND products.shop_id = ?",
			tagID, models.ProductStatusOnline, validShopID).
		Find(&products).Error

	if err != nil {
		h.logger.Errorf("查询标签关联商品失败: %v", err)
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

	// 批量创建关联
	var productTags []models.ProductTag

	// 批量查询商品的店铺信息
	var validProducts []models.Product
	if err := h.DB.Select("id").Where("id IN (?) AND shop_id = ?", req.ProductIDs, tag.ShopID).Find(&validProducts).Error; err != nil {
		h.logger.Errorf("批量查询商品失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, "批量操作失败")
		return
	}

	// 构建有效商品ID集合
	validProductMap := make(map[snowflake.ID]bool)
	for _, p := range validProducts {
		validProductMap[p.ID] = true
	}

	// 过滤无效商品并生成关联记录
	var successCount int
	for _, productID := range req.ProductIDs {
		if validProductMap[productID] {
			productTags = append(productTags, models.ProductTag{
				ProductID: productID,
				TagID:     req.TagID,
				ShopID:    validShopID,
			})
			successCount++
		}
	}

	// 使用 INSERT IGNORE 避免重复插入错误
	if err := h.DB.Clauses(clause.OnConflict{
		DoNothing: true,
	}).Create(&productTags).Error; err != nil {
		h.logger.Errorf("批量打标签失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, "批量打标签失败:"+err.Error())
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
		var unbindCount int64
		h.DB.Raw(`SELECT COUNT(*) FROM products 
        WHERE shop_id = ? AND id NOT IN (SELECT product_id FROM product_tags)`).Scan(&unbindCount)

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
	offset := (page - 1) * pageSize

	var products []models.Product
	var total int64

	// 查询未绑定该标签的商品
	err = h.DB.Raw(`
		SELECT * FROM products
		WHERE id NOT IN (
			SELECT product_id FROM product_tags
			WHERE tag_id = ? AND shop_id =?
		) ORDER BY created_at DESC LIMIT ? OFFSET ?`, tagID, validShopID, pageSize, offset).Scan(&products).Error

	if err != nil {
		h.logger.Errorf("查询未绑定商品失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, "查询失败")
		return
	}

	// 获取总数
	h.DB.Raw(`
		SELECT COUNT(*) FROM products
		WHERE id NOT IN (
			SELECT product_id FROM product_tags
			WHERE tag_id = ? AND shop_id =?
		) ORDER BY created_at DESC`, tagID, validShopID).Scan(&total)

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
	offset := (page - 1) * pageSize

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

	var tags []models.Tag
	var total int64

	// 查询没有绑定商品的标签
	err = h.DB.Raw(`
		SELECT * FROM tags
		WHERE shop_id = ? ANS id NOT IN (
			SELECT DISTINCT tag_id FROM product_tags
		) ORDER BY created_at DESC LIMIT ? OFFSET ?`, validShopID, pageSize, offset).Scan(&tags).Error

	if err != nil {
		h.logger.Errorf("查询未绑定商品标签失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, "查询失败")
		return
	}

	// 获取总数
	h.DB.Raw(`
		SELECT COUNT(*) FROM tags
		WHERE shop_id = ? AND id NOT IN (
			SELECT DISTINCT tag_id FROM product_tags
		) ORDER BY created_at DESC`, validShopID).Scan(&total)

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
	offset := (page - 1) * pageSize

	// 处理未绑定商品查询
	if tagID == "-1" {
		products, total, err := h.getUnboundProducts(validShopID, page, pageSize)
		if err != nil {
			h.logger.Errorf("查询未绑定商品失败: %v", err)
			errorResponse(c, http.StatusInternalServerError, "查询失败")
			return
		}
		successResponse(c, gin.H{
			"total":    total,
			"page":     page,
			"pageSize": pageSize,
			"data":     products,
		})
		return
	}
	// 原有绑定标签商品的查询逻辑保持不变...
	var total int64

	// 获取已绑定商品的ID列表
	var productIDs []snowflake.ID
	err = h.DB.Raw(`
		SELECT product_id FROM product_tags
		WHERE tag_id = ? AND shop_id = ?`, tagID, validShopID).Pluck("product_id", &productIDs).Error

	if err != nil {
		h.logger.Errorf("获取商品ID列表失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, "查询失败")
		return
	}

	log2.Debugf("获取商品ID列表: %v", productIDs)

	// 查询完整商品数据并预加载选项
	var products []models.Product
	err = h.DB.Where("id IN (?)", productIDs).
		Preload("OptionCategories.Options").
		Order("created_at DESC").
		Limit(pageSize).Offset(offset).
		Find(&products).Error

	if err != nil {
		h.logger.Errorf("查询商品详情失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, "查询失败")
		return
	}

	// 获取总数
	err = h.DB.Model(&models.Product{}).
		Where("id IN (?)", productIDs).
		Count(&total).Error

	if err != nil {
		h.logger.Errorf("获取商品总数失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, "查询失败")
		return
	}

	successResponse(c, gin.H{
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
		"data":     products,
	})
}

// 新增独立方法处理未绑定商品查询
func (h *Handler) getUnboundProducts(shopID uint64, page int, pageSize int) ([]models.Product, int64, error) {
	offset := (page - 1) * pageSize
	var products []models.Product
	var total int64

	query := h.DB.Model(&models.Product{}).
		Where("shop_id = ? AND id NOT IN (SELECT product_id FROM product_tags)", shopID)

	// 获取总数
	if err := query.
		Model(&models.Product{}).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Offset(offset).
		Limit(pageSize).Order("created_at DESC").
		Preload("OptionCategories.Options").
		Find(&products).Error; err != nil {
		return nil, 0, err
	}

	return products, total, nil
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
	_ = tag // 使用tag避免unused变量警告

	// 批量删除关联
	result := h.DB.Where("shop_id = ? AND tag_id = ? AND product_id IN (?)", req.ShopID, req.TagID, req.ProductIDs).
		Delete(&models.ProductTag{})

	if result.Error != nil {
		h.logger.Errorf("批量解绑标签失败: %v", result.Error)
		errorResponse(c, http.StatusInternalServerError, "批量解绑标签失败")
		return
	}

	successResponse(c, gin.H{
		"message":    "批量解绑标签成功",
		"total":      len(req.ProductIDs),
		"successful": result.RowsAffected,
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

	// 替换原有计算和更新逻辑
	added, deleted, err := h.updateProductTags(currentTags, req.TagIDs, req.ProductID, req.ShopID)
	if err != nil {
		h.logger.Errorf("批量更新标签失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, "批量更新标签失败")
		return
	}

	successResponse(c, gin.H{
		"message":       "批量更新标签成功",
		"added_count":   added,
		"deleted_count": deleted,
	})
}

// 新增方法：更新商品标签关联
func (h *Handler) updateProductTags(currentTags []models.Tag, newTagIDs []int, productID snowflake.ID, shopID uint64) (int, int, error) {
	// 计算差异
	currentTagMap := make(map[int]bool)
	for _, tag := range currentTags {
		currentTagMap[tag.ID] = true
	}

	newTagMap := make(map[int]bool)
	for _, tagID := range newTagIDs {
		newTagMap[tagID] = true
	}

	// 准备操作数据
	var tagsToAdd []models.ProductTag
	var tagsToDelete []int

	// 计算需要添加的标签
	for _, tagID := range newTagIDs {
		if !currentTagMap[tagID] {
			tagsToAdd = append(tagsToAdd, models.ProductTag{
				ProductID: productID,
				TagID:     tagID,
				ShopID:    shopID,
			})
		}
	}

	// 计算需要删除的标签
	for _, tag := range currentTags {
		if !newTagMap[tag.ID] {
			tagsToDelete = append(tagsToDelete, tag.ID)
		}
	}

	// 执行事务操作
	err := h.DB.Transaction(func(tx *gorm.DB) error {
		if len(tagsToAdd) > 0 {
			if err := tx.Create(&tagsToAdd).Error; err != nil {
				return err
			}
		}

		if len(tagsToDelete) > 0 {
			if err := tx.Where("product_id = ? AND tag_id IN (?)", productID, tagsToDelete).
				Delete(&models.ProductTag{}).Error; err != nil {
				return err
			}
		}
		return nil
	})

	return len(tagsToAdd), len(tagsToDelete), err
}
