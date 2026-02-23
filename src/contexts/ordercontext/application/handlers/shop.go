package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/bwmarrin/snowflake"
	"github.com/gin-gonic/gin"
	"gorm.io/datatypes"

	shopdomain "orderease/contexts/ordercontext/domain/shop"
	"orderease/models"
	"orderease/utils"
	"orderease/utils/log2"
)

// GetShopTags 获取店铺标签列表
func (h *Handler) GetShopTags(c *gin.Context) {
	shopID, err := utils.StringToSnowflakeID(c.Param("shopId"))
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

	// 转换店铺ID为雪花ID
	shopSnowflakeID, err := utils.StringToSnowflakeID(shopID)
	if err != nil || shopSnowflakeID <= 0 {
		errorResponse(c, http.StatusBadRequest, "无效的店铺ID")
		return
	}

	// 使用 Repository 获取店铺及其标签
	shop, err := h.shopRepo.GetWithTags(shopSnowflakeID)
	if err != nil {
		if err.Error() == "店铺不存在" {
			errorResponse(c, http.StatusNotFound, "店铺不存在")
			return
		}
		h.logger.Errorf("查询店铺失败，ID: %d，错误: %v", shopSnowflakeID, err)
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

	// 使用 Repository 检查用户名是否已存在
	exists, err := h.shopRepo.CheckUsernameExists(shopData.OwnerUsername)
	if err != nil {
		h.logger.Errorf("检查用户名失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, "检查用户名失败")
		return
	}
	if exists {
		errorResponse(c, http.StatusConflict, "店主用户名已存在")
		return
	}

	// 使用 Domain Service 处理有效期
	validUntil, err := h.shopService.ProcessValidUntil(shopData.ValidUntil)
	if err != nil {
		errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	// 使用 Domain Service 解析订单状态流转配置
	orderStatusFlow, err := h.shopService.ParseOrderStatusFlow(shopData.OrderStatusFlow)
	if err != nil {
		h.logger.Errorf("解析订单流转配置失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, "解析订单流转配置失败")
		return
	}

	// 使用 Domain 实体创建店铺（密码哈希由 ToModel() 处理）
	shopDomain := shopdomain.NewShop(shopData.Name, shopData.OwnerUsername, validUntil)
	shopDomain.SetID(utils.GenerateSnowflakeID())
	shopDomain.SetOwnerPassword(shopData.OwnerPassword)
	shopDomain.SetContactPhone(shopData.ContactPhone)
	shopDomain.SetContactEmail(shopData.ContactEmail)
	shopDomain.SetDescription(shopData.Description)
	shopDomain.SetAddress(shopData.Address)
	shopDomain.SetOrderStatusFlow(orderStatusFlow)
	if len(shopData.Settings) > 0 {
		shopDomain.SetSettings(shopData.Settings)
	}

	// 转换为 Model（自动处理密码哈希）
	newShop := shopDomain.ToModel()

	// 使用 Repository 创建店铺
	if err := h.shopRepo.Create(newShop); err != nil {
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
		ID              snowflake.ID            `json:"id" binding:"required"`
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
	// 处理密码更新：如果密码不为null，则使用 Domain 实体更新密码
	if updateData.OwnerPassword != nil {
		// 转换为 Domain 实体（shop 已经是指针，不需要再取地址）
		shopEntity := shopdomain.ShopFromModel(shop)
		// 设置明文密码
		shopEntity.SetOwnerPassword(*updateData.OwnerPassword)
		// 转换回 Model（自动处理密码哈希）
		shopModel := shopEntity.ToModel()
		// 更新密码字段
		shop.OwnerPassword = shopModel.OwnerPassword
	}

	// 使用 Repository 更新店铺
	if err := h.shopRepo.Update(shop); err != nil {
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
	shopID, err := utils.StringToSnowflakeID(c.Query("shop_id"))
	if err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的店铺ID")
		return
	}

	// 使用 Shop Domain Service 处理删除逻辑
	if err := h.shopService.DeleteShop(shopID); err != nil {
		// 根据错误类型返回不同的 HTTP 状态码
		if err.Error() == "店铺不存在" {
			errorResponse(c, http.StatusNotFound, err.Error())
			return
		}
		if err.Error() == "存在关联商品，无法删除店铺" ||
		   err.Error() == "存在关联订单，无法删除店铺" {
			errorResponse(c, http.StatusConflict, err.Error())
			return
		}
		h.logger.Errorf("删除店铺失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, "删除店铺失败")
		return
	}

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

	id, err := utils.StringToSnowflakeID(c.Query("id"))
	if err != nil {
		errorResponse(c, http.StatusBadRequest, "缺少店铺ID")
		return
	}

	// 使用 Repository 查询店铺
	shop, err := h.shopRepo.GetShopByID(id)
	if err != nil {
		if err.Error() == "店铺不存在" {
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

	// 使用 Media Service 验证文件类型
	if _, err := h.mediaService.ValidateImageType(file.Header.Get("Content-Type")); err != nil {
		errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	// 使用 Media Service 验证文件大小
	if err := h.mediaService.ValidateImageSize(file, maxFileSize); err != nil {
		errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	// 创建上传目录
	uploadDir := "./uploads/shops"
	if err := h.mediaService.CreateUploadDir(uploadDir); err != nil {
		h.logger.Errorf("创建上传目录失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, "创建上传目录失败")
		return
	}

	// 验证目录是否可写
	testFile := uploadDir + "/.write_test"
	if f, err := os.Create(testFile); err != nil {
		log2.Errorf("上传目录不可写: %v, 路径: %s", err, uploadDir)
		errorResponse(c, http.StatusInternalServerError, "上传目录不可写")
		return
	} else {
		f.Close()
		os.Remove(testFile)
	}

	// 使用 Media Service 删除旧图片
	if err := h.mediaService.RemoveOldImage(shop.ImageURL); err != nil {
		log2.Errorf("删除旧图片失败: %v", err)
	}

	// 使用 Media Service 生成文件名
	filename := h.mediaService.GenerateUniqueFileName("shop", uint64(id), file.Filename)

	// 使用 Media Service 构建文件路径
	filePath := h.mediaService.BuildFilePath(uploadDir, filename)

	// 打开源文件
	src, err := file.Open()
	if err != nil {
		log2.Errorf("打开上传文件失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, "打开上传文件失败")
		return
	}
	defer src.Close()

	// 创建目标文件
	dst, err := os.Create(filePath)
	if err != nil {
		log2.Errorf("创建目标文件失败: %v, 路径: %s", err, filePath)
		errorResponse(c, http.StatusInternalServerError, "创建目标文件失败")
		return
	}
	defer dst.Close()

	// 复制文件内容
	if _, err := dst.ReadFrom(src); err != nil {
		log2.Errorf("写入文件失败: %v, 路径: %s", err, filePath)
		errorResponse(c, http.StatusInternalServerError, "写入文件失败")
		return
	}

	// 压缩图片（继续使用 utils.CompressImage，未来可迁移到 Media Service）
	compressedSize, err := utils.CompressImage(filePath, maxZipSize)
	if err != nil {
		log2.Errorf("压缩图片失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, "压缩图片失败")
		return
	}

	if compressedSize > 0 {
		log2.Infof("图片压缩成功，原始大小: %d 字节，压缩后: %d 字节", file.Size, compressedSize)
	}

	// 使用 Repository 更新店铺图片URL
	if err := h.shopRepo.UpdateImageURL(shop.ID, filename); err != nil {
		log2.Errorf("更新店铺图片失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, "更新店铺图片失败")
		return
	}

	// 使用 Media Service 获取消息和操作类型
	message := h.mediaService.GetUploadMessage(shop.ImageURL == "" && filename == "")
	operationType := h.mediaService.GetOperationType(message)

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
	shopID, err := utils.StringToSnowflakeID(shopIDStr)
	if err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的店铺ID")
		return
	}

	// 使用 Repository 验证店铺是否存在
	_, err = h.shopRepo.GetShopByID(shopID)
	if err != nil {
		if err.Error() == "店铺不存在" {
			errorResponse(c, http.StatusNotFound, "店铺不存在")
			return
		}
		errorResponse(c, http.StatusInternalServerError, "查询店铺失败")
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
		ShopID          snowflake.ID           `json:"shop_id" binding:"required"`
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

	// 使用 Repository 更新店铺
	if err := h.shopRepo.Update(shop); err != nil {
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
