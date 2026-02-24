// Package user (service) 提供用户领域服务。
//
// 职责：
//   - 用户注册（管理员创建、前端注册）
//   - 用户信息更新（手机号、密码）
//   - 用户查询
//
// 注册类型：
//   - Register:              管理员创建，使用强密码规则
//   - SimpleRegister:        前端快速注册，6位简单密码
//   - RegisterWithPasswordValidation: 前端注册，6-20位标准密码
//
// 验证规则：
//   - 用户名唯一性
//   - 手机号唯一性
//   - 密码强度（根据注册类型不同）
package user

import (
	"errors"

	"orderease/contexts/ordercontext/domain/shared/value_objects"
)

// Service 用户领域服务
//
// 职责边界：
//   - 用户注册完整流程（验证+创建+持久化）
//   - 用户信息更新（手机号、密码）
//   - 用户查询
//
// 依赖：
//   - Repository: 用户数据访问
//
// 注意：
//   - 不处理登录认证（认证在 Handler 层）
//   - 不处理事务（由 Repository 或 Handler 管理）
type Service struct {
	repo Repository
}

// NewService 创建用户领域服务
func NewService(repo Repository) *Service {
	return &Service{
		repo: repo,
	}
}

// RegisterUserDTO 用户注册DTO
type RegisterUserDTO struct {
	Username string
	Phone    string
	Password string
	UserType string
	Role     string
}

// SimpleRegisterDTO 简单注册DTO（用于前端用户）
type SimpleRegisterDTO struct {
	Username string
	Password string
}

// 错误定义
var (
	ErrUsernameAlreadyExists = errors.New("用户名已存在")
	ErrPhoneAlreadyExists    = errors.New("手机号已注册")
	ErrInvalidPassword       = errors.New("密码格式无效")
	ErrInvalidUserType       = errors.New("无效的用户类型")
	ErrInvalidRole           = errors.New("无效的角色")
)

// Register 用户注册（管理员创建用户）
func (s *Service) Register(dto RegisterUserDTO) (*User, error) {
	// 1. 验证用户类型
	if dto.UserType != string(UserTypeDelivery) && dto.UserType != string(UserTypePickup) {
		return nil, ErrInvalidUserType
	}

	// 2. 验证角色
	if dto.Role != string(UserRolePrivate) && dto.Role != string(UserRolePublic) {
		return nil, ErrInvalidRole
	}

	// 3. 检查用户名唯一性
	exists, err := s.repo.UsernameExists(dto.Username)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrUsernameAlreadyExists
	}

	// 4. 检查手机号唯一性
	if dto.Phone != "" {
		exists, err = s.repo.PhoneExists(dto.Phone)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, ErrPhoneAlreadyExists
		}
	}

	// 5. 验证密码（管理员创建用户使用弱密码规则：6-20位字母或数字）
	_, err = value_objects.NewWeakPassword(dto.Password)
	if err != nil {
		return nil, ErrInvalidPassword
	}

	// 6. 创建用户实体
	phoneVO, _ := value_objects.NewPhone(dto.Phone)
	passwordVO := value_objects.Password(dto.Password)
	user := &User{
		id:       NewUserID(),
		name:     dto.Username,
		phone:    phoneVO,
		password: passwordVO,
		userType: UserType(dto.UserType),
		role:     UserRole(dto.Role),
	}

	// 7. 持久化
	if err := s.repo.Create(user); err != nil {
		return nil, err
	}

	return user, nil
}

// SimpleRegister 简单注册（前端用户）
func (s *Service) SimpleRegister(dto SimpleRegisterDTO) (*User, error) {
	// 1. 检查用户名唯一性
	exists, err := s.repo.UsernameExists(dto.Username)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrUsernameAlreadyExists
	}

	// 2. 创建用户实体（使用简单密码验证）
	user, err := NewSimpleUser(dto.Username, dto.Password)
	if err != nil {
		return nil, ErrInvalidPassword
	}

	// 3. 持久化
	if err := s.repo.Create(user); err != nil {
		return nil, err
	}

	return user, nil
}

// RegisterWithPasswordValidationDTO 注册DTO（带密码验证）
type RegisterWithPasswordValidationDTO struct {
	Username string
	Password string
}

// RegisterWithPasswordValidation 前端用户注册（6-20位密码）
func (s *Service) RegisterWithPasswordValidation(dto RegisterWithPasswordValidationDTO) (*User, error) {
	// 1. 检查用户名唯一性
	exists, err := s.repo.UsernameExists(dto.Username)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrUsernameAlreadyExists
	}

	// 2. 使用弱密码规则 (6-20位 + 字母或数字)
	passwordVO, err := value_objects.NewWeakPassword(dto.Password)
	if err != nil {
		return nil, ErrInvalidPassword
	}

	// 3. 创建用户实体
	phoneVO, _ := value_objects.NewPhone("")
	user := &User{
		id:       NewUserID(),
		name:     dto.Username,
		phone:    phoneVO,
		password: passwordVO,
		userType: UserTypeDelivery,
		role:     UserRolePublic,
	}

	// 4. 持久化
	if err := s.repo.Create(user); err != nil {
		return nil, err
	}

	return user, nil
}

// GetByID 根据ID获取用户
func (s *Service) GetByID(id UserID) (*User, error) {
	return s.repo.GetByID(id)
}

// GetByUsername 根据用户名获取用户
func (s *Service) GetByUsername(username string) (*User, error) {
	return s.repo.GetByUsername(username)
}

// UpdatePhone 更新用户手机号
func (s *Service) UpdatePhone(id UserID, phone string) error {
	user, err := s.repo.GetByID(id)
	if err != nil {
		return err
	}

	// 如果新手机号与当前手机号相同，直接返回（无需更新）
	currentPhone := user.Phone()
	if currentPhone == phone {
		return nil
	}

	// 检查新手机号是否已被其他用户使用
	if phone != "" {
		exists, err := s.repo.PhoneExists(phone)
		if err != nil {
			return err
		}
		if exists {
			return ErrPhoneAlreadyExists
		}
	}

	// 使用实体的 Setter 方法（会自动验证）
	return user.SetPhone(phone)
}

// UpdatePassword 更新用户密码
func (s *Service) UpdatePassword(id UserID, newPassword string) error {
	user, err := s.repo.GetByID(id)
	if err != nil {
		return err
	}

	// 使用实体的 Setter 方法（会自动验证）
	if err := user.SetPassword(newPassword); err != nil {
		return ErrInvalidPassword
	}

	return s.repo.Update(user)
}

// DeleteUser 删除用户
func (s *Service) DeleteUser(id UserID) error {
	user, err := s.repo.GetByID(id)
	if err != nil {
		return err
	}

	return s.repo.Delete(user)
}
