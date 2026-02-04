package user

import "fmt"

// CreateUserRequest 创建用户请求 DTO
type CreateUserRequest struct {
	Name     string `json:"name" binding:"required"`
	Password string `json:"password" binding:"required"`
	Phone    string `json:"phone"`
	Type     string `json:"type" binding:"required,oneof=delivery pickup"`
	Address  string `json:"address"`
	Role     string `json:"role"`
}

// Validate 验证创建用户请求
func (r *CreateUserRequest) Validate() error {
	if r.Name == "" {
		return fmt.Errorf("用户名不能为空")
	}
	if r.Password == "" {
		return fmt.Errorf("密码不能为空")
	}
	if r.Type != "delivery" && r.Type != "pickup" {
		return fmt.Errorf("用户类型必须是 delivery 或 pickup")
	}
	return nil
}

// FrontendUserRegisterRequest 前端用户注册请求 DTO
type FrontendUserRegisterRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required,min=6,max=20"`
}

// Validate 验证前端用户注册请求
func (r *FrontendUserRegisterRequest) Validate() error {
	if r.Username == "" {
		return fmt.Errorf("用户名不能为空")
	}
	if r.Password == "" {
		return fmt.Errorf("密码不能为空")
	}
	if len(r.Password) < 6 || len(r.Password) > 20 {
		return fmt.Errorf("密码长度必须在6-20位之间")
	}
	return nil
}

// FrontendUserLoginRequest 前端用户登录请求 DTO
type FrontendUserLoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// Validate 验证前端用户登录请求
func (r *FrontendUserLoginRequest) Validate() error {
	if r.Username == "" {
		return fmt.Errorf("用户名不能为空")
	}
	if r.Password == "" {
		return fmt.Errorf("密码不能为空")
	}
	return nil
}
