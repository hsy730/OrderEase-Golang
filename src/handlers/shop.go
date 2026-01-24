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

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/datatypes"
	"gorm.io/gorm"

	"orderease/models"
	"orderease/utils"
	"orderease/utils/log2"
)

// GetShopTags 获取店铺标签列表
func (h *Handler) GetShopTags(c *gin.Context) {
	shopID, err := strconv.ParseUint(c.Param("shopId"), 10, 64)
	if err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的店铺ID")
		return
	}

	tags, err := h.productRepo.GetShopTagsByID(shopID)
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

	// 调用 repository 获取店铺列表
	shops, total, err := h.shopRepo.GetShopList(page, pageSize, search)
	if err != nil {
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

	// 对密码进行哈希
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(shopData.OwnerPassword), bcrypt.DefaultCost)
	if err != nil {
		h.logger.Errorf("密码加密失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, "创建店铺失败")
		return
	}

	newShop := models.Shop{
		Name:            shopData.Name,
		OwnerUsername:   shopData.OwnerUsername,
		OwnerPassword:   string(hashedPassword),
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
	shop, err := h.shopRepo.GetShopByID(updateData.ID)
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
		// 对密码进行哈希
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(*updateData.OwnerPassword), bcrypt.DefaultCost)
		if err != nil {
			h.logger.Errorf("密码加密失败: %v", err)
			errorResponse(c, http.StatusInternalServerError, "更新店铺失败")
			return
		}
		shop.OwnerPassword = string(hashedPassword)
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
func (h *Handler) CheckShopNameExists(c *gin.Context) {
	shopName := c.Query("name")
	if shopName == "" {
		errorResponse(c, http.StatusBadRequest, "商店名称不能为空")
		return
	}

	exists, err := h.shopRepo.CheckShopNameExists(shopName)
	if err != nil {
		h.logger.Errorf("检查商店名称失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, "检查商店名称失败")
		return
	}

	successResponse(c, gin.H{
		"exists": exists,
	})
}

// 上传店铺图片
func (h *Handler) UploadShopImage(c *gin.Context) {
	// 限制文件大小
	const maxFileSize = 2 * 1024 * 1024 // 2MB
	const maxZipSize = 512 * 1024
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxFileSize)

	id, err := strconv.ParseUint(c.Query("id"), 10, 64)
	if err != nil {
		errorResponse(c, http.StatusBadRequest, "缺少店铺ID")
		return
	}

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

// 获取店铺图片
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
func (h *Handler) GetShopTempToken(c *gin.Context) {
	// 从URL参数中获取shopID
	shopIDStr := c.Query("shop_id")
	shopID, err := strconv.ParseUint(shopIDStr, 10, 64)
	if err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的店铺ID")
		return
	}

	// 验证店铺是否存在
	var shop models.Shop
	if err := h.DB.Where("id = ?", shopID).First(&shop).Error; err != nil {
		errorResponse(c, http.StatusNotFound, "店铺不存在")
		return
	}

	// 获取有效令牌
	token, err := h.tempTokenService.GetValidTempToken(shopID)
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
	shop, err := h.shopRepo.GetShopByID(req.ShopID)
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
