package http

import (
	"net/http"
	"orderease/application/dto"
	"orderease/application/services"
	"orderease/domain/shared"
	"orderease/utils/log2"
	"strconv"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	userService *services.UserService
}

func NewUserHandler(userService *services.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

func (h *UserHandler) CreateUser(c *gin.Context) {
	var req dto.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求数据: "+err.Error())
		return
	}

	response, err := h.userService.CreateUser(&req)
	if err != nil {
		log2.Errorf("创建用户失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	successResponse(c, response)
}

func (h *UserHandler) GetUser(c *gin.Context) {
	idStr := c.Query("id")
	if idStr == "" {
		errorResponse(c, http.StatusBadRequest, "缺少用户ID")
		return
	}

	id, err := shared.ParseIDFromString(idStr)
	if err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的用户ID")
		return
	}

	response, err := h.userService.GetUser(id)
	if err != nil {
		log2.Errorf("查询用户失败: %v", err)
		errorResponse(c, http.StatusNotFound, err.Error())
		return
	}

	successResponse(c, response)
}

func (h *UserHandler) GetUsers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	search := c.Query("search")

	if page < 1 || pageSize < 1 || pageSize > 100 {
		errorResponse(c, http.StatusBadRequest, "无效的分页参数")
		return
	}

	response, err := h.userService.GetUsers(page, pageSize, search)
	if err != nil {
		log2.Errorf("查询用户列表失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, "查询失败")
		return
	}

	successResponse(c, response)
}

func (h *UserHandler) UpdateUser(c *gin.Context) {
	idStr := c.Query("id")
	if idStr == "" {
		errorResponse(c, http.StatusBadRequest, "缺少用户ID")
		return
	}

	id, err := shared.ParseIDFromString(idStr)
	if err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的用户ID")
		return
	}

	var req dto.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求数据: "+err.Error())
		return
	}

	req.ID = id
	response, err := h.userService.UpdateUser(&req)
	if err != nil {
		log2.Errorf("更新用户失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	successResponse(c, response)
}

func (h *UserHandler) DeleteUser(c *gin.Context) {
	idStr := c.Query("id")
	if idStr == "" {
		errorResponse(c, http.StatusBadRequest, "缺少用户ID")
		return
	}

	id, err := shared.ParseIDFromString(idStr)
	if err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的用户ID")
		return
	}

	if err := h.userService.DeleteUser(id); err != nil {
		log2.Errorf("删除用户失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	successResponse(c, gin.H{"message": "用户删除成功"})
}

func (h *UserHandler) CheckFrontendUsernameExists(c *gin.Context) {
	username := c.Query("username")
	if username == "" {
		errorResponse(c, http.StatusBadRequest, "缺少用户名")
		return
	}

	exists, err := h.userService.CheckUsernameExists(username)
	if err != nil {
		log2.Errorf("检查用户名失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, "检查失败")
		return
	}

	successResponse(c, gin.H{
		"exists":    exists,
		"username":  username,
		"available": !exists,
	})
}

// GetUserSimpleList 获取用户简单列表（用于下拉选择等场景）
func (h *UserHandler) GetUserSimpleList(c *gin.Context) {
	response, err := h.userService.GetUsers(1, 1000, "")
	if err != nil {
		log2.Errorf("查询用户简单列表失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, "查询失败")
		return
	}

	// 返回简化的用户列表格式
	users := make([]gin.H, 0, len(response.Data))
	for _, user := range response.Data {
		users = append(users, gin.H{
			"id":   user.ID.String(),
			"name": user.Name,
		})
	}

	successResponse(c, users)
}
