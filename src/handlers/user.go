package handlers

import (
	"errors"
	"net/http"
	"orderease/models"
	userdomain "orderease/domain/user"
	"orderease/utils"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

// 创建用户
func (h *Handler) CreateUser(c *gin.Context) {
	req := userdomain.CreateUserRequest{}
	if err := c.ShouldBindJSON(&req); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的用户数据: "+err.Error())
		return
	}

	// 使用 Domain DTO 的验证方法
	if err := req.Validate(); err != nil {
		errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	// 设置默认角色
	role := req.Role
	if role == "" {
		role = models.UserRolePublic
	}

	// 调用 Domain Service 进行用户注册
	userDomain, err := h.userDomain.Register(userdomain.RegisterUserDTO{
		Username: req.Name,
		Phone:    req.Phone,
		Password: req.Password,
		UserType: req.Type,
		Role:     role,
	})
	if err != nil {
		if errors.Is(err, userdomain.ErrUsernameAlreadyExists) {
			errorResponse(c, http.StatusConflict, "用户名已存在")
		} else if errors.Is(err, userdomain.ErrPhoneAlreadyExists) {
			errorResponse(c, http.StatusConflict, "该手机号已注册")
		} else if errors.Is(err, userdomain.ErrInvalidPassword) {
			errorResponse(c, http.StatusBadRequest, "密码长度必须在6-20位且包含字母和数字")
		} else if errors.Is(err, userdomain.ErrInvalidUserType) {
			errorResponse(c, http.StatusBadRequest, "无效的用户类型")
		} else if errors.Is(err, userdomain.ErrInvalidRole) {
			errorResponse(c, http.StatusBadRequest, "无效的角色")
		} else {
			h.logger.Errorf("创建用户失败: %v", err)
			errorResponse(c, http.StatusInternalServerError, "创建用户失败")
		}
		return
	}

	// 转换为 Model 以获取正确格式的数据
	userModel := userDomain.ToModel()

	// 移除敏感字段后返回
	responseData := gin.H{
		"id":         userModel.ID,
		"name":       userModel.Name,
		"phone":      userModel.Phone,
		"type":       userModel.Type,
		"role":       userModel.Role,
		"created_at": userModel.CreatedAt.Format(time.RFC3339),
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

	// 验证角色
	if updateData.Role != "" && updateData.Role != models.UserRolePrivate && updateData.Role != models.UserRolePublic {
		errorResponse(c, http.StatusBadRequest, "无效的角色")
		return
	}

	// 使用 Domain Service 更新手机号（带验证和唯一性检查）
	if updateData.Phone != "" {
		userID := userdomain.UserID(id)
		if err := h.userDomain.UpdatePhone(userID, updateData.Phone); err != nil {
			if errors.Is(err, userdomain.ErrPhoneAlreadyExists) {
				errorResponse(c, http.StatusConflict, "该手机号已注册")
			} else {
				errorResponse(c, http.StatusBadRequest, err.Error())
			}
			return
		}
	}

	// 使用 Domain Service 更新密码（带验证和哈希）
	if updateData.Password != "" {
		userID := userdomain.UserID(id)
		if err := h.userDomain.UpdatePassword(userID, updateData.Password); err != nil {
			if errors.Is(err, userdomain.ErrInvalidPassword) {
				errorResponse(c, http.StatusBadRequest, "密码长度必须在6-20位且包含字母和数字")
			} else {
				errorResponse(c, http.StatusInternalServerError, "更新用户失败")
			}
			return
		}
	}

	// 查询现有用户并更新其他字段
	userModel, err := h.userRepo.GetUserByID(id)
	if err != nil {
		h.logger.Errorf("更新用户失败, ID: %s, 错误: %v", id, err)
		errorResponse(c, http.StatusNotFound, "用户未找到")
		return
	}

	// 更新其他字段
	if updateData.Type != "" {
		userModel.Type = updateData.Type
	}
	if updateData.Address != "" {
		userModel.Address = updateData.Address
	}
	if updateData.Role != "" {
		userModel.Role = updateData.Role
	}

	if err := h.userRepo.Update(userModel); err != nil {
		h.logger.Errorf("更新用户失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, "更新用户失败")
		return
	}

	// 重新获取更新后的用户信息
	userModel, err = h.userRepo.GetUserByID(id)
	if err != nil {
		h.logger.Errorf("获取更新后的用户信息失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, "获取更新后的用户信息失败")
		return
	}

	successResponse(c, userModel)
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

// 前端用户注册（使用 Domain DTO）
func (h *Handler) FrontendUserRegister(c *gin.Context) {
	req := userdomain.FrontendUserRegisterRequest{}
	if err := c.ShouldBindJSON(&req); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的注册数据: "+err.Error())
		return
	}

	// 使用 Domain DTO 的验证方法
	if err := req.Validate(); err != nil {
		errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	// 调用 Domain Service
	userDomain, err := h.userDomain.RegisterWithPasswordValidation(userdomain.RegisterWithPasswordValidationDTO{
		Username: req.Username,
		Password: req.Password, // 传递明文密码，由 Domain 层处理
	})
	if err != nil {
		if errors.Is(err, userdomain.ErrUsernameAlreadyExists) {
			errorResponse(c, http.StatusConflict, "用户名已存在")
		} else if errors.Is(err, userdomain.ErrInvalidPassword) {
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

// 前端用户登录（使用 Domain DTO）
func (h *Handler) FrontendUserLogin(c *gin.Context) {
	req := userdomain.FrontendUserLoginRequest{}
	if err := c.ShouldBindJSON(&req); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的登录数据: "+err.Error())
		return
	}

	// 使用 Domain DTO 的验证方法
	if err := req.Validate(); err != nil {
		errorResponse(c, http.StatusBadRequest, err.Error())
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
