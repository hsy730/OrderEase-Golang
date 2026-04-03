package user

import (
	"fmt"
	"time"

	"orderease/contexts/thirdparty/domain/oauth"
	"orderease/models"
	"orderease/utils"
	"orderease/utils/log2"

	"gorm.io/gorm"
)

type Service struct {
	db          *gorm.DB
	bindingRepo UserBindingRepository
}

func NewService(db *gorm.DB, bindingRepo UserBindingRepository) *Service {
	return &Service{
		db:          db,
		bindingRepo: bindingRepo,
	}
}

// OAuthResultAdapter OAuth 结果适配器接口，解耦领域服务与具体的 OAuth 实现
type OAuthResultAdapter interface {
	GetOpenID() string
	GetUnionID() string
	GetAccessToken() string
	GetRefreshToken() string
	GetExpiresIn() int64
	GetRawData() map[string]interface{}
}

// FindOrCreateByOpenID 通过 OpenID 查找或创建用户（含绑定管理）
func (s *Service) FindOrCreateByOpenID(result OAuthResultAdapter) (*models.User, error) {
	binding, err := s.bindingRepo.FindByProviderAndUserID(oauth.ProviderWeChat, result.GetOpenID())
	if err == nil && binding != nil {
		return s.updateExistingUser(binding, result)
	}

	if err != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("query binding failed: %w", err)
	}

	return s.createNewUserWithBinding(result)
}

// FindOrCreateByName 通过用户名查找或创建用户
func (s *Service) FindOrCreateByName(username, nickname string) (*models.User, error) {
	var userModel models.User
	err := s.db.Where("name = ?", username).First(&userModel).Error

	if err == nil {
		if nickname != "" && userModel.Nickname == "" {
			userModel.Nickname = nickname
			if err := s.db.Save(&userModel).Error; err != nil {
				return nil, fmt.Errorf("update user nickname failed: %w", err)
			}
		}
		return &userModel, nil
	}

	if err != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("query user failed: %w", err)
	}

	userModel = models.User{
		ID:       utils.GenerateSnowflakeID(),
		Name:     username,
		Nickname: nickname,
		Type:     "public_user",
		Role:     "public_user",
	}

	if userModel.Nickname == "" {
		if len(username) > 8 {
			userModel.Nickname = username[len(username)-6:]
		} else {
			userModel.Nickname = username
		}
	}

	if err := s.db.Create(&userModel).Error; err != nil {
		return nil, fmt.Errorf("create user failed: %w", err)
	}

	return &userModel, nil
}

// updateExistingUser 更新已存在的用户信息和绑定信息
func (s *Service) updateExistingUser(binding *models.UserThirdpartyBinding, result OAuthResultAdapter) (*models.User, error) {
	var userModel models.User
	if err := s.db.First(&userModel, binding.UserID).Error; err != nil {
		return nil, fmt.Errorf("find user by binding failed: %w", err)
	}

	now := time.Now()
	binding.LastLoginAt = &now

	unionID := result.GetUnionID()
	if unionID != "" && binding.UnionID == "" {
		binding.UnionID = unionID
	}
	if binding.Metadata == nil {
		binding.Metadata = make(models.Metadata)
	}
	accessToken := result.GetAccessToken()
	if accessToken != "" {
		binding.Metadata.SetAccessToken(accessToken)
	}
	refreshToken := result.GetRefreshToken()
	if refreshToken != "" {
		binding.Metadata.SetRefreshToken(refreshToken)
	}

	rawData := result.GetRawData()
	if nickname, ok := rawData["nickname"].(string); ok {
		binding.Nickname = nickname
	}
	if avatar, ok := rawData["headimgurl"].(string); ok {
		binding.AvatarURL = avatar
	}

	if err := s.bindingRepo.Update(binding); err != nil {
		log2.Warnf("update binding failed: %v", err)
	}

	userUpdated := false
	newName := generateUserNameFromResult(result)
	if userModel.Name != newName {
		userModel.Name = newName
		userUpdated = true
	}
	newNickname := extractNicknameFromResult(result)
	if newNickname != "" && userModel.Nickname != newNickname {
		userModel.Nickname = newNickname
		userUpdated = true
	}
	newAvatar := extractAvatarFromResult(result)
	if newAvatar != "" && userModel.Avatar != newAvatar {
		userModel.Avatar = newAvatar
		userUpdated = true
	}
	if userUpdated {
		if err := s.db.Save(&userModel).Error; err != nil {
			log2.Warnf("update wechat user info failed: %v", err)
		} else {
			log2.Infof("Updated wechat user: ID=%d, Name=%s, Nickname=%s, Avatar=%s", userModel.ID, userModel.Name, userModel.Nickname, userModel.Avatar)
		}
	}

	return &userModel, nil
}

// createNewUserWithBinding 创建新用户并建立绑定关系（事务内）
func (s *Service) createNewUserWithBinding(result OAuthResultAdapter) (*models.User, error) {
	openID := result.GetOpenID()
	user := &models.User{
		ID:   utils.GenerateSnowflakeID(),
		Name: generateUserNameFromResult(result),
		Type: "public_user",
		Role: "public_user",
	}

	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := tx.Create(user).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("create user failed: %w", err)
	}

	binding := &models.UserThirdpartyBinding{
		UserID:         user.ID,
		Provider:       oauth.ProviderWeChat.String(),
		ProviderUserID: openID,
		UnionID:        result.GetUnionID(),
		Nickname:       extractNicknameFromResult(result),
		AvatarURL:      extractAvatarFromResult(result),
		Metadata:       buildMetadataFromResult(result),
		IsActive:       true,
		LastLoginAt:    &[]time.Time{time.Now()}[0],
	}

	if err := tx.Create(binding).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("create binding failed: %w", err)
	}

	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("commit transaction failed: %w", err)
	}

	log2.Infof("Created new user from WeChat: ID=%d, OpenID=%s, Name=%s", user.ID, openID, user.Name)

	return user, nil
}

func generateUserNameFromResult(result OAuthResultAdapter) string {
	rawData := result.GetRawData()
	if nickname, ok := rawData["nickname"].(string); ok && nickname != "" {
		return nickname
	}
	openID := result.GetOpenID()
	if len(openID) >= 6 {
		return fmt.Sprintf("微信用户_%s", openID[len(openID)-6:])
	}
	return fmt.Sprintf("微信用户_%s", openID)
}

func extractNicknameFromResult(result OAuthResultAdapter) string {
	rawData := result.GetRawData()
	if nickname, ok := rawData["nickname"].(string); ok {
		return nickname
	}
	return ""
}

func extractAvatarFromResult(result OAuthResultAdapter) string {
	rawData := result.GetRawData()
	if avatar, ok := rawData["headimgurl"].(string); ok {
		return avatar
	}
	return ""
}

func buildMetadataFromResult(result OAuthResultAdapter) models.Metadata {
	metadata := make(models.Metadata)
	accessToken := result.GetAccessToken()
	if accessToken != "" {
		metadata.SetAccessToken(accessToken)
	}
	refreshToken := result.GetRefreshToken()
	if refreshToken != "" {
		metadata.SetRefreshToken(refreshToken)
	}
	expiresIn := result.GetExpiresIn()
	if expiresIn > 0 {
		metadata["expires_in"] = expiresIn
	}
	metadata["token_obtained_at"] = time.Now().Unix()
	for k, v := range result.GetRawData() {
		if k != "nickname" && k != "headimgurl" {
			metadata[k] = v
		}
	}
	return metadata
}
