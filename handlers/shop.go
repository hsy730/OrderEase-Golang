package handlers

import (
	"errors"
	"orderease/models"
	"orderease/utils/log2"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/datatypes"
	"gorm.io/gorm"

	"net/http"
	"strconv"
)

// GetShopTags 获取店铺标签列表
func (h *Handler) GetShopTags(c *gin.Context) {
	shopID, err := strconv.ParseUint(c.Param("shop_id"), 10, 64)
	if err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的店铺ID")
		return
	}

	tags, err := h.productRepo.GetShopTagsByID(shopID)
	if err != nil {
		h.logger.Printf("查询店铺标签失败，ID: %d，错误: %v", shopID, err)
		errorResponse(c, http.StatusInternalServerError, "查询失败")
		return
	}

	successResponse(c, gin.H{
		"total": len(tags),
		"tags":  tags,
	})
}

// GetShopInfo 获取店铺详细信息
func (h *Handler) GetShopInfo(c *gin.Context) {
	shopID := c.Query("shop_id")

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
		"id":             shop.ID,
		"name":           shop.Name,
		"owner_username": shop.OwnerUsername,
		"contact_phone":  shop.ContactPhone,
		"contact_email":  shop.ContactEmail,
		"address":        shop.Address,
		"description":    shop.Description,
		"created_at":     shop.CreatedAt.Format(time.RFC3339),
		"updated_at":     shop.UpdatedAt.Format(time.RFC3339),
		"valid_until":    shop.ValidUntil.Format(time.RFC3339),
		"settings":       shop.Settings,
		// "tags":           shop.Tags,
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

// CreateShop 创建店铺
func (h *Handler) CreateShop(c *gin.Context) {
	var shopData struct {
		Name          string `json:"name" binding:"required"`
		OwnerUsername string `json:"owner_username" binding:"required"`
		OwnerPassword string `json:"owner_password" binding:"required"`
		ContactPhone  string `json:"contact_phone"`
		ContactEmail  string `json:"contact_email"`
		Description   string `json:"description"`
		ValidUntil    string `json:"valid_until"`
		Address       string `json:"address"`
	}

	if err := c.ShouldBindJSON(&shopData); err != nil {
		log2.Errorf("Bind Json failed: %v", err)
		errorResponse(c, http.StatusBadRequest, "无效的请求数据")
		return
	}

	// 检查用户名是否已存在
	var count int64
	h.DB.Model(&models.Shop{}).Where("owner_username = ?", shopData.OwnerUsername).Count(&count)
	if count > 0 {
		errorResponse(c, http.StatusConflict, "店主用户名已存在")
		return
	}

	// 处理有效期
	validUntil := time.Now().AddDate(1, 0, 0) // 默认有效期1年
	if shopData.ValidUntil != "" {
		parsedValidUntil, err := time.Parse(time.RFC3339, shopData.ValidUntil)
		if err != nil {
			errorResponse(c, http.StatusBadRequest, "无效的有效期格式")
			return
		}
		validUntil = parsedValidUntil
	}

	newShop := models.Shop{
		Name:          shopData.Name,
		OwnerUsername: shopData.OwnerUsername,
		OwnerPassword: shopData.OwnerPassword, // 密码将在BeforeSave钩子中加密
		ContactPhone:  shopData.ContactPhone,
		ContactEmail:  shopData.ContactEmail,
		Description:   shopData.Description,
		ValidUntil:    validUntil,
		Address:       shopData.Address,
		Settings:      datatypes.JSON(`{}`), // 初始化为空对象
	}

	if err := h.DB.Create(&newShop).Error; err != nil {
		h.logger.Printf("创建店铺失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, "创建店铺失败")
		return
	}

	successResponse(c, gin.H{
		"code": 200,
		"data": gin.H{
			"id":             newShop.ID,
			"name":           newShop.Name,
			"description":    newShop.Description,
			"owner_username": newShop.OwnerUsername,
			"contact_phone":  newShop.ContactPhone,
			"address":        newShop.Address,
			"contact_email":  newShop.ContactEmail,
			"valid_until":    newShop.ValidUntil.Format(time.RFC3339),
			"settings":       newShop.Settings,
		},
	})
}

// UpdateShop 更新店铺信息
func (h *Handler) UpdateShop(c *gin.Context) {
	var updateData struct {
		ID            uint64  `json:"id" binding:"required"`
		OwnerUsername string  `json:"owner_username" binding:"required"`
		OwnerPassword *string `json:"owner_password"` // 使用指针类型以区分null和空字符串
		Name          string  `json:"name"`
		ContactPhone  string  `json:"contact_phone"`
		ContactEmail  string  `json:"contact_email"`
		Description   string  `json:"description"`
		ValidUntil    string  `json:"valid_until"`
	}

	if err := c.ShouldBindJSON(&updateData); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求数据")
		return
	}

	// 查询现有店铺
	shop, err := h.productRepo.GetShopByID(updateData.ID)
	if err != nil {
		errorResponse(c, http.StatusNotFound, "店铺不存在")
		return
	}

	// 验证店主用户名
	if shop.OwnerUsername != updateData.OwnerUsername {
		errorResponse(c, http.StatusUnauthorized, "用户名不匹配")
		return
	}

	// 更新字段
	if updateData.Name != "" {
		shop.Name = updateData.Name
	}
	if updateData.ContactPhone != "" {
		shop.ContactPhone = updateData.ContactPhone
	}
	if updateData.ContactEmail != "" {
		shop.ContactEmail = updateData.ContactEmail
	}
	if updateData.Description != "" {
		shop.Description = updateData.Description
	}
	if updateData.ValidUntil != "" {
		validUntil, err := time.Parse(time.RFC3339, updateData.ValidUntil)
		if err != nil {
			errorResponse(c, http.StatusBadRequest, "无效的有效期格式")
			return
		}
		shop.ValidUntil = validUntil
	}
	// 处理密码更新：如果密码不为null，则更新密码
	if updateData.OwnerPassword != nil {
		shop.OwnerPassword = *updateData.OwnerPassword
	}

	if err := h.DB.Save(&shop).Error; err != nil {
		h.logger.Printf("更新店铺失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, "更新店铺失败")
		return
	}

	// 构建响应数据（不包含密码）
	successResponse(c, gin.H{
		"code": 200,
		"data": gin.H{
			"id":             shop.ID,
			"name":           shop.Name,
			"description":    shop.Description,
			"owner_username": shop.OwnerUsername,
			"contact_phone":  shop.ContactPhone,
			"address":        shop.Address,
			"contact_email":  shop.ContactEmail,
			"valid_until":    shop.ValidUntil.Format(time.RFC3339),
			"settings":       shop.Settings,
		},
	})
}

// 删除店铺及关联数据
func (h *Handler) DeleteShop(c *gin.Context) {
	shopID, err := strconv.ParseUint(c.Query("shop_id"), 10, 64)
	if err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的店铺ID")
		return
	}

	// 开启事务
	tx := h.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 检查是否存在关联商品
	var productCount int64
	if err := tx.Model(&models.Product{}).Where("shop_id = ?", shopID).Count(&productCount).Error; err != nil {
		tx.Rollback()
		h.logger.Printf("查询关联商品失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, "删除店铺失败")
		return
	}

	if productCount > 0 {
		tx.Rollback()
		errorResponse(c, http.StatusConflict, "存在关联商品，无法删除店铺")
		return
	}

	// 删除店铺记录
	if err := tx.Where("id = ?", shopID).Delete(&models.Shop{}).Error; err != nil {
		tx.Rollback()
		h.logger.Printf("删除店铺失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, "删除店铺失败")
		return
	}

	tx.Commit()
	successResponse(c, gin.H{"message": "店铺删除成功"})
}
