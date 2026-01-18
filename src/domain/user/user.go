package user

import (
	"errors"
	"time"

	"orderease/domain/shared"
)

type UserRole string

const (
	UserRolePrivate UserRole = "private_user"
	UserRolePublic  UserRole = "public_user"
)

type UserType string

const (
	UserTypeDelivery UserType = "delivery"
	UserTypePickup   UserType = "pickup"
	UserTypeSystem   UserType = "system"
)

func (r UserRole) IsValid() bool {
	return r == UserRolePrivate || r == UserRolePublic
}

func (t UserType) IsValid() bool {
	return t == UserTypeDelivery || t == UserTypePickup || t == UserTypeSystem
}

type User struct {
	ID        shared.ID
	Name      string
	Role      UserRole
	Password  string
	Phone     string
	Address   string
	Type      UserType
	CreatedAt time.Time
	UpdatedAt time.Time
}

func NewUser(name string, role UserRole, userType UserType, password string) (*User, error) {
	if name == "" {
		return nil, errors.New("用户名不能为空")
	}

	if !role.IsValid() {
		return nil, errors.New("无效的用户角色")
	}

	if !userType.IsValid() {
		return nil, errors.New("无效的用户类型")
	}

	now := time.Now()

	return &User{
		ID:        shared.NewID(),
		Name:      name,
		Role:      role,
		Type:      userType,
		Password:  password,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

func (u *User) UpdateBasicInfo(name, phone, address string) error {
	if name != "" {
		u.Name = name
	}
	if phone != "" {
		u.Phone = phone
	}
	if address != "" {
		u.Address = address
	}
	u.UpdatedAt = time.Now()
	return nil
}

func (u *User) UpdatePassword(newPassword string) error {
	if newPassword == "" {
		return errors.New("新密码不能为空")
	}
	u.Password = newPassword
	u.UpdatedAt = time.Now()
	return nil
}

func (u *User) IsSystemUser() bool {
	return u.Type == UserTypeSystem
}

func (u *User) IsPublicUser() bool {
	return u.Role == UserRolePublic
}

func (u *User) IsPrivateUser() bool {
	return u.Role == UserRolePrivate
}
