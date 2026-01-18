package services

import (
	"fmt"
	"time"

	"gorm.io/gorm"
	"orderease/domain/shared"
	"orderease/models"
	"orderease/utils"
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
func (s *TempTokenService) CreateShopSystemUser(shopID shared.ID) (models.User, error) {
	// 检查是否已存在系统用户
	expectedName := fmt.Sprintf("shop_%s_system", shopID.String())
	var existingUser models.User
	if err := s.db.Where("type = ? AND name = ?", "system", expectedName).First(&existingUser).Error; err == nil {
		return existingUser, nil
	}

	// 创建新的系统用户
	user := models.User{
		Name:     expectedName,
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
// forceRefresh: 是否强制刷新令牌（false 则仅当令牌不存在或过期时才生成新令牌）
func (s *TempTokenService) GenerateTempToken(shopID shared.ID, forceRefresh bool) (models.TempToken, error) {
	// 为店铺创建系统用户
	user, err := s.CreateShopSystemUser(shopID)
	if err != nil {
		return models.TempToken{}, err
	}

	// 检查是否已存在该店铺的令牌
	var existingToken models.TempToken
	if err := s.db.Where("shop_id = ?", shopID.Value()).First(&existingToken).Error; err == nil {
		// 令牌已存在
		if !forceRefresh {
			// 检查现有令牌是否仍然有效
			if !utils.IsTokenExpired(existingToken.ExpiresAt) {
				// 令牌仍然有效，直接返回现有令牌
				return existingToken, nil
			}
		}

		// 强制刷新或令牌已过期，生成新令牌
		token := utils.GenerateTempToken()
		expiresAt := time.Now().Add(1 * time.Hour)

		existingToken.Token = token
		existingToken.ExpiresAt = expiresAt
		existingToken.UserID = uint64(user.ID)
		if err := s.db.Save(&existingToken).Error; err != nil {
			return models.TempToken{}, err
		}
		return existingToken, nil
	}

	// 令牌不存在，创建新令牌
	token := utils.GenerateTempToken()
	expiresAt := time.Now().Add(1 * time.Hour)

	tempToken := models.TempToken{
		ShopID:    shopID.Value(),
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
func (s *TempTokenService) GetValidTempToken(shopID shared.ID) (models.TempToken, error) {
	// 不强制刷新，仅当令牌不存在或过期时才生成新令牌
	return s.GenerateTempToken(shopID, false)
}

// ValidateTempToken 验证临时令牌是否有效
func (s *TempTokenService) ValidateTempToken(shopID shared.ID, token string) (bool, models.User, error) {
	var tempToken models.TempToken
	// 使用 shopID.Value() 获取 snowflake.ID 进行查询，与 GenerateTempToken 保持一致
	if err := s.db.Where("shop_id = ? AND token = ?", shopID.Value(), token).First(&tempToken).Error; err != nil {
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
