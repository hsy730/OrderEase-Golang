package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/snowflake"
	"github.com/gin-gonic/gin"
	"gorm.io/datatypes"
	"gorm.io/gorm"

	"orderease/models"
	"orderease/utils"
	"orderease/utils/log2"
)

// CreateShopRequest 创建店铺请求
type CreateShopRequest struct {
	Name            string                  `json:"name" binding:"required" example:"美食店"`
	OwnerUsername   string                  `json:"owner_username" binding:"required" example:"shopowner"`
	OwnerPassword   string                  `json:"owner_password" binding:"required" example:"password123"`
	ContactPhone    string                  `json:"contact_phone" example:"13800138000"`
	ContactEmail    string                  `json:"contact_email" example:"shop@example.com"`
	Description     string                  `json:"description" example:"专营川菜"`
	ValidUntil      string                  `json:"valid_until" example:"2025-12-31T23:59:59Z"`
	Address         string                  `json:"address" example:"北京市朝阳区"`
	Settings        datatypes.JSON          `json:"settings"`
	OrderStatusFlow *models.OrderStatusFlow `json:"order_status_flow"`
}

// UpdateShopRequest 更新店铺请求
type UpdateShopRequest struct {
	ID           uint64         `json:"id" binding:"required" example:"1"`
	Name         string         `json:"name" example:"美食店"`
	ContactPhone string         `json:"contact_phone" example:"13800138000"`
	ContactEmail string         `json:"contact_email" example:"shop@example.com"`
	Description  string         `json:"description" example:"专营川菜"`
	Address      string         `json:"address" example:"北京市朝阳区"`
	Settings     datatypes.JSON `json:"settings"`
}

// OrderStatusFlowRequest 订单状态流转请求
type OrderStatusFlowRequest struct {
	ShopID          uint64                  `json:"shop_id" binding:"required" example:"1"`
	OrderStatusFlow *models.OrderStatusFlow `json:"order_status_flow" binding:"required"`
}

// GetShopTags 获取店铺标签列表
// @Summary 获取店铺标签列表
// @Description 获取指定店铺的标签列表
// @Tags 标签
// @Accept json
// @Produce json
// @Param shopId path string true "店铺ID"
// @Success 200 {object} map[string]interface{} "查询成功"
// @Security BearerAuth
// @Router /shop/{shopId}/tags [get]
func (h *Handler) GetShopTags(c *gin.Context) {
	shopIDUint64, err := strconv.ParseUint(c.Param("shopId"), 10, 64)
	if err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的店铺ID")
		return
	}
	shopID := snowflake.ID(shopIDUint64)

	tags, err := h.productRepo.GetShopTagsByID(uint64(shopID))
	if err != nil {
		h.logger.Errorf("查询店铺标签失败，ID: %d，错误: %v", shopID, err)
		errorResponse(c, http.StatusInternalServerError, "查询失败")
		return
	}

	successResponse(c, gin.H{
		"total": len(tags),
		"tags":  tags,
	})
}

// GetShopInfo 获取店铺详细信息
// @Summary 获取店铺详情
// @Description 获取指定店铺的详细信息
// @Tags 店铺管理
// @Accept json
// @Produce json
// @Param shopId query string true "店铺ID"
// @Success 200 {object} map[string]interface{} "查询成功"
// @Security BearerAuth
// @Router /admin/shop/detail [get]
// @Router /shopOwner/shop/detail [get]
// @Router /shop/detail [get]
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
		h.logger.Errorf("查询店铺失败，ID: %d，错误: %v", shopIDInt, err)
		errorResponse(c, http.StatusInternalServerError, "查询失败")
		return
	}

	successResponse(c, gin.H{
		"id":                shop.ID,
		"name":              shop.Name,
		"owner_username":    shop.OwnerUsername,
		"contact_phone":     shop.ContactPhone,
		"contact_email":     shop.ContactEmail,
		"address":           shop.Address,
		"description":       shop.Description,
		"created_at":        shop.CreatedAt.Format(time.RFC3339),
		"updated_at":        shop.UpdatedAt.Format(time.RFC3339),
		"valid_until":       shop.ValidUntil.Format(time.RFC3339),
		"settings":          shop.Settings,
		"tags":              shop.Tags,
		"image_url":         shop.ImageURL,
		"order_status_flow": shop.OrderStatusFlow,
	})
}

// GetShopList 获取店铺列表（分页+搜索）
// @Summary 获取店铺列表
// @Description 获取店铺列表，支持分页和筛选
// @Tags 店铺管理
// @Accept json
// @Produce json
// @Param page query int false "页码"
// @Param pageSize query int false "每页数量"
// @Param status query string false "店铺状态"
// @Success 200 {object} map[string]interface{} "查询成功"
// @Security BearerAuth
// @Router /admin/shop/list [get]
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
		h.logger.Errorf("查询店铺总数失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, "查询失败")
		return
	}

	// 执行分页查询
	var shops []models.Shop
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&shops).Error; err != nil {
		h.logger.Errorf("查询店铺列表失败: %v", err)
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
// @Summary 创建店铺
// @Description 创建新店铺
// @Tags 店铺管理
// @Accept json
// @Produce json
// @Param shop body CreateShopRequest true "店铺信息"
// @Success 200 {object} map[string]interface{} "创建成功"
// @Security BearerAuth
// @Router /admin/shop/create [post]
func (h *Handler) CreateShop(c *gin.Context) {
	var shopData struct {
		Name            string                  `json:"name" binding:"required"`
		OwnerUsername   string                  `json:"owner_username" binding:"required"`
		OwnerPassword   string                  `json:"owner_password" binding:"required"`
		ContactPhone    string                  `json:"contact_phone"`
		ContactEmail    string                  `json:"contact_email"`
		Description     string                  `json:"description"`
		ValidUntil      string                  `json:"valid_until"`
		Address         string                  `json:"address"`
		Settings        datatypes.JSON          `json:"settings"`
		OrderStatusFlow *models.OrderStatusFlow `json:"order_status_flow"`
	}

	if err := c.ShouldBindJSON(&shopData); err != nil {
		log2.Errorf("Bind Json failed: %v", err)
		errorResponse(c, http.StatusBadRequest, "无效的请求数据: "+err.Error())
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

	// 解析默认订单流转配置
	var orderStatusFlow models.OrderStatusFlow
	if err := json.Unmarshal([]byte(models.DefaultOrderStatusFlow), &orderStatusFlow); err != nil {
		h.logger.Errorf("解析默认订单流转配置失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, "解析默认订单流转配置失败")
		return
	}

	// 如果提供了订单流转配置，则使用提供的配置
	if shopData.OrderStatusFlow != nil {
		orderStatusFlow = *shopData.OrderStatusFlow
	}

	newShop := models.Shop{
		Name:            shopData.Name,
		OwnerUsername:   shopData.OwnerUsername,
		OwnerPassword:   shopData.OwnerPassword, // 密码将在BeforeSave钩子中加密
		ContactPhone:    shopData.ContactPhone,
		ContactEmail:    shopData.ContactEmail,
		Description:     shopData.Description,
		ValidUntil:      validUntil,
		Address:         shopData.Address,
		Settings:        json.RawMessage(shopData.Settings), // 转换为json.RawMessage
		OrderStatusFlow: orderStatusFlow,
	}

	if err := h.DB.Create(&newShop).Error; err != nil {
		h.logger.Errorf("创建店铺失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, "创建店铺失败")
		return
	}

	successResponse(c, gin.H{
		"code": 200,
		"data": gin.H{
			"id":                newShop.ID,
			"name":              newShop.Name,
			"description":       newShop.Description,
			"owner_username":    newShop.OwnerUsername,
			"contact_phone":     newShop.ContactPhone,
			"address":           newShop.Address,
			"contact_email":     newShop.ContactEmail,
			"valid_until":       newShop.ValidUntil.Format(time.RFC3339),
			"settings":          newShop.Settings,
			"order_status_flow": newShop.OrderStatusFlow,
		},
	})
}

// UpdateShop 更新店铺信息
// @Summary 更新店铺信息
// @Description 更新店铺基本信息
// @Tags 店铺管理
// @Accept json
// @Produce json
// @Param shop body UpdateShopRequest true "店铺信息"
// @Success 200 {object} map[string]interface{} "更新成功"
// @Security BearerAuth
// @Router /admin/shop/update [put]
// @Router /shopOwner/shop/update [put]
func (h *Handler) UpdateShop(c *gin.Context) {
	var updateData struct {
		ID              uint64                  `json:"id" binding:"required"`
		OwnerUsername   string                  `json:"owner_username" binding:"required"`
		OwnerPassword   *string                 `json:"owner_password"` // 使用指针类型以区分null和空字符串
		Name            string                  `json:"name"`
		ContactPhone    string                  `json:"contact_phone"`
		ContactEmail    string                  `json:"contact_email"`
		Description     string                  `json:"description"`
		ValidUntil      string                  `json:"valid_until"`
		Address         string                  `json:"address"`
		Settings        datatypes.JSON          `json:"settings"`
		OrderStatusFlow *models.OrderStatusFlow `json:"order_status_flow"`
	}

	if err := c.ShouldBindJSON(&updateData); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求数据: "+err.Error())
		return
	}

	// 检查OrderStatusFlow的Statuses是否为null，如果是则将整个OrderStatusFlow设为nil
	if updateData.OrderStatusFlow != nil && updateData.OrderStatusFlow.Statuses == nil {
		updateData.OrderStatusFlow = nil
	}

	// 查询现有店铺
	shop, err := h.productRepo.GetShopByID(updateData.ID)
	if err != nil {
		errorResponse(c, http.StatusNotFound, "店铺不存在")
		return
	}

	// 获取用户信息
	userInfo, exists := c.Get("userInfo")
	if !exists {
		errorResponse(c, http.StatusUnauthorized, "未获取到用户信息")
		return
	}

	user := userInfo.(models.UserInfo)

	// 检查是否在修改店铺过期时间，如果是，需要管理员权限
	if updateData.ValidUntil != "" && !user.IsAdmin {
		updateData.ValidUntil = ""
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

	if updateData.Address != "" {
		shop.Address = updateData.Address
	}

	if updateData.Settings != nil {
		shop.Settings = json.RawMessage(updateData.Settings)
	}

	// 如果提供了订单流转配置，则更新
	if updateData.OrderStatusFlow != nil {
		shop.OrderStatusFlow = *updateData.OrderStatusFlow
	} else {
		// 如果数据库中也没有订单流转信息，则填充默认配置
		if len(shop.OrderStatusFlow.Statuses) == 0 {
			var defaultOrderStatusFlow models.OrderStatusFlow
			if err := json.Unmarshal([]byte(models.DefaultOrderStatusFlow), &defaultOrderStatusFlow); err != nil {
				h.logger.Errorf("解析默认订单流转配置失败: %v", err)
				errorResponse(c, http.StatusInternalServerError, "解析默认订单流转配置失败")
				return
			}
			shop.OrderStatusFlow = defaultOrderStatusFlow
		}
	}
	if updateData.ValidUntil != "" {
		validUntil, err := time.Parse(time.RFC3339, updateData.ValidUntil)
		if err != nil {
			errorResponse(c, http.StatusBadRequest, "无效的有效期格式")
			return
		}
		shop.ValidUntil = validUntil
	}

	if updateData.OwnerUsername != "" {
		shop.OwnerUsername = updateData.OwnerUsername
	}
	// 处理密码更新：如果密码不为null，则更新密码
	if updateData.OwnerPassword != nil {
		shop.OwnerPassword = *updateData.OwnerPassword
	}

	if err := h.DB.Save(&shop).Error; err != nil {
		h.logger.Errorf("更新店铺失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, "更新店铺失败")
		return
	}

	// 构建响应数据（不包含密码）
	successResponse(c, gin.H{
		"code": 200,
		"data": gin.H{
			"id":                shop.ID,
			"name":              shop.Name,
			"description":       shop.Description,
			"owner_username":    shop.OwnerUsername,
			"contact_phone":     shop.ContactPhone,
			"address":           shop.Address,
			"contact_email":     shop.ContactEmail,
			"valid_until":       shop.ValidUntil.Format(time.RFC3339),
			"settings":          shop.Settings,
			"order_status_flow": shop.OrderStatusFlow,
		},
	})
}

// DeleteShop 删除店铺及关联数据
// @Summary 删除店铺
// @Description 删除指定店铺
// @Tags 店铺管理
// @Accept json
// @Produce json
// @Param shopId query string true "店铺ID"
// @Success 200 {object} map[string]interface{} "删除成功"
// @Security BearerAuth
// @Router /admin/shop/delete [delete]
func (h *Handler) DeleteShop(c *gin.Context) {
	shopIDUint64, err := strconv.ParseUint(c.Query("shop_id"), 10, 64)
	if err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的店铺ID")
		return
	}
	shopID := snowflake.ID(shopIDUint64)

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
		h.logger.Errorf("查询关联商品失败: %v", err)
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
		h.logger.Errorf("删除店铺失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, "删除店铺失败")
		return
	}

	tx.Commit()
	successResponse(c, gin.H{"message": "店铺删除成功"})
}

// CheckShopNameExists 检查商店名称是否存在
// CheckShopNameExists 检查店铺名称是否存在
// @Summary 检查店铺名称是否存在
// @Description 检查店铺名称是否已被使用
// @Tags 店铺管理
// @Accept json
// @Produce json
// @Param name query string true "店铺名称"
// @Success 200 {object} map[string]interface{} "查询成功"
// @Security BearerAuth
// @Router /admin/shop/check-name [get]
func (h *Handler) CheckShopNameExists(c *gin.Context) {
	shopName := c.Query("name")
	if shopName == "" {
		errorResponse(c, http.StatusBadRequest, "商店名称不能为空")
		return
	}

	var count int64
	if err := h.DB.Model(&models.Shop{}).Where("name = ?", shopName).Count(&count).Error; err != nil {
		h.logger.Errorf("检查商店名称失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, "检查商店名称失败")
		return
	}

	successResponse(c, gin.H{
		"exists": count > 0,
	})
}

// UploadShopImage 上传店铺图片
// @Summary 上传店铺图片
// @Description 上传店铺图片
// @Tags 店铺管理
// @Accept multipart/form-data
// @Produce json
// @Param image formData file true "店铺图片"
// @Success 200 {object} map[string]interface{} "上传成功"
// @Security BearerAuth
// @Router /admin/shop/upload-image [post]
func (h *Handler) UploadShopImage(c *gin.Context) {
	// 限制文件大小
	const maxFileSize = 2 * 1024 * 1024 // 2MB
	const maxZipSize = 512 * 1024
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxFileSize)

	idUint64, err := strconv.ParseUint(c.Query("id"), 10, 64)
	if err != nil {
		errorResponse(c, http.StatusBadRequest, "缺少店铺ID")
		return
	}
	id := snowflake.ID(idUint64)

	// 查询店铺
	var shop models.Shop
	if err := h.DB.First(&shop, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			errorResponse(c, http.StatusNotFound, "店铺不存在")
			return
		}
		h.logger.Errorf("查询店铺失败，ID: %d，错误: %v", id, err)
		errorResponse(c, http.StatusInternalServerError, "查询店铺失败")
		return
	}

	// 获取上传文件
	file, err := c.FormFile("image")
	if err != nil {
		log2.Errorf("获取上传文件失败: %v", err)
		errorResponse(c, http.StatusBadRequest, "获取上传文件失败")
		return
	}

	// 检查文件类型
	if !utils.IsValidImageType(file.Header.Get("Content-Type")) {
		errorResponse(c, http.StatusBadRequest, "不支持的文件类型")
		return
	}

	// 创建上传目录
	uploadDir := "./uploads/shops"
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		log2.Errorf("创建上传目录失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, "创建上传目录失败")
		return
	}

	// 如果已有图片，先删除旧图片
	if shop.ImageURL != "" {
		oldImagePath := strings.TrimPrefix(shop.ImageURL, "/")
		if err := os.Remove(oldImagePath); err != nil && !os.IsNotExist(err) {
			log2.Errorf("删除旧图片失败: %v", err)
		}
	}

	// 生成新文件名
	filename := fmt.Sprintf("shop_%d_%d%s",
		id,
		time.Now().Unix(),
		filepath.Ext(file.Filename))

	// 构建图片URL和文件路径
	filePath := fmt.Sprintf("%s/%s", uploadDir, filename)

	// 保存上传文件
	if err := c.SaveUploadedFile(file, filePath); err != nil {
		log2.Errorf("保存文件失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, "保存文件失败")
		return
	}
	// 压缩图片
	compressedSize, err := utils.CompressImage(filePath, maxZipSize)
	if err != nil {
		log2.Errorf("压缩图片失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, "压缩图片失败")
		return
	}

	if compressedSize > 0 {
		log2.Infof("图片压缩成功，原始大小: %d 字节，压缩后: %d 字节", file.Size, compressedSize)
	}

	// 更新店铺图片URL
	if err := h.DB.Model(&shop).Update("image_url", filename).Error; err != nil {
		log2.Errorf("更新店铺图片失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, "更新店铺图片失败")
		return
	}

	// 构建响应消息
	message := "图片更新成功"
	if shop.ImageURL == "" {
		message = "图片上传成功"
	}

	operationType := "update"
	if message == "图片上传成功" {
		operationType = "create"
	}

	c.JSON(http.StatusOK, gin.H{
		"message": message,
		"url":     filename,
		"type":    operationType,
	})
}

// GetShopImage 获取店铺图片
// @Summary 获取店铺图片
// @Description 获取指定店铺的图片
// @Tags 店铺管理
// @Accept json
// @Produce json
// @Param shopId query string true "店铺ID"
// @Success 200 {object} map[string]interface{} "查询成功"
// @Security BearerAuth
// @Router /admin/shop/image [get]
// @Router /shopOwner/shop/image [get]
// @Router /shop/image [get]
func (h *Handler) GetShopImage(c *gin.Context) {
	fileName := c.Query("path")
	if fileName == "" {
		errorResponse(c, http.StatusBadRequest, "缺少图片路径")
		return
	}

	if err := utils.ValidateImageURL(fileName, "shop"); err != nil {
		log2.Errorf("图片路径验证失败: %v", err)
		errorResponse(c, http.StatusBadRequest, "无效的图片路径")
		return
	}

	imagePath := fmt.Sprintf("./uploads/shops/%s", fileName)

	if _, err := os.Stat(imagePath); os.IsNotExist(err) {
		log2.Errorf("图片文件不存在: %s", imagePath)
		errorResponse(c, http.StatusNotFound, "图片不存在")
		return
	}

	c.File(imagePath)
}

// GetShopTempToken 获取店铺的临时令牌
// @Summary 获取店铺临时令牌
// @Description 为店铺生成临时访问令牌
// @Tags 店铺管理
// @Accept json
// @Produce json
// @Param shopId query string true "店铺ID"
// @Success 200 {object} map[string]interface{} "查询成功"
// @Security BearerAuth
// @Router /admin/shop/temp-token [get]
func (h *Handler) GetShopTempToken(c *gin.Context) {
	// 从URL参数中获取shopID
	shopIDStr := c.Query("shop_id")
	shopIDUint64, err := strconv.ParseUint(shopIDStr, 10, 64)
	if err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的店铺ID")
		return
	}
	shopID := snowflake.ID(shopIDUint64)

	// 验证店铺是否存在
	var shop models.Shop
	if err := h.DB.Where("id = ?", shopID).First(&shop).Error; err != nil {
		errorResponse(c, http.StatusNotFound, "店铺不存在")
		return
	}

	// 获取有效令牌
	token, err := h.tempTokenService.GetValidTempToken(uint64(shopID))
	if err != nil {
		errorResponse(c, http.StatusInternalServerError, "获取临时令牌失败")
		return
	}

	successResponse(c, gin.H{
		"shop_id":    token.ShopID,
		"token":      token.Token,
		"expires_at": token.ExpiresAt,
	})
}

// UpdateOrderStatusFlow 更新店铺订单流转状态配置
// @Summary 更新订单状态流转
// @Description 更新店铺的订单状态流转配置
// @Tags 店铺管理
// @Accept json
// @Produce json
// @Param flow body OrderStatusFlowRequest true "状态流转信息"
// @Success 200 {object} map[string]interface{} "更新成功"
// @Security BearerAuth
// @Router /admin/shop/update-order-status-flow [put]
// @Router /shopOwner/shop/update-order-status-flow [put]
func (h *Handler) UpdateOrderStatusFlow(c *gin.Context) {
	var req struct {
		ShopID          uint64                 `json:"shop_id" binding:"required"`
		OrderStatusFlow models.OrderStatusFlow `json:"order_status_flow" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求数据")
		return
	}

	// 查询店铺是否存在
	shop, err := h.productRepo.GetShopByID(req.ShopID)
	if err != nil {
		errorResponse(c, http.StatusNotFound, "店铺不存在")
		return
	}

	// 更新订单流转状态配置
	shop.OrderStatusFlow = req.OrderStatusFlow

	if err := h.DB.Save(&shop).Error; err != nil {
		h.logger.Errorf("更新店铺订单流转状态配置失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, "更新店铺订单流转状态配置失败")
		return
	}

	successResponse(c, gin.H{
		"code":    200,
		"message": "店铺订单流转状态配置更新成功",
		"data": gin.H{
			"shop_id":           shop.ID,
			"order_status_flow": shop.OrderStatusFlow,
		},
	})
}
