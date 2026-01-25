package handlers

import (
	"errors"
	"net/http"
	"orderease/models"
	value_objects "orderease/domain/shared/value_objects"
	"orderease/domain/user"
	"orderease/utils"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

// 创建用户请求结构体
type CreateUserRequest struct {
	Name     string `json:"name" binding:"required"`
	Password string `json:"password" binding:"required"`
	Phone    string `json:"phone"`
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

	// 验证用户类型
	if req.Type != models.UserTypeDelivery && req.Type != models.UserTypePickup {
		errorResponse(c, http.StatusBadRequest, "无效的用户类型")
		return
	}

	if req.Phone != "" { // 电话选填
		// 增强版手机号验证
		if !utils.ValidatePhoneWithRegex(req.Phone) {
			h.logger.Errorf("无效的手机号格式: %s", req.Phone)
			errorResponse(c, http.StatusBadRequest, "手机号必须为11位数字且以1开头")
			return
		}

		// 检查手机号唯一性
		exists, err := h.userRepo.CheckPhoneExists(req.Phone)
		if err != nil {
			h.logger.Errorf("检查手机号失败: %v", err)
			errorResponse(c, http.StatusInternalServerError, "检查手机号失败")
			return
		}
		if exists {
			errorResponse(c, http.StatusConflict, "该手机号已注册")
			return
		}
	}

	// 检查用户名唯一性
	exists, err := h.userRepo.CheckUsernameExists(req.Name)
	if err != nil {
		h.logger.Errorf("检查用户名失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, "检查用户名失败")
		return
	}
	if exists {
		errorResponse(c, http.StatusConflict, "用户名已存在")
		return
	}

	// 使用 Domain 值对象验证密码（保持与现有行为一致：宽松规则）
	_, err = value_objects.NewPassword(req.Password)
	if err != nil {
		h.logger.Errorf("密码验证失败: %v", err)
		errorResponse(c, http.StatusBadRequest, "密码长度必须在6-20位且包含字母和数字")
		return
	}

	// 对密码进行哈希
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		h.logger.Errorf("密码加密失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, "创建用户失败")
		return
	}

	// 创建用户对象
	user := models.User{
		ID:       utils.GenerateSnowflakeID(),
		Name:     req.Name,
		Phone:    req.Phone,
		Password: string(hashedPassword), // 存储哈希后的密码
		Type:     req.Type,
		Role:     req.Role,    // 明确设置默认值
		Address:  req.Address, // 初始化地址字段
	}

	if err := h.userRepo.Create(&user); err != nil {
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

// 获取用户列表
func (h *Handler) GetUsers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	search := c.Query("search") // 获取用户名搜索参数

	if err := ValidatePaginationParams(page, pageSize); err != nil {
		errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	users, total, err := h.userRepo.GetUsers(page, pageSize, search)
	if err != nil {
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

	user, err := h.userRepo.GetUserByID(id)
	if err != nil {
		errorResponse(c, http.StatusNotFound, "用户未找到")
		return
	}

	successResponse(c, user)
}

// 检查用户名是否存在
func (h *Handler) CheckUsernameExists(c *gin.Context) {
	username := c.Query("username")
	if username == "" {
		errorResponse(c, http.StatusBadRequest, "用户名不能为空")
		return
	}

	exists, err := h.userRepo.CheckUsernameExists(username)
	if err != nil {
		errorResponse(c, http.StatusInternalServerError, "检查用户名失败")
		return
	}

	successResponse(c, gin.H{
		"exists": exists,
	})
}

// 更新用户信息
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
	user, err := h.userRepo.GetUserByID(id)
	if err != nil {
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
	// 处理密码更新：如果密码不为空字符串，则验证并哈希
	if updateData.Password != "" {
		// 使用 Domain 值对象验证密码
		_, err = value_objects.NewPassword(updateData.Password)
		if err != nil {
			h.logger.Errorf("密码验证失败: %v", err)
			errorResponse(c, http.StatusBadRequest, "密码长度必须在6-20位且包含字母和数字")
			return
		}
		// 对密码进行哈希
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(updateData.Password), bcrypt.DefaultCost)
		if err != nil {
			h.logger.Errorf("密码加密失败: %v", err)
			errorResponse(c, http.StatusInternalServerError, "更新用户失败")
			return
		}
		user.Password = string(hashedPassword)
	}
	if updateData.Role != "" {
		user.Role = updateData.Role
	}

	if err := h.userRepo.Update(user); err != nil {
		h.logger.Errorf("更新用户失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, "更新用户失败")
		return
	}

	// 重新获取更新后的用户信息
	user, err = h.userRepo.GetUserByID(id)
	if err != nil {
		h.logger.Errorf("获取更新后的用户信息失败: %v", err)
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

	user, err := h.userRepo.GetUserByID(id)
	if err != nil {
		h.logger.Errorf("删除用户失败, ID: %s, 错误: %v", id, err)
		errorResponse(c, http.StatusNotFound, "用户不存在")
		return
	}

	if err := h.userRepo.Delete(user); err != nil {
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

// 获取简单用户列表（只返回ID和名称）
func (h *Handler) GetUserSimpleList(c *gin.Context) {
	// 获取分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))

	// 校验分页参数
	if err := ValidatePaginationParams(page, pageSize); err != nil {
		errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	// 获取搜索关键词
	search := c.Query("search")

	// 调用repository
	users, total, err := h.userRepo.GetUserSimpleList(page, pageSize, search)
	if err != nil {
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

	// 调用 Domain Service
	userDomain, err := h.userDomain.RegisterWithPasswordValidation(user.RegisterWithPasswordValidationDTO{
		Username: req.Username,
		Password: req.Password, // 传递明文密码，由 Domain 层处理
	})
	if err != nil {
		if errors.Is(err, user.ErrUsernameAlreadyExists) {
			errorResponse(c, http.StatusConflict, "用户名已存在")
		} else if errors.Is(err, user.ErrInvalidPassword) {
			errorResponse(c, http.StatusBadRequest, "密码必须为6-20位，且包含字母和数字")
		} else {
			h.logger.Errorf("注册失败: %v", err)
			errorResponse(c, http.StatusInternalServerError, "注册失败")
		}
		return
	}

	// 转换为 Model 以获取正确格式的 ID
	userModel := userDomain.ToModel()

	// 返回注册成功信息
	responseData := gin.H{
		"message": "注册成功",
		"user": gin.H{
			"id":   userModel.ID,
			"name": userDomain.Name(),
			"type": userDomain.UserType(),
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
	user, err := h.userRepo.GetByUsername(req.Username)
	if err != nil {
		errorResponse(c, http.StatusUnauthorized, "用户名或密码错误")
		return
	}

	// 验证密码（使用bcrypt验证加密后的密码）
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		errorResponse(c, http.StatusUnauthorized, "用户名或密码错误")
		return
	}

	// 生成token
	token, expiredAt, err := utils.GenerateToken(uint64(user.ID), user.Name)
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
