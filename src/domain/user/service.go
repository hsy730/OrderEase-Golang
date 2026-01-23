package user

import (
	"errors"
)

// Service 用户领域服务
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

	// 5. 创建用户实体（值对象会自动验证）
	user, err := NewUser(
		dto.Username,
		dto.Phone,
		dto.Password,
		UserType(dto.UserType),
		UserRole(dto.Role),
	)
	if err != nil {
		if errors.Is(err, ErrInvalidPassword) {
			return nil, ErrInvalidPassword
		}
		return nil, err
	}

	// 6. 持久化
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
