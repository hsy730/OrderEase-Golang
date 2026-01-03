package handlers

import (
	"net/http"
	"orderease/models"
	"orderease/utils"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

// CreateUserRequest 创建用户请求结构体
type CreateUserRequest struct {
	Name     string `json:"name" binding:"required" example:"张三"`
	Password string `json:"password" binding:"required" example:"password123"`
	Phone    string `json:"phone" example:"13800138000"`
	Type     string `json:"type" binding:"required,oneof=delivery pickup" example:"delivery"`
	Address  string `json:"address" example:"北京市朝阳区"`
	Role     string `json:"role" example:"public"`
}

// UpdateUserRequest 更新用户请求结构体
type UpdateUserRequest struct {
	ID       string `json:"id" binding:"required" example:"1"`
	Type     string `json:"type" example:"delivery"`
	Phone    string `json:"phone" example:"13800138000"`
	Password string `json:"password" example:"newpass123"`
	Address  string `json:"address" example:"北京市朝阳区"`
	Role     string `json:"role" example:"public"`
}

// CreateUser 创建用户
// @Summary 创建用户
// @Description 创建新用户
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param user body CreateUserRequest true "用户信息"
// @Success 200 {object} map[string]interface{} "创建成功"
// @Security BearerAuth
// @Router /admin/user/create [post]
// @Router /shopOwner/user/create [post]
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

	if user.Phone != "" { // 电话选填
		// 增强版手机号验证
		if !utils.ValidatePhoneWithRegex(user.Phone) {
			h.logger.Errorf("无效的手机号格式: %s", user.Phone)
			errorResponse(c, http.StatusBadRequest, "手机号必须为11位数字且以1开头")
			return
		}

		// 检查手机号唯一性
		var existingUser models.User
		if h.DB.Where("phone = ?", user.Phone).First(&existingUser).Error == nil {
			errorResponse(c, http.StatusConflict, "该手机号已注册")
			return
		}
	}

	// 生成用户ID
	user.ID = utils.GenerateSnowflakeID()

	if err := h.DB.Create(&user).Error; err != nil {
		h.logger.Errorf("创建用户失败: %v", err)
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

// GetUsers 获取用户列表
// @Summary 获取用户列表
// @Description 获取用户列表，支持分页和筛选
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param page query int false "页码"
// @Param pageSize query int false "每页数量"
// @Param status query string false "用户状态"
// @Success 200 {object} map[string]interface{} "查询成功"
// @Security BearerAuth
// @Router /admin/user/list [get]
// @Router /shopOwner/user/list [get]
func (h *Handler) GetUsers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	search := c.Query("search") // 获取用户名搜索参数

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

	baseQuery := h.DB.Model(&models.User{})

	// 如果提供了用户名参数，则添加模糊匹配条件
	if search != "" {
		baseQuery = baseQuery.Where("name LIKE ?", "%"+search+"%")
	}

	if err := baseQuery.Count(&total).Error; err != nil {
		h.logger.Errorf("获取用户总数失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, "获取用户列表失败")
		return
	}

	offset := (page - 1) * pageSize
	if err := baseQuery.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&users).Error; err != nil {
		h.logger.Errorf("查询用户列表失败: %v", err)
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

// GetUser 获取用户详情
// @Summary 获取用户详情
// @Description 获取指定用户的详细信息
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param userId query string true "用户ID"
// @Success 200 {object} map[string]interface{} "查询成功"
// @Security BearerAuth
// @Router /admin/user/detail [get]
// @Router /shopOwner/user/detail [get]
func (h *Handler) GetUser(c *gin.Context) {
	id := c.Query("id")
	if id == "" {
		errorResponse(c, http.StatusBadRequest, "缺少用户ID")
		return
	}

	// requestShopID, err := strconv.ParseUint(c.Query("shop_id"), 10, 64)
	// if err != nil {
	// 	errorResponse(c, http.StatusBadRequest, "无效的店铺ID")
	// 	return
	// }

	// validShopID, err := h.validAndReturnShopID(c, requestShopID)
	// if err != nil {
	// 	errorResponse(c, http.StatusBadRequest, err.Error())
	// 	return
	// }

	var user models.User
	if err := h.DB.First(&user, id).Error; err != nil {
		h.logger.Errorf("查询用户失败, ID: %s, 错误: %v", id, err)
		errorResponse(c, http.StatusNotFound, "用户未找到")
		return
	}

	successResponse(c, user)
}

// CheckUsernameExists 检查用户名是否存在
// @Summary 检查用户名是否存在
// @Description 检查用户名是否已被注册
// @Tags 用户
// @Accept json
// @Produce json
// @Param username query string true "用户名"
// @Success 200 {object} map[string]interface{} "查询成功"
// @Security BearerAuth
// @Router /user/check-username [get]
func (h *Handler) CheckUsernameExists(c *gin.Context) {
	username := c.Query("username")
	if username == "" {
		errorResponse(c, http.StatusBadRequest, "用户名不能为空")
		return
	}

	var user models.User
	err := h.DB.Where("name = ?", username).First(&user).Error
	if err != nil {
		// 用户不存在，返回false
		successResponse(c, gin.H{
			"exists": false,
		})
		return
	}

	// 用户存在，返回true
	successResponse(c, gin.H{
		"exists": true,
	})
}

// UpdateUser 更新用户信息
// @Summary 更新用户信息
// @Description 更新用户基本信息
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param user body UpdateUserRequest true "用户信息"
// @Success 200 {object} map[string]interface{} "更新成功"
// @Security BearerAuth
// @Router /admin/user/update [put]
// @Router /shopOwner/user/update [put]
func (h *Handler) UpdateUser(c *gin.Context) {
	// 定义更新数据结构体
	var updateData struct {
		ID       string `json:"id"`
		Type     string `json:"type"`
		Phone    string `json:"phone"`
		Password string `json:"password"`
		Address  string `json:"address"`
		Role     string `json:"role"`
		// 其他需要更新的字段
	}

	if err := c.ShouldBindJSON(&updateData); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的更新数据: "+err.Error())
		return
	}

	id := updateData.ID
	if id == "" {
		errorResponse(c, http.StatusBadRequest, "缺少用户ID")
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

	// 验证角色
	if updateData.Role != "" && updateData.Role != models.UserRolePrivate && updateData.Role != models.UserRolePublic {
		errorResponse(c, http.StatusBadRequest, "无效的角色")
		return
	}

	// 查询现有用户
	var user models.User
	if err := h.DB.First(&user, id).Error; err != nil {
		h.logger.Errorf("更新用户失败, ID: %s, 错误: %v", id, err)
		errorResponse(c, http.StatusNotFound, "用户未找到")
		return
	}

	// 更新字段
	if updateData.Type != "" {
		user.Type = updateData.Type
	}
	if updateData.Phone != "" {
		user.Phone = updateData.Phone
	}
	if updateData.Address != "" {
		user.Address = updateData.Address
	}
	// 处理密码更新：如果密码不为空字符串，则更新密码
	if updateData.Password != "" {
		user.Password = updateData.Password
	}
	if updateData.Role != "" {
		user.Role = updateData.Role
	}

	if err := h.DB.Save(&user).Error; err != nil {
		h.logger.Errorf("更新用户失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, "更新用户失败")
		return
	}

	// 重新获取更新后的用户信息
	if err := h.DB.First(&user, id).Error; err != nil {
		h.logger.Errorf("获取更新后的用户信息失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, "获取更新后的用户信息失败")
		return
	}

	successResponse(c, user)
}

// DeleteUser 删除用户
// @Summary 删除用户
// @Description 删除指定用户
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param userId query string true "用户ID"
// @Success 200 {object} map[string]interface{} "删除成功"
// @Security BearerAuth
// @Router /admin/user/delete [delete]
// @Router /shopOwner/user/delete [delete]
func (h *Handler) DeleteUser(c *gin.Context) {
	id := c.Query("id")
	if id == "" {
		errorResponse(c, http.StatusBadRequest, "缺少用户ID")
		return
	}

	// requestShopID, err := strconv.ParseUint(c.Query("shop_id"), 10, 64)
	// if err != nil {
	// 	errorResponse(c, http.StatusBadRequest, "无效的店铺ID")
	// 	return
	// }

	// validShopID, err := h.validAndReturnShopID(c, requestShopID)
	// if err != nil {
	// 	errorResponse(c, http.StatusBadRequest, err.Error())
	// 	return
	// }

	var user models.User
	if err := h.DB.First(&user, id).Error; err != nil {
		h.logger.Errorf("删除用户失败, ID: %s, 错误: %v", id, err)
		errorResponse(c, http.StatusNotFound, "用户不存在")
		return
	}

	if err := h.DB.Delete(&user).Error; err != nil {
		h.logger.Errorf("删除用户记录失败: %v", err)
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

// GetUserSimpleList 获取简单用户列表（只返回ID和名称）
// @Summary 获取用户简单列表
// @Description 获取用户简单信息列表，用于下拉选择等场景
// @Tags 用户管理
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{} "查询成功"
// @Security BearerAuth
// @Router /admin/user/simple-list [get]
// @Router /shopOwner/user/simple-list [get]
func (h *Handler) GetUserSimpleList(c *gin.Context) {
	var users []struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}
	// 获取分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))

	// 校验分页参数
	if page < 1 {
		errorResponse(c, http.StatusBadRequest, "页码必须大于0")
		return
	}

	if pageSize < 1 || pageSize > 100 {
		errorResponse(c, http.StatusBadRequest, "每页数量必须在1-100之间")
		return
	}

	// 获取搜索关键词
	search := c.Query("search")

	// 构建查询
	query := h.DB.Model(&models.User{}).Select("id, name")

	// 如果有搜索关键词，添加模糊搜索条件
	if search != "" {
		query = query.Where("name LIKE ?", "%"+search+"%")
	}

	// 获取总数
	var total int64
	if err := query.Model(&models.User{}).Count(&total).Error; err != nil {
		h.logger.Errorf("查询用户总数失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, "获取用户列表失败")
		return
	}

	// 计算偏移量
	offset := (page - 1) * pageSize

	// 查询分页数据
	if err := query.Offset(offset).Limit(pageSize).Find(&users).Error; err != nil {
		h.logger.Errorf("查询用户列表失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, "获取用户列表失败")
		return
	}

	if err := query.Find(&users).Error; err != nil {
		h.logger.Errorf("查询用户列表失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, "获取用户列表失败")
		return
	}

	successResponse(c, users)
}

// 前端用户注册请求结构体
type FrontendUserRegisterRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required,min=6,max=20"`
}

// 前端用户注册
// @Summary 前端用户注册
// @Description 前端用户注册接口，密码为6位字母或数字
// @Tags 前端用户
// @Accept json
// @Produce json
// @Param request body FrontendUserRegisterRequest true "注册信息"
// @Success 200 {object} map[string]interface{} "注册成功"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 409 {object} map[string]interface{} "用户名或手机号已存在"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /user/register [post]
func (h *Handler) FrontendUserRegister(c *gin.Context) {
	req := FrontendUserRegisterRequest{}
	if err := c.ShouldBindJSON(&req); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的注册数据: "+err.Error())
		return
	}

	// 验证密码格式：6位字母或数字
	if !isValidPassword(req.Password) {
		errorResponse(c, http.StatusBadRequest, "密码为6位以上字母或数字")
		return
	}

	// 检查用户名是否已存在
	var existingUser models.User
	if h.DB.Where("name = ?", req.Username).First(&existingUser).Error == nil {
		errorResponse(c, http.StatusConflict, "用户名已存在")
		return
	}

	// 创建用户对象
	user := models.User{
		ID:       utils.GenerateSnowflakeID(),
		Name:     req.Username,
		Password: req.Password,            // 存储明文密码（6位字母或数字）
		Type:     models.UserTypeDelivery, // 默认邮寄配送
		Role:     models.UserRolePublic,   // 默认公开用户
	}

	if err := h.DB.Create(&user).Error; err != nil {
		h.logger.Errorf("创建用户失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, "注册失败")
		return
	}

	// 返回注册成功信息，移除敏感字段
	responseData := gin.H{
		"message": "注册成功",
		"user": gin.H{
			"id":   user.ID,
			"name": user.Name,
			"type": user.Type,
		},
	}
	successResponse(c, responseData)
}

// 前端用户登录请求结构体
type FrontendUserLoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// 前端用户登录
// @Summary 前端用户登录
// @Description 前端用户登录接口
// @Tags 前端用户
// @Accept json
// @Produce json
// @Param request body FrontendUserLoginRequest true "登录信息"
// @Success 200 {object} map[string]interface{} "登录成功"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 401 {object} map[string]interface{} "用户名或密码错误"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /user/login [post]
func (h *Handler) FrontendUserLogin(c *gin.Context) {
	req := FrontendUserLoginRequest{}
	if err := c.ShouldBindJSON(&req); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的登录数据: "+err.Error())
		return
	}

	// 查询用户
	var user models.User
	if err := h.DB.Where("name = ?", req.Username).First(&user).Error; err != nil {
		errorResponse(c, http.StatusUnauthorized, "用户名或密码错误")
		return
	}

	// 验证密码（使用bcrypt验证加密后的密码）
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		errorResponse(c, http.StatusUnauthorized, "用户名或密码错误")
		return
	}

	// 生成token
	token, expiredAt, err := utils.GenerateToken(uint64(user.ID), user.Name, false)
	if err != nil {
		h.logger.Errorf("生成token失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, "登录失败")
		return
	}

	// 返回登录成功信息
	responseData := gin.H{
		"message": "登录成功",
		"user": gin.H{
			"id":   user.ID,
			"name": user.Name,
			"type": user.Type,
		},
		"token":     token,
		"expiredAt": expiredAt.Unix(),
	}
	successResponse(c, responseData)
}

// 验证密码格式：最少6位，必须包含数字或字母，允许特殊字符
func isValidPassword(password string) bool {
	if len(password) < 6 {
		return false
	}
	// 检查是否至少包含一个字母或数字
	for _, c := range password {
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') {
			return true
		}
	}
	return false
}
