package services

import (
	"fmt"
	"time"

	"orderease/contexts/ordercontext/infrastructure/repositories"
	"orderease/contexts/thirdparty/domain/oauth"
	"orderease/contexts/thirdparty/domain/user"
	"orderease/contexts/thirdparty/infrastructure/external/wechat"
	"orderease/models"
	"orderease/utils"
	"orderease/utils/log2"

	"gorm.io/gorm"
)

type MiniProgramAuthService struct {
	db          *gorm.DB
	userRepo    *repositories.UserRepository
	bindingRepo user.UserBindingRepository
}

func NewMiniProgramAuthService(db *gorm.DB, userRepo *repositories.UserRepository, bindingRepo user.UserBindingRepository) *MiniProgramAuthService {
	return &MiniProgramAuthService{
		db:          db,
		userRepo:    userRepo,
		bindingRepo: bindingRepo,
	}
}

type LoginResult struct {
	User     *models.User
	IsNewUser bool
}

func (s *MiniProgramAuthService) FindOrCreateUser(sessionInfo *wechat.SessionInfo, nickname, avatarURL string) (*LoginResult, error) {
	providerUserID := sessionInfo.OpenID

	binding, err := s.bindingRepo.FindByProviderAndUserID(oauth.ProviderWeChat, providerUserID)
	if err == nil && binding != nil {
		return s.handleExistingUser(binding, sessionInfo, nickname, avatarURL)
	}

	if err != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("query binding failed: %w", err)
	}

	return s.createNewUser(sessionInfo, providerUserID, nickname, avatarURL)
}

func (s *MiniProgramAuthService) handleExistingUser(binding *models.UserThirdpartyBinding, sessionInfo *wechat.SessionInfo, nickname, avatarURL string) (*LoginResult, error) {
	userModel, err := s.userRepo.GetUserByID(string(binding.UserID))
	if err != nil {
		return nil, fmt.Errorf("find user by binding failed: %w", err)
	}

	now := time.Now()
	binding.LastLoginAt = &now
	if sessionInfo.UnionID != "" && binding.UnionID == "" {
		binding.UnionID = sessionInfo.UnionID
	}
	if nickname != "" && binding.Nickname != nickname {
		binding.Nickname = nickname
	}
	if avatarURL != "" && binding.AvatarURL != avatarURL {
		binding.AvatarURL = avatarURL
	}
	if err := s.bindingRepo.Update(binding); err != nil {
		log2.Warnf("update binding failed: %v", err)
	}

	userUpdated := false
	if nickname != "" && userModel.Nickname != nickname {
		userModel.Nickname = nickname
		userUpdated = true
	}
	if avatarURL != "" && userModel.Avatar != avatarURL {
		userModel.Avatar = avatarURL
		userUpdated = true
	}
	if userUpdated {
		if err := s.userRepo.Update(userModel); err != nil {
			log2.Warnf("update user info failed: %v", err)
		} else {
			log2.Infof("Updated existing user: ID=%d, Nickname=%s, Avatar=%s", userModel.ID, userModel.Nickname, userModel.Avatar)
		}
	}

	return &LoginResult{
		User:     userModel,
		IsNewUser: false,
	}, nil
}

func (s *MiniProgramAuthService) createNewUser(sessionInfo *wechat.SessionInfo, providerUserID, nickname, avatarURL string) (*LoginResult, error) {
	username := s.generateUsername(sessionInfo.OpenID, nickname)

	user := &models.User{
		ID:       utils.GenerateSnowflakeID(),
		Name:     username,
		Nickname: nickname,
		Type:     "public_user",
		Role:     "public_user",
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
		ProviderUserID: providerUserID,
		UnionID:        sessionInfo.UnionID,
		Nickname:       nickname,
		AvatarURL:      avatarURL,
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

	log2.Infof("创建新微信用户: ID=%d, OpenID=%s, Username=%s", user.ID, providerUserID, username)

	return &LoginResult{
		User:     user,
		IsNewUser: true,
	}, nil
}

func (s *MiniProgramAuthService) FindBindingByUserIDAndProvider(userID uint64, provider oauth.Provider) (*models.UserThirdpartyBinding, error) {
	return s.bindingRepo.FindByUserIDAndProvider(userID, provider)
}

func (s *MiniProgramAuthService) generateUsername(openID, nickName string) string {
	if nickName != "" {
		suffix := ""
		if len(openID) >= 6 {
			suffix = openID[len(openID)-6:]
		}
		return fmt.Sprintf("wx_%s_%s", nickName, suffix)
	}
	if len(openID) >= 8 {
		return fmt.Sprintf("wx_user_%s", openID[len(openID)-8:])
	}
	return fmt.Sprintf("wx_user_%s", openID)
}
