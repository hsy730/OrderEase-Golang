package handlers

import (
	"fmt"
	"net/http"
	"orderease/contexts/thirdparty/domain/oauth"
	"orderease/contexts/thirdparty/infrastructure/config"
	"orderease/contexts/thirdparty/infrastructure/external/wechat"
	"orderease/contexts/thirdparty/infrastructure/persistence/repositories"
	"orderease/models"
	"orderease/utils"
	"orderease/utils/log2"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// MiniUserInfo 小程序用户信息
type MiniUserInfo struct {
	NickName      string `json:"nickName"`
	AvatarURL     string `json:"avatarUrl"`
	Gender        int    `json:"gender"`
	EncryptedData string `json:"encryptedData"`
	IV            string `json:"iv"`
	RawData       string `json:"rawData"`
	Signature     string `json:"signature"`
}

// MiniProgramLoginRequest 小程序登录请求
type MiniProgramLoginRequest struct {
	Code    string       `json:"code" binding:"required"`
	UserInfo MiniUserInfo `json:"userInfo"`
}

// MiniProgramAuthHandler 小程序认证处理器
type MiniProgramAuthHandler struct {
	db                *gorm.DB
	miniProgramClient *wechat.MiniProgramClient
	config            *config.MiniProgramConfig
	bindingRepo       *repositories.UserThirdpartyBindingRepository
}

// NewMiniProgramAuthHandler 创建小程序认证处理器
func NewMiniProgramAuthHandler(db *gorm.DB) (*MiniProgramAuthHandler, error) {
	miniConfig := config.LoadMiniProgramConfig()
	if err := miniConfig.Validate(); err != nil {
		return nil, fmt.Errorf("invalid miniprogram config: %w", err)
	}

	miniClient := wechat.NewMiniProgramClient(miniConfig.AppID, miniConfig.AppSecret)
	bindingRepo := repositories.NewUserThirdpartyBindingRepository(db)

	return &MiniProgramAuthHandler{
		db:                db,
		miniProgramClient: miniClient,
		config:            miniConfig,
		bindingRepo:       bindingRepo,
	}, nil
}

// WeChatMiniProgramLogin 微信小程序登录
// POST /api/order-ease/v1/user/wechat-login
func (h *MiniProgramAuthHandler) WeChatMiniProgramLogin(c *gin.Context) {
	var req MiniProgramLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log2.Errorf("参数绑定失败: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "无效的参数",
			"error":   err.Error(),
		})
		return
	}

	log2.Debugf("微信小程序登录请求: code=%s, nickName=%s", req.Code, req.UserInfo.NickName)

	// 1. 通过 code 换取 openid 和 session_key
	sessionInfo, err := h.miniProgramClient.Code2Session(c.Request.Context(), req.Code)
	if err != nil {
		log2.Errorf("Code2Session 失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "微信登录失败",
			"error":   "获取用户信息失败",
		})
		return
	}

	log2.Debugf("获取到 openid: %s", sessionInfo.OpenID)

	// 2. 查找或创建用户
	user, isNewUser, err := h.findOrCreateUser(sessionInfo, &req.UserInfo)
	if err != nil {
		log2.Errorf("查找或创建用户失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "登录失败",
			"error":   "用户处理失败",
		})
		return
	}

	// 3. 生成 JWT token
	token, expiredAt, err := utils.GenerateToken(uint64(user.ID), user.Name)
	if err != nil {
		log2.Errorf("生成 token 失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "登录失败",
			"error":   "生成令牌失败",
		})
		return
	}

	log2.Infof("微信小程序登录成功: ID=%d, OpenID=%s, isNewUser=%v", user.ID, sessionInfo.OpenID, isNewUser)

	// 4. 返回登录结果
	c.JSON(http.StatusOK, gin.H{
		"message": "登录成功",
		"data": gin.H{
			"user": gin.H{
				"id":         user.ID,
				"name":       user.Name,
				"nickname":   user.Nickname,
				"avatarUrl":  h.extractAvatarURL(user, sessionInfo.OpenID),
				"gender":     h.extractGender(user),
				"role":       user.Role,
				"type":       user.Type,
				"created_at": user.CreatedAt.Format(time.RFC3339),
			},
			"token":       token,
			"expiredAt":   expiredAt.Unix(),
			"first_login": isNewUser,
		},
	})
}

// findOrCreateUser 查找或创建用户
func (h *MiniProgramAuthHandler) findOrCreateUser(sessionInfo *wechat.SessionInfo, userInfo *MiniUserInfo) (*models.User, bool, error) {
	providerUserID := sessionInfo.OpenID

	// 1. 先通过绑定表查找用户
	binding, err := h.bindingRepo.FindByProviderAndUserID(oauth.ProviderWeChat, providerUserID)
	if err == nil && binding != nil {
		// 找到绑定，获取用户
		var user models.User
		if err := h.db.First(&user, binding.UserID).Error; err != nil {
			return nil, false, fmt.Errorf("find user by binding failed: %w", err)
		}

		// 更新绑定信息
		now := time.Now()
		binding.LastLoginAt = &now
		if sessionInfo.UnionID != "" && binding.UnionID == "" {
			binding.UnionID = sessionInfo.UnionID
		}
		if userInfo.NickName != "" && binding.Nickname != userInfo.NickName {
			binding.Nickname = userInfo.NickName
		}
		if userInfo.AvatarURL != "" && binding.AvatarURL != userInfo.AvatarURL {
			binding.AvatarURL = userInfo.AvatarURL
		}
		if err := h.bindingRepo.Update(binding); err != nil {
			log2.Warnf("update binding failed: %v", err)
		}

		return &user, false, nil
	}

	if err != gorm.ErrRecordNotFound {
		return nil, false, fmt.Errorf("query binding failed: %w", err)
	}

	// 2. 用户不存在，创建新用户
	username := h.generateUsername(sessionInfo.OpenID, userInfo.NickName)

	user := &models.User{
		ID:       utils.GenerateSnowflakeID(),
		Name:     username,
		Nickname: userInfo.NickName,
		Type:     "public_user",
		Role:     "public_user",
	}

	// 开始事务
	tx := h.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 创建用户
	if err := tx.Create(user).Error; err != nil {
		tx.Rollback()
		return nil, false, fmt.Errorf("create user failed: %w", err)
	}

	// 创建绑定
	binding = &models.UserThirdpartyBinding{
		UserID:         uint64(user.ID),
		Provider:       oauth.ProviderWeChat.String(),
		ProviderUserID: providerUserID,
		UnionID:        sessionInfo.UnionID,
		Nickname:       userInfo.NickName,
		AvatarURL:      userInfo.AvatarURL,
		Gender:         userInfo.Gender,
		Metadata:       h.buildMetadata(userInfo),
		IsActive:       true,
		LastLoginAt:    &[]time.Time{time.Now()}[0],
	}

	if err := tx.Create(binding).Error; err != nil {
		tx.Rollback()
		return nil, false, fmt.Errorf("create binding failed: %w", err)
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return nil, false, fmt.Errorf("commit transaction failed: %w", err)
	}

	log2.Infof("创建新微信用户: ID=%d, OpenID=%s, Username=%s", user.ID, providerUserID, username)

	return user, true, nil
}

// generateUsername 生成用户名
func (h *MiniProgramAuthHandler) generateUsername(openID, nickName string) string {
	if nickName != "" {
		// 使用昵称 + OpenID 后6位 避免重复
		suffix := ""
		if len(openID) >= 6 {
			suffix = openID[len(openID)-6:]
		}
		return fmt.Sprintf("wx_%s_%s", nickName, suffix)
	}
	// 使用 OpenID 后8位
	if len(openID) >= 8 {
		return fmt.Sprintf("wx_user_%s", openID[len(openID)-8:])
	}
	return fmt.Sprintf("wx_user_%s", openID)
}

// extractAvatarURL 提取头像 URL
func (h *MiniProgramAuthHandler) extractAvatarURL(user *models.User, openID string) string {
	var binding models.UserThirdpartyBinding
	if err := h.db.Where("user_id = ? AND provider = ?", uint64(user.ID), oauth.ProviderWeChat.String()).
		First(&binding).Error; err == nil {
		if binding.AvatarURL != "" {
			return binding.AvatarURL
		}
	}
	return ""
}

// extractGender 提取性别
func (h *MiniProgramAuthHandler) extractGender(user *models.User) int {
	// 从绑定表查询
	var binding models.UserThirdpartyBinding
	if err := h.db.Where("user_id = ? AND provider = ?", uint64(user.ID), oauth.ProviderWeChat.String()).First(&binding).Error; err == nil {
		return binding.Gender
	}
	return 0
}

// buildMetadata 构建 metadata
func (h *MiniProgramAuthHandler) buildMetadata(userInfo *MiniUserInfo) models.Metadata {
	metadata := make(models.Metadata)
	metadata["encrypted_data"] = userInfo.EncryptedData
	metadata["signature"] = userInfo.Signature
	metadata["raw_data"] = userInfo.RawData
	metadata["login_at"] = time.Now().Unix()
	return metadata
}
