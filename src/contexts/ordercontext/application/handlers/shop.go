package handlers

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

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
	// 从URL参数或用户上下文中获取店铺ID
	shopSnowflakeID, err := h.getShopIDFromQueryOrContext(c)
	if err != nil {
		errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}
	if shopSnowflakeID <= 0 {
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
		ID              models.SnowflakeString            `json:"id" binding:"required"`
		OwnerUsername   string                  `json:"owner_username" binding:"required"`
		OwnerPassword   *string                 `json:"owner_password"`
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

	if updateData.OrderStatusFlow != nil && updateData.OrderStatusFlow.Statuses == nil {
		updateData.OrderStatusFlow = nil
	}

	userInfo, exists := c.Get("userInfo")
	if !exists {
		errorResponse(c, http.StatusUnauthorized, "未获取到用户信息")
		return
	}
	user := userInfo.(models.UserInfo)

	if updateData.ValidUntil != "" && !user.IsAdmin {
		updateData.ValidUntil = ""
	}

	// 使用领域服务更新店铺信息（封装所有字段更新逻辑）
	updatedShop, err := h.shopService.UpdateInfo(updateData.ID.ToSnowflakeID(), shopdomain.ShopUpdates{
		Name:            updateData.Name,
		ContactPhone:    updateData.ContactPhone,
		ContactEmail:    updateData.ContactEmail,
		Description:     updateData.Description,
		Address:         updateData.Address,
		Settings:        updateData.Settings,
		OrderStatusFlow: updateData.OrderStatusFlow,
		ValidUntil:      updateData.ValidUntil,
		OwnerUsername:   updateData.OwnerUsername,
		OwnerPassword:   updateData.OwnerPassword,
	})
	if err != nil {
		h.logger.Errorf("更新店铺失败: %v", err)
		if err.Error() == "店铺不存在" {
			errorResponse(c, http.StatusNotFound, "店铺不存在")
		} else if err.Error() == "无效的有效期格式" {
			errorResponse(c, http.StatusBadRequest, err.Error())
		} else {
			errorResponse(c, http.StatusInternalServerError, "更新店铺失败")
		}
		return
	}

	successResponse(c, gin.H{
		"code": 200,
		"data": gin.H{
			"id":                updatedShop.ID,
			"name":              updatedShop.Name,
			"description":       updatedShop.Description,
			"owner_username":    updatedShop.OwnerUsername,
			"contact_phone":     updatedShop.ContactPhone,
			"address":           updatedShop.Address,
			"contact_email":     updatedShop.ContactEmail,
			"valid_until":       updatedShop.ValidUntil.Format(time.RFC3339),
			"settings":          updatedShop.Settings,
			"order_status_flow": updatedShop.OrderStatusFlow,
		},
	})
}

// 删除店铺及关联数据
func (h *Handler) DeleteShop(c *gin.Context) {
	// 从URL参数或用户上下文中获取店铺ID
	shopID, err := h.getShopIDFromQueryOrContext(c)
	if err != nil {
		errorResponse(c, http.StatusBadRequest, err.Error())
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
	// 从URL参数或用户上下文中获取店铺ID
	shopID, err := h.getShopIDFromQueryOrContext(c)
	if err != nil {
		errorResponse(c, http.StatusBadRequest, err.Error())
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
		ShopID          models.SnowflakeString           `json:"shop_id"`
		OrderStatusFlow models.OrderStatusFlow `json:"order_status_flow" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求数据")
		return
	}

	// 验证并获取有效的店铺ID（validAndReturnShopID 会自动处理店主接口的逻辑）
	validShopID, err := h.validAndReturnShopID(c, req.ShopID.ToSnowflakeID())
	if err != nil {
		errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	// 使用领域服务更新订单流转状态配置
	updatedShop, err := h.shopService.UpdateOrderStatusFlow(validShopID, req.OrderStatusFlow)
	if err != nil {
		h.logger.Errorf("更新店铺订单流转状态配置失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	successResponse(c, gin.H{
		"code":    200,
		"message": "店铺订单流转状态配置更新成功",
		"data": gin.H{
			"shop_id":           updatedShop.ID,
			"order_status_flow": updatedShop.OrderStatusFlow,
		},
	})
}
