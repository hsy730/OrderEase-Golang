package handlers

import (
	"net/http"
	"orderease/models"
	"orderease/utils"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// 创建用户请求结构体
type CreateUserRequest struct {
	Name     string `json:"name" binding:"required"`
	Phone    string `json:"phone" binding:"required"`
	Password string `json:"password" binding:"required"`
	Type     string `json:"type" binding:"required,oneof=delivery pickup"`
	Address  string `json:"address"`
	Role     string `json:"role"`
}

// 创建用户
func (h *Handler) CreateUser(c *gin.Context) {
	req := CreateUserRequest{}
	if err := c.ShouldBindJSON(&req); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的用户数据: "+err.Error())
		return
	}
	// 创建用户对象并设置密码
	user := models.User{
		Name:     req.Name,
		Phone:    req.Phone,
		Password: req.Password, // 存储哈希后的密码
		Type:     req.Type,
		Role:     req.Role,    // 明确设置默认值
		Address:  req.Address, // 初始化地址字段
	}

	// 验证用户类型
	if user.Type != models.UserTypeDelivery && user.Type != models.UserTypePickup {
		errorResponse(c, http.StatusBadRequest, "无效的用户类型")
		return
	}

	// 增强版手机号验证
	if !utils.ValidatePhoneWithRegex(user.Phone) {
		h.logger.Printf("无效的手机号格式: %s", user.Phone)
		errorResponse(c, http.StatusBadRequest, "手机号必须为11位数字且以1开头")
		return
	}

	// 检查手机号唯一性
	var existingUser models.User
	if h.DB.Where("phone = ?", user.Phone).First(&existingUser).Error == nil {
		errorResponse(c, http.StatusConflict, "该手机号已注册")
		return
	}

	// 生成用户ID
	user.ID = utils.GenerateSnowflakeID()

	if err := h.DB.Create(&user).Error; err != nil {
		h.logger.Printf("创建用户失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, "创建用户失败")
		return
	}

	// 移除敏感字段后返回
	responseData := gin.H{
		"id":         user.ID,
		"name":       user.Name,
		"phone":      user.Phone,
		"type":       user.Type,
		"created_at": user.CreatedAt.Format(time.RFC3339),
	}
	successResponse(c, responseData)
}

// 获取用户列表
func (h *Handler) GetUsers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))

	if page < 1 {
		errorResponse(c, http.StatusBadRequest, "页码必须大于0")
		return
	}

	if pageSize < 1 || pageSize > 100 {
		errorResponse(c, http.StatusBadRequest, "每页数量必须在1-100之间")
		return
	}

	var users []models.User
	var total int64

	requestShopID, err := strconv.ParseUint(c.Query("shop_id"), 10, 64)
	if err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的店铺ID")
		return
	}

	_, err = h.validAndReturnShopID(c, requestShopID)
	if err != nil {
		errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	baseQuery := h.DB.Model(&models.User{})

	if err := baseQuery.Count(&total).Error; err != nil {
		h.logger.Printf("获取用户总数失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, "获取用户列表失败")
		return
	}

	offset := (page - 1) * pageSize
	if err := baseQuery.Offset(offset).Limit(pageSize).Find(&users).Error; err != nil {
		h.logger.Printf("查询用户列表失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, "获取用户列表失败")
		return
	}

	successResponse(c, gin.H{
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
		"data":     users,
	})
}

// 获取用户详情
func (h *Handler) GetUser(c *gin.Context) {
	id := c.Query("id")
	if id == "" {
		errorResponse(c, http.StatusBadRequest, "缺少用户ID")
		return
	}

	requestShopID, err := strconv.ParseUint(c.Query("shop_id"), 10, 64)
	if err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的店铺ID")
		return
	}

	_, err = h.validAndReturnShopID(c, requestShopID)
	if err != nil {
		errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	var user models.User
	if err := h.DB.First(&user, id).Error; err != nil {
		h.logger.Printf("查询用户失败, ID: %s, 错误: %v", id, err)
		errorResponse(c, http.StatusNotFound, "用户未找到")
		return
	}

	successResponse(c, user)
}

// 更新用户信息
func (h *Handler) UpdateUser(c *gin.Context) {
	id := c.Query("id")
	if id == "" {
		errorResponse(c, http.StatusBadRequest, "缺少用户ID")
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

	var user models.User
	if err := h.DB.Where("shop_id = ?", validShopID).First(&user, id).Error; err != nil {
		h.logger.Printf("更新用户失败, ID: %s, 错误: %v", id, err)
		errorResponse(c, http.StatusNotFound, "用户未找到")
		return
	}

	var updateData models.User
	if err := c.ShouldBindJSON(&updateData); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的更新数据: "+err.Error())
		return
	}

	// 验证用户类型
	if updateData.Type != "" && updateData.Type != models.UserTypeDelivery && updateData.Type != models.UserTypePickup {
		errorResponse(c, http.StatusBadRequest, "无效的用户类型")
		return
	}

	// 验证手机号
	if updateData.Phone != "" && !isValidPhone(updateData.Phone) {
		errorResponse(c, http.StatusBadRequest, "无效的手机号")
		return
	}

	if err := h.DB.Model(&user).Updates(updateData).Error; err != nil {
		h.logger.Printf("更新用户失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, "更新用户失败")
		return
	}

	// 重新获取更新后的用户信息
	if err := h.DB.First(&user, id).Error; err != nil {
		h.logger.Printf("获取更新后的用户信息失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, "获取更新后的用户信息失败")
		return
	}

	successResponse(c, user)
}

// 删除用户
func (h *Handler) DeleteUser(c *gin.Context) {
	id := c.Query("id")
	if id == "" {
		errorResponse(c, http.StatusBadRequest, "缺少用户ID")
		return
	}

	requestShopID, err := strconv.ParseUint(c.Query("shop_id"), 10, 64)
	if err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的店铺ID")
		return
	}

	_, err = h.validAndReturnShopID(c, requestShopID)
	if err != nil {
		errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	var user models.User
	if err := h.DB.First(&user, id).Error; err != nil {
		h.logger.Printf("删除用户失败, ID: %s, 错误: %v", id, err)
		errorResponse(c, http.StatusNotFound, "用户不存在")
		return
	}

	if err := h.DB.Delete(&user).Error; err != nil {
		h.logger.Printf("删除用户记录失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, "删除用户失败")
		return
	}

	successResponse(c, gin.H{"message": "用户删除成功"})
}

// 验证手机号
func isValidPhone(phone string) bool {
	// 简单的手机号验证：11位数字，以1开头
	if len(phone) != 11 || phone[0] != '1' {
		return false
	}
	for _, c := range phone {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

// 获取简单用户列表（只返回ID和名称）
func (h *Handler) GetUserSimpleList(c *gin.Context) {
	var users []struct {
		ID   uint   `json:"id"`
		Name string `json:"name"`
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

	if err := h.DB.Where("shop_id = ?", validShopID).Model(&models.User{}).Select("id, name").Find(&users).Error; err != nil {
		h.logger.Printf("查询用户列表失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, "获取用户列表失败")
		return
	}

	successResponse(c, users)
}
