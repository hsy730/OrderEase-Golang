package services

import (
	"fmt"
	"orderease/models"
	"orderease/utils"
	"time"

	"github.com/bwmarrin/snowflake"
	"github.com/robfig/cron/v3"
	"gorm.io/gorm"
)

// TempTokenService 临时令牌服务
type TempTokenService struct {
	db *gorm.DB
}

// NewTempTokenService 创建临时令牌服务实例
func NewTempTokenService(db *gorm.DB) *TempTokenService {
	return &TempTokenService{
		db: db,
	}
}

// CreateShopSystemUser 为店铺创建系统用户
func (s *TempTokenService) CreateShopSystemUser(shopID snowflake.ID) (models.User, error) {
	// 检查是否已存在系统用户
	var existingUser models.User
	if err := s.db.Where("type = ? AND name = ?", "system", fmt.Sprintf("shop_%d_system", shopID)).First(&existingUser).Error; err == nil {
		return existingUser, nil
	}

	// 创建新的系统用户
	user := models.User{
		Name:     fmt.Sprintf("shop_%d_system", shopID),
		Role:     models.UserRolePublic,
		Type:     "system",
		Phone:    "",
		Address:  "",
		Password: "", // 系统用户无需密码
	}

	if err := s.db.Create(&user).Error; err != nil {
		return models.User{}, err
	}

	return user, nil
}

// GenerateTempToken 为店铺生成临时令牌
func (s *TempTokenService) GenerateTempToken(shopID snowflake.ID) (models.TempToken, error) {
	// 为店铺创建系统用户
	user, err := s.CreateShopSystemUser(shopID)
	if err != nil {
		return models.TempToken{}, err
	}

	// 生成6位令牌
	token := utils.GenerateTempToken()

	// 设置过期时间为1小时后
	expiresAt := time.Now().Add(1 * time.Hour)

	// 检查是否已存在该店铺的令牌
	var existingToken models.TempToken
	if err := s.db.Where("shop_id = ?", shopID).First(&existingToken).Error; err == nil {
		// 更新现有令牌
		existingToken.Token = token
		existingToken.ExpiresAt = expiresAt
		existingToken.UserID = uint64(user.ID)
		if err := s.db.Save(&existingToken).Error; err != nil {
			return models.TempToken{}, err
		}
		return existingToken, nil
	}

	// 创建新令牌
	tempToken := models.TempToken{
		ShopID:    shopID,
		UserID:    uint64(user.ID),
		Token:     token,
		ExpiresAt: expiresAt,
	}

	if err := s.db.Create(&tempToken).Error; err != nil {
		return models.TempToken{}, err
	}

	return tempToken, nil
}

// GetValidTempToken 获取店铺的有效临时令牌，若过期则自动刷新
func (s *TempTokenService) GetValidTempToken(shopID snowflake.ID) (models.TempToken, error) {
	var token models.TempToken
	if err := s.db.Where("shop_id = ?", shopID).First(&token).Error; err != nil {
		// 令牌不存在，生成新令牌
		return s.GenerateTempToken(shopID)
	}

	// 检查令牌是否过期
	if utils.IsTokenExpired(token.ExpiresAt) {
		// 令牌过期，刷新令牌
		return s.GenerateTempToken(shopID)
	}

	return token, nil
}

// ValidateTempToken 验证临时令牌是否有效
func (s *TempTokenService) ValidateTempToken(shopID snowflake.ID, token string) (bool, models.User, error) {
	var tempToken models.TempToken
	if err := s.db.Where("shop_id = ? AND token = ?", shopID, token).First(&tempToken).Error; err != nil {
		return false, models.User{}, err
	}

	// 检查令牌是否过期
	if utils.IsTokenExpired(tempToken.ExpiresAt) {
		return false, models.User{}, fmt.Errorf("令牌已过期")
	}

	// 获取关联的系统用户
	var user models.User
	if err := s.db.Where("id = ?", tempToken.UserID).First(&user).Error; err != nil {
		return false, models.User{}, err
	}

	return true, user, nil
}

// RefreshAllTempTokens 刷新所有店铺的临时令牌
func (s *TempTokenService) RefreshAllTempTokens() error {
	// 获取所有店铺
	var shops []models.Shop
	if err := s.db.Find(&shops).Error; err != nil {
		return err
	}

	// 为每个店铺刷新令牌
	for _, shop := range shops {
		if _, err := s.GenerateTempToken(shop.ID); err != nil {
			return err
		}
	}

	return nil
}

// SetupCronJob 设置定时刷新任务
func (s *TempTokenService) SetupCronJob() {
	c := cron.New()
	
	// 每小时执行一次刷新任务
	_, err := c.AddFunc("0 * * * *", func() {
		s.RefreshAllTempTokens()
	})
	
	if err != nil {
		panic(err)
	}
	
	// 启动定时任务
	c.Start()
}
