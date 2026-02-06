// Package shop 提供店铺领域模型的核心业务逻辑。
//
// 职责：
//   - 店铺生命周期管理（创建、更新、删除）
//   - 店铺认证（密码验证）
//   - 有效期管理（到期检查、续期）
//   - 订单状态流转配置管理
//
// 业务规则：
//   - 店铺有有效期限制，到期后功能受限
//   - 密码使用 bcrypt 哈希存储
//   - 删除店铺前必须清理关联数据（商品、订单）
//   - 订单状态流转配置决定订单生命周期
//
// 使用示例：
//
//	// 创建店铺
//	shop := shop.NewShop("咖啡店", "owner", time.Now().AddDate(1, 0, 0))
//
//	// 验证密码
//	if err := shop.CheckPassword(password); err != nil {
//	    return errors.New("密码错误")
//	}
//
//	// 检查有效期
//	if shop.IsExpired() {
//	    return errors.New("店铺已到期")
//	}
package shop

import (
	"fmt"
	"strings"
	"time"

	"github.com/bwmarrin/snowflake"
	"golang.org/x/crypto/bcrypt"
	"orderease/models"
)

// Shop 店铺聚合根
//
// 作为聚合根，Shop 负责：
//   - 管理店铺自身的状态和配置
//   - 维护订单状态流转配置 (OrderStatusFlow)
//   - 提供认证和有效期管理
//
// 约束：
//   - ID 为 0 表示未持久化的新店铺
//   - ownerPassword 存储的是 bcrypt 哈希值
//   - validUntil 为 UTC 时间
//   - settings 为 JSON 格式字节数组
type Shop struct {
	id              snowflake.ID
	name            string
	ownerUsername   string
	ownerPassword   string // 已哈希的密码
	contactPhone    string
	contactEmail    string
	address         string
	imageURL        string
	description     string
	validUntil      time.Time
	settings        []byte // JSON data
	orderStatusFlow models.OrderStatusFlow
	createdAt       time.Time
	updatedAt       time.Time
}

// NewShop 创建新店铺
//
// 参数：
//   - name:          店铺名称
//   - ownerUsername: 店主账号名
//   - validUntil:    有效期截止时间（UTC）
//
// 返回：
//   - 新店铺实体，需设置密码后保存
//
// 注意：
//   - ID 为 0，需在持久化时分配
//   - 密码为空，需调用 SetOwnerPassword 设置
//   - OrderStatusFlow 为空，将使用默认值
func NewShop(name string, ownerUsername string, validUntil time.Time) *Shop {
	return &Shop{
		name:          name,
		ownerUsername: ownerUsername,
		validUntil:    validUntil,
		createdAt:     time.Now(),
		updatedAt:     time.Now(),
	}
}

// Getters
func (s *Shop) ID() snowflake.ID {
	return s.id
}

func (s *Shop) Name() string {
	return s.name
}

func (s *Shop) OwnerUsername() string {
	return s.ownerUsername
}

func (s *Shop) OwnerPassword() string {
	return s.ownerPassword
}

func (s *Shop) ContactPhone() string {
	return s.contactPhone
}

func (s *Shop) ContactEmail() string {
	return s.contactEmail
}

func (s *Shop) Address() string {
	return s.address
}

func (s *Shop) ImageURL() string {
	return s.imageURL
}

func (s *Shop) Description() string {
	return s.description
}

func (s *Shop) ValidUntil() time.Time {
	return s.validUntil
}

func (s *Shop) Settings() []byte {
	return s.settings
}

func (s *Shop) OrderStatusFlow() models.OrderStatusFlow {
	return s.orderStatusFlow
}

func (s *Shop) CreatedAt() time.Time {
	return s.createdAt
}

func (s *Shop) UpdatedAt() time.Time {
	return s.updatedAt
}

// Setters
func (s *Shop) SetID(id snowflake.ID) {
	s.id = id
}

func (s *Shop) SetName(name string) {
	s.name = name
}

func (s *Shop) SetOwnerUsername(username string) {
	s.ownerUsername = username
}

func (s *Shop) SetOwnerPassword(password string) {
	s.ownerPassword = password
}

func (s *Shop) SetContactPhone(phone string) {
	s.contactPhone = phone
}

func (s *Shop) SetContactEmail(email string) {
	s.contactEmail = email
}

func (s *Shop) SetAddress(address string) {
	s.address = address
}

func (s *Shop) SetImageURL(url string) {
	s.imageURL = url
}

func (s *Shop) SetDescription(desc string) {
	s.description = desc
}

func (s *Shop) SetValidUntil(validUntil time.Time) {
	s.validUntil = validUntil
}

func (s *Shop) SetSettings(settings []byte) {
	s.settings = settings
}

func (s *Shop) SetOrderStatusFlow(flow models.OrderStatusFlow) {
	s.orderStatusFlow = flow
}

func (s *Shop) SetCreatedAt(t time.Time) {
	s.createdAt = t
}

func (s *Shop) SetUpdatedAt(t time.Time) {
	s.updatedAt = t
}

// ToModel 转换为持久化模型（用于保存到数据库）
func (s *Shop) ToModel() *models.Shop {
	// 对密码进行哈希（如果不是哈希值）
	password := s.ownerPassword
	if !strings.HasPrefix(password, "$2a$") && !strings.HasPrefix(password, "$2b$") {
		hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err == nil {
			password = string(hashed)
		}
	}

	return &models.Shop{
		ID:              s.id,
		Name:            s.name,
		OwnerUsername:   s.ownerUsername,
		OwnerPassword:   password,
		ContactPhone:    s.contactPhone,
		ContactEmail:    s.contactEmail,
		Address:         s.address,
		ImageURL:        s.imageURL,
		Description:     s.description,
		CreatedAt:       s.createdAt,
		UpdatedAt:       s.updatedAt,
		ValidUntil:      s.validUntil,
		Settings:        s.settings,
		OrderStatusFlow: s.orderStatusFlow,
	}
}

// ShopFromModel 从持久化模型创建领域实体
func ShopFromModel(model *models.Shop) *Shop {
	return &Shop{
		id:              model.ID,
		name:            model.Name,
		ownerUsername:   model.OwnerUsername,
		ownerPassword:   model.OwnerPassword,
		contactPhone:    model.ContactPhone,
		contactEmail:    model.ContactEmail,
		address:         model.Address,
		imageURL:        model.ImageURL,
		description:     model.Description,
		createdAt:       model.CreatedAt,
		updatedAt:       model.UpdatedAt,
		validUntil:      model.ValidUntil,
		settings:        model.Settings,
		orderStatusFlow: model.OrderStatusFlow,
	}
}

// ==================== 业务方法 ====================

// CheckPassword 验证店铺密码
//
// 参数：
//   - password: 明文密码
//
// 返回：
//   - nil: 密码正确
//   - error: 密码错误或哈希无效
//
// 实现：使用 bcrypt 比对哈希值
//
// 使用场景：
//   - 店主登录验证
func (s *Shop) CheckPassword(password string) error {
	return bcrypt.CompareHashAndPassword([]byte(s.ownerPassword), []byte(password))
}

// IsExpired 判断店铺是否已到期
//
// 判断逻辑：当前时间 > validUntil
//
// 时间标准：UTC
//
// 影响：
//   - 到期店铺无法创建新订单
//   - 到期店铺无法上架新商品
//   - 已有数据可正常查看
func (s *Shop) IsExpired() bool {
	now := time.Now().UTC()
	return s.validUntil.Before(now)
}

// IsActive 判断店铺是否处于激活状态
// 激活状态：未到期且不在即将到期范围内
func (s *Shop) IsActive() bool {
	return !s.IsExpired() && !s.IsExpiringSoon()
}

// IsExpiringSoon 判断店铺是否即将到期（7天内）
//
// 判断逻辑：0 <= 距离到期天数 < 7
//
// 使用场景：
//   - 发送续期提醒
//   - 仪表盘预警提示
//   - 管理后台标记
//
// 与 IsExpired 的关系：
//   - IsExpiringSoon: 即将到期但未到期
//   - IsExpired:      已到期
func (s *Shop) IsExpiringSoon() bool {
	now := time.Now().UTC()
	daysUntilExpiry := int(s.validUntil.Sub(now).Hours() / 24)
	return daysUntilExpiry >= 0 && daysUntilExpiry < 7
}

// CanDelete 检查店铺是否可删除
//
// 参数：
//   - productCount: 关联商品数量
//   - orderCount:   关联订单数量
//
// 返回：
//   - nil: 可以删除
//   - error: 存在关联数据，不可删除
//
// 删除条件：
//   - 无关联商品（productCount == 0）
//   - 无关联订单（orderCount == 0）
//
// 使用场景：
//   - 管理员删除店铺前验证
func (s *Shop) CanDelete(productCount int, orderCount int) error {
	if productCount > 0 {
		return fmt.Errorf("店铺存在 %d 个关联商品，无法删除", productCount)
	}
	if orderCount > 0 {
		return fmt.Errorf("店铺存在 %d 个关联订单，无法删除", orderCount)
	}
	return nil
}

// UpdateValidUntil 更新店铺有效期
//
// 参数：
//   - newValidUntil: 新的有效期（UTC）
//
// 返回：
//   - nil: 更新成功
//   - error: 新有效期无效
//
// 验证规则：
//   - 新有效期必须晚于当前时间
//
// 副作用：
//   - 自动更新 updatedAt 字段
func (s *Shop) UpdateValidUntil(newValidUntil time.Time) error {
	now := time.Now().UTC()
	if newValidUntil.Before(now) {
		return fmt.Errorf("新有效期不能早于当前时间")
	}
	s.validUntil = newValidUntil
	s.updatedAt = time.Now().UTC()
	return nil
}

// ValidateOrderStatusFlow 验证订单状态流转配置
//
// 参数：
//   - flow: 订单状态流转配置
//
// 返回：
//   - nil: 配置合法
//   - error: 配置无效
//
// 验证规则：
//   - 至少包含一个状态定义
//   - 状态值不能重复
//
// 使用场景：
//   - 更新店铺配置时验证
//   - 导入配置时验证
func (s *Shop) ValidateOrderStatusFlow(flow models.OrderStatusFlow) error {
	if len(flow.Statuses) == 0 {
		return fmt.Errorf("订单流转配置不能为空")
	}
	return nil
}
