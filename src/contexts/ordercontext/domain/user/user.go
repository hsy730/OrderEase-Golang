// Package user 提供用户领域模型的核心业务逻辑。
//
// 职责：
//   - 用户生命周期管理（注册、信息更新）
//   - 用户认证（密码验证）
//   - 用户信息验证（手机号、密码格式）
//
// 业务规则：
//   - 手机号使用 Phone 值对象验证（11位，1开头）
//   - 密码使用 Password 值对象验证（6-20位，字母+数字）
//   - 前端用户使用 SimplePassword（6位）
//   - 密码使用 bcrypt 哈希存储
//
// 用户类型：
//   - delivery: 邮寄配送
//   - pickup:   自提
//
// 用户角色：
//   - private_user: 私有用户
//   - public_user:  公开用户
//
// 使用示例：
//
//	// 创建普通用户
//	user, err := user.NewUser("张三", "13800138000", "Pass123", user.UserTypeDelivery, user.UserRolePublic)
//
//	// 验证密码
//	if err := user.VerifyPassword(inputPassword); err != nil {
//	    return errors.New("密码错误")
//	}
package user

import (
	"errors"
	"strings"

	"github.com/bwmarrin/snowflake"
	"golang.org/x/crypto/bcrypt"
	"orderease/contexts/ordercontext/domain/shared/value_objects"
	"orderease/models"
	"orderease/utils"
)

// User 用户实体（充血模型）
//
// 充血模型特点：
//   - 封装业务逻辑（密码验证、格式校验）
//   - 通过值对象保证数据有效性
//   - 自包含业务规则
//
// 约束：
//   - phone 使用 Phone 值对象，保证格式正确
//   - password 使用 Password 值对象，保证强度
//   - id 由系统自动生成，不可修改
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

// NewUser 创建用户实体（带完整验证）
//
// 参数：
//   - name:     用户姓名
//   - phone:    手机号（11位，1开头）
//   - password: 密码（6-20位，必须包含字母和数字）
//   - userType: 用户类型（delivery/pickup）
//   - role:     用户角色（private_user/public_user）
//
// 返回：
//   - *User: 创建成功的用户实体
//   - error: 验证失败（手机号格式/密码强度）
//
// 验证流程：
//   1. 手机号格式验证（Phone 值对象）
//   2. 密码强度验证（Password 值对象）
//   3. 生成唯一用户ID
//
// 使用场景：
//   - 后端用户注册
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

// NewSimpleUser 创建简单用户（前端用户注册专用）
//
// 参数：
//   - name:     用户姓名
//   - password: 简单密码（6位，纯数字）
//
// 返回：
//   - *User: 创建成功的用户实体
//   - error: 验证失败
//
// 特点：
//   - 无手机号要求（Phone 为空值对象）
//   - 6位简单密码（SimplePassword 值对象）
//   - 默认类型：delivery
//   - 默认角色：public_user
//
// 使用场景：
//   - 前端用户快速注册
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

// ValidatePassword 验证密码格式（不比对）
//
// 参数：
//   - plainPassword: 明文密码
//
// 返回：
//   - nil: 密码格式符合要求
//   - error: 密码格式错误
//
// 与 VerifyPassword 的区别：
//   - ValidatePassword: 仅验证格式，用于注册时
//   - VerifyPassword:   比对哈希值，用于登录时
func (u *User) ValidatePassword(plainPassword string) error {
	_, err := value_objects.NewPassword(plainPassword)
	return err
}

// HasPhone 是否有手机号
func (u *User) HasPhone() bool {
	return !u.phone.IsEmpty()
}

// VerifyPassword 验证用户密码（登录用）
//
// 参数：
//   - plainPassword: 明文密码
//
// 返回：
//   - nil: 密码正确
//   - error: 密码错误或哈希无效
//
// 验证逻辑：
//   1. 获取存储的密码哈希
//   2. 如果是明文（开发环境），直接比对
//   3. 如果是哈希，使用 bcrypt 比对
//
// 使用场景：
//   - 用户登录验证
//   - 敏感操作二次验证
func (u *User) VerifyPassword(plainPassword string) error {
	hashedPassword := u.password.String()

	// 如果密码未哈希（开发测试环境），直接比对
	if !strings.HasPrefix(hashedPassword, "$2a$") && !strings.HasPrefix(hashedPassword, "$2b$") {
		if hashedPassword == plainPassword {
			return nil
		}
		return errors.New("密码错误")
	}

	// 使用 bcrypt 验证哈希密码
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(plainPassword))
}

// ToModel 转换为持久化模型
//
// 转换过程：
//   1. 解析或生成用户ID（snowflake ID）
//   2. 密码哈希处理（如果未哈希）
//   3. 映射所有字段到 models.User
//
// 安全处理：
//   - 如果密码不是 bcrypt 哈希格式，自动进行哈希
//   - 支持开发环境明文密码（自动转换）
//
// 返回：可直接用于 GORM 创建的 models.User
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
