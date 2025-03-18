package handlers

import (
	"net/http"
	"orderease/models"
	"orderease/utils"
	"strconv"

	"github.com/gin-gonic/gin"
)

// 创建用户
func (h *Handler) CreateUser(c *gin.Context) {
	var user models.User
	if err := c.ShouldBindJSON(&user); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的用户数据: "+err.Error())
		return
	}

	// 验证用户类型
	if user.Type != models.UserTypeDelivery && user.Type != models.UserTypePickup {
		errorResponse(c, http.StatusBadRequest, "无效的用户类型")
		return
	}

	// 验证手机号
	if !isValidPhone(user.Phone) {
		errorResponse(c, http.StatusBadRequest, "无效的手机号")
		return
	}

	// 生成用户ID
	user.ID = utils.GenerateSnowflakeID()
	if err := h.DB.Create(&user).Error; err != nil {
		h.logger.Printf("创建用户失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, "创建用户失败")
		return
	}

	successResponse(c, user)
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

	if err := h.DB.Model(&models.User{}).Count(&total).Error; err != nil {
		h.logger.Printf("获取用户总数失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, "获取用户列表失败")
		return
	}

	offset := (page - 1) * pageSize
	if err := h.DB.Offset(offset).Limit(pageSize).Find(&users).Error; err != nil {
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

	var user models.User
	if err := h.DB.First(&user, id).Error; err != nil {
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

	if err := h.DB.Model(&models.User{}).Select("id, name").Find(&users).Error; err != nil {
		h.logger.Printf("查询用户列表失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, "获取用户列表失败")
		return
	}

	successResponse(c, users)
}
