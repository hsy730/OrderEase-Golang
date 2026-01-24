package shop

import (
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
	"orderease/models"
)

// Shop 店铺聚合根
type Shop struct {
	id              uint64
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
func (s *Shop) ID() uint64 {
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
func (s *Shop) SetID(id uint64) {
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
