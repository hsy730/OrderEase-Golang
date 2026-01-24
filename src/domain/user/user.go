package user

import (
	"strings"

	"github.com/bwmarrin/snowflake"
	"golang.org/x/crypto/bcrypt"
	"orderease/domain/shared/value_objects"
	"orderease/models"
	"orderease/utils"
)

// User 用户实体（充血模型）
type User struct {
	id       UserID
	name     string
	phone    value_objects.Phone
	password value_objects.Password
	userType UserType
	role     UserRole
	address  string
}

// UserID 用户ID类型
type UserID string

// NewUserID 创建新的用户ID
func NewUserID() UserID {
	return UserID(utils.GenerateSnowflakeID().String())
}

// NewUserIDFromSnowflake 从雪花ID创建UserID
func NewUserIDFromSnowflake(id snowflake.ID) UserID {
	return UserID(id.String())
}

// UserType 用户类型
type UserType string

const (
	UserTypeDelivery UserType = "delivery" // 邮寄配送
	UserTypePickup   UserType = "pickup"   // 自提
)

// UserRole 用户角色
type UserRole string

const (
	UserRolePrivate UserRole = "private_user" // 私有用户
	UserRolePublic  UserRole = "public_user"  // 公开用户
)

// NewUser 创建用户实体（带验证）
func NewUser(name string, phone string, password string, userType UserType, role UserRole) (*User, error) {
	// 使用值对象进行验证
	phoneVO, err := value_objects.NewPhone(phone)
	if err != nil {
		return nil, err
	}

	passwordVO, err := value_objects.NewPassword(password)
	if err != nil {
		return nil, err
	}

	return &User{
		id:       NewUserID(),
		name:     name,
		phone:    phoneVO,
		password: passwordVO,
		userType: userType,
		role:     role,
	}, nil
}

// NewSimpleUser 创建简单用户（用于前端注册，6位密码）
func NewSimpleUser(name string, password string) (*User, error) {
	// 前端用户无手机号，使用空Phone值对象
	phoneVO, _ := value_objects.NewPhone("")

	// 使用简单密码验证
	passwordVO, err := value_objects.NewSimplePassword(password)
	if err != nil {
		return nil, err
	}

	return &User{
		id:       NewUserID(),
		name:     name,
		phone:    phoneVO,
		password: passwordVO,
		userType: UserTypeDelivery,
		role:     UserRolePublic,
	}, nil
}

// Getters 方法

func (u *User) ID() UserID {
	return u.id
}

func (u *User) Name() string {
	return u.name
}

func (u *User) Phone() string {
	return u.phone.String()
}

func (u *User) Password() string {
	return u.password.String()
}

func (u *User) UserType() UserType {
	return u.userType
}

func (u *User) Role() UserRole {
	return u.role
}

func (u *User) Address() string {
	return u.address
}

// Setters 方法

func (u *User) SetName(name string) {
	u.name = name
}

func (u *User) SetPhone(phone string) error {
	phoneVO, err := value_objects.NewPhone(phone)
	if err != nil {
		return err
	}
	u.phone = phoneVO
	return nil
}

func (u *User) SetPassword(password string) error {
	passwordVO, err := value_objects.NewPassword(password)
	if err != nil {
		return err
	}
	u.password = passwordVO
	return nil
}

func (u *User) SetUserType(userType UserType) {
	u.userType = userType
}

func (u *User) SetRole(role UserRole) {
	u.role = role
}

func (u *User) SetAddress(address string) {
	u.address = address
}

// 业务方法

// ValidatePassword 验证密码
func (u *User) ValidatePassword(plainPassword string) error {
	_, err := value_objects.NewPassword(plainPassword)
	return err
}

// HasPhone 是否有手机号
func (u *User) HasPhone() bool {
	return !u.phone.IsEmpty()
}

// ToModel 转换为持久化模型（用于保存到数据库）
func (u *User) ToModel() *models.User {
	id := utils.GenerateSnowflakeID()
	if u.id != "" {
		// 如果已有ID，解析它
		if parsedID, err := snowflake.ParseString(string(u.id)); err == nil {
			id = parsedID
		}
	}

	// 对密码进行哈希（如果不是哈希值）
	password := u.password.String()
	if !strings.HasPrefix(password, "$2a$") && !strings.HasPrefix(password, "$2b$") {
		hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err == nil {
			password = string(hashed)
		}
	}

	return &models.User{
		ID:       id,
		Name:     u.name,
		Phone:    u.phone.String(),
		Password: password,
		Type:     string(u.userType),
		Role:     string(u.role),
		Address:  u.address,
	}
}

// UserFromModel 从持久化模型创建领域实体
func UserFromModel(model *models.User) *User {
	return &User{
		id:       NewUserIDFromSnowflake(model.ID),
		name:     model.Name,
		phone:    value_objects.Phone(model.Phone),
		password: value_objects.Password(model.Password),
		userType: UserType(model.Type),
		role:     UserRole(model.Role),
		address:  model.Address,
	}
}
