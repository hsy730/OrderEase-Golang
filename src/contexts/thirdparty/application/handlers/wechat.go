package handlers

import (
	"fmt"
	"net/http"
	"orderease/contexts/thirdparty/domain/oauth"
	"orderease/contexts/thirdparty/infrastructure/external/wechat"
	wechatservice "orderease/contexts/thirdparty/domain/wechat"
	"orderease/contexts/thirdparty/infrastructure/config"
	"orderease/contexts/thirdparty/infrastructure/persistence/repositories"
	"orderease/models"
	"orderease/utils"
	"orderease/utils/log2"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// WeChatHandler 微信 OAuth 处理器
type WeChatHandler struct {
	db           *gorm.DB
	oauthService *wechatservice.Service
	stateRepo    *repositories.OAuthStateRepository
	bindingRepo  *repositories.UserThirdpartyBindingRepository
	jwtService   *JWTService
	config       *config.WeChatConfig
}

// NewWeChatHandler 创建微信 OAuth 处理器
func NewWeChatHandler(db *gorm.DB) (*WeChatHandler, error) {
	// 加载微信配置
	wechatConfig := config.LoadWeChatConfig()
	if err := wechatConfig.Validate(); err != nil {
		return nil, fmt.Errorf("invalid wechat config: %w", err)
	}

	// 创建微信 API 客户端
	apiClient := wechat.NewClient(wechatConfig.AppID, wechatConfig.AppSecret)

	// 创建微信 OAuth 领域服务
	oauthService, err := wechatservice.NewService(&wechatservice.Config{
		AppID:       wechatConfig.AppID,
		AppSecret:   wechatConfig.AppSecret,
		RedirectURI: wechatConfig.RedirectURI,
		Scope:       wechatConfig.Scope,
	}, apiClient)
	if err != nil {
		return nil, fmt.Errorf("create oauth service failed: %w", err)
	}

	// 创建 State 仓储
	stateRepo := repositories.NewOAuthStateRepository(db)

	// 创建绑定仓储
	bindingRepo := repositories.NewUserThirdpartyBindingRepository(db)

	// 创建 JWT 服务
	jwtService := NewJWTService()

	return &WeChatHandler{
		db:           db,
		oauthService: oauthService,
		stateRepo:    stateRepo,
		bindingRepo:  bindingRepo,
		jwtService:   jwtService,
		config:       wechatConfig,
	}, nil
}

// Authorize 获取微信授权 URL
// GET /api/order-ease/v1/thirdparty/wechat/authorize
func (h *WeChatHandler) Authorize(c *gin.Context) {
	// 生成 state 参数（防 CSRF）
	state := h.stateRepo.Generate()

	// 存储 state（10分钟有效）
	if err := h.stateRepo.Save(state, oauth.ProviderWeChat, 600); err != nil {
		log2.Errorf("save oauth state failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "生成授权状态失败",
		})
		return
	}

	// 获取授权 URL
	authURL := h.oauthService.GetAuthorizeURL(state, h.config.RedirectURI)

	c.JSON(http.StatusOK, gin.H{
		"authorize_url": authURL,
		"state":         state,
	})
}

// Callback 处理微信授权回调
// GET /api/order-ease/v1/thirdparty/wechat/callback
func (h *WeChatHandler) Callback(c *gin.Context) {
	code := c.Query("code")
	state := c.Query("state")

	log2.Debugf("WeChat callback: code=%s, state=%s", code, state)

	// 1. 验证 state
	if err := h.stateRepo.VerifyAndDelete(state); err != nil {
		log2.Errorf("verify state failed: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的授权状态",
		})
		return
	}

	// 2. 通过 code 获取 OpenID
	result, err := h.oauthService.HandleCallback(c.Request.Context(), code, state)
	if err != nil {
		log2.Errorf("handle callback failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "获取授权信息失败",
		})
		return
	}

	log2.Debugf("OAuth result: openID=%s, unionID=%s", result.OpenID, result.UnionID)

	// 3. 查找或创建用户
	user, err := h.findOrCreateUser(result)
	if err != nil {
		log2.Errorf("find or create user failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "用户处理失败",
		})
		return
	}

	log2.Debugf("User: ID=%d, Name=%s", user.ID, user.Name)

	// 4. 生成 JWT token
	token, _, err := h.jwtService.GenerateToken(user)
	if err != nil {
		log2.Errorf("generate token failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "生成令牌失败",
		})
		return
	}

	// 5. 构建前端回调 URL
	frontendURL := fmt.Sprintf("/order-ease-iui/wechat-login-callback?token=%s", token)

	// 6. 重定向到前端
	c.Redirect(http.StatusFound, frontendURL)
}

// findOrCreateUser 查找或创建用户（使用通用绑定表）
func (h *WeChatHandler) findOrCreateUser(result *oauth.OAuthResult) (*models.User, error) {
	// 1. 先通过绑定表查找用户
	binding, err := h.bindingRepo.FindByProviderAndUserID(oauth.ProviderWeChat, result.OpenID)
	if err == nil && binding != nil {
		// 找到绑定，获取用户
		var user models.User
		if err := h.db.First(&user, binding.UserID).Error; err != nil {
			return nil, fmt.Errorf("find user by binding failed: %w", err)
		}

		// 更新绑定信息
		now := time.Now()
		binding.LastLoginAt = &now
		if result.UnionID != "" && binding.UnionID == "" {
			binding.UnionID = result.UnionID
		}
		// 更新 metadata
		if binding.Metadata == nil {
			binding.Metadata = make(models.Metadata)
		}
		if result.AccessToken != "" {
			binding.Metadata.SetAccessToken(result.AccessToken)
		}
		if result.RefreshToken != "" {
			binding.Metadata.SetRefreshToken(result.RefreshToken)
		}
		// 更新第三方用户信息
		if nickname, ok := result.RawData["nickname"].(string); ok {
			binding.Nickname = nickname
		}
		if avatar, ok := result.RawData["headimgurl"].(string); ok {
			binding.AvatarURL = avatar
		}

		if err := h.bindingRepo.Update(binding); err != nil {
			log2.Warnf("update binding failed: %v", err)
		}

		return &user, nil
	}

	if err != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("query binding failed: %w", err)
	}

	// 2. 用户不存在，创建新用户和新绑定
	user := &models.User{
		ID:   utils.GenerateSnowflakeID(),
		Name: h.generateUserName(result),
		Type: "public_user",
		Role: "public_user",
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
		return nil, fmt.Errorf("create user failed: %w", err)
	}

	// 创建绑定
	binding = &models.UserThirdpartyBinding{
		UserID:         uint64(user.ID),
		Provider:       oauth.ProviderWeChat.String(),
		ProviderUserID: result.OpenID,
		UnionID:        result.UnionID,
		Nickname:       h.extractNickname(result),
		AvatarURL:      h.extractAvatar(result),
		Metadata:       h.buildMetadata(result),
		IsActive:       true,
		LastLoginAt:    &[]time.Time{time.Now()}[0],
	}

	if err := tx.Create(binding).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("create binding failed: %w", err)
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("commit transaction failed: %w", err)
	}

	log2.Infof("Created new user from WeChat: ID=%d, OpenID=%s, Name=%s", user.ID, result.OpenID, user.Name)

	return user, nil
}

// generateUserName 生成用户名
func (h *WeChatHandler) generateUserName(result *oauth.OAuthResult) string {
	// 优先使用第三方昵称
	if nickname, ok := result.RawData["nickname"].(string); ok && nickname != "" {
		return nickname
	}
	// 使用 OpenID 后6位作为默认用户名
	if len(result.OpenID) >= 6 {
		return fmt.Sprintf("微信用户_%s", result.OpenID[len(result.OpenID)-6:])
	}
	return fmt.Sprintf("微信用户_%s", result.OpenID)
}

// extractNickname 提取昵称
func (h *WeChatHandler) extractNickname(result *oauth.OAuthResult) string {
	if nickname, ok := result.RawData["nickname"].(string); ok {
		return nickname
	}
	return ""
}

// extractAvatar 提取头像
func (h *WeChatHandler) extractAvatar(result *oauth.OAuthResult) string {
	if avatar, ok := result.RawData["headimgurl"].(string); ok {
		return avatar
	}
	return ""
}

// buildMetadata 构建 metadata
func (h *WeChatHandler) buildMetadata(result *oauth.OAuthResult) models.Metadata {
	metadata := make(models.Metadata)
	if result.AccessToken != "" {
		metadata.SetAccessToken(result.AccessToken)
	}
	if result.RefreshToken != "" {
		metadata.SetRefreshToken(result.RefreshToken)
	}
	if result.ExpiresIn > 0 {
		metadata["expires_in"] = result.ExpiresIn
	}
	metadata["token_obtained_at"] = time.Now().Unix()
	// 保存其他原始数据
	for k, v := range result.RawData {
		if k != "nickname" && k != "headimgurl" { // 这些字段已经单独存储
			metadata[k] = v
		}
	}
	return metadata
}

// GetConfig 获取微信配置（用于前端判断是否显示微信登录按钮）
// GET /api/order-ease/v1/thirdparty/wechat/config
func (h *WeChatHandler) GetConfig(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"enabled": h.config.Enabled,
	})
}

// LoginByOpenID 通过 OpenID 直接登录（用于个人公众号）
// POST /api/order-ease/v1/thirdparty/wechat/login-by-openid
func (h *WeChatHandler) LoginByOpenID(c *gin.Context) {
	var req struct {
		OpenID   string `json:"open_id" binding:"required"`
		Nickname string `json:"nickname"` // 可选，首次登录时的昵称
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		log2.Errorf("参数绑定失败: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的参数",
		})
		return
	}

	// 添加微信前缀
	username := fmt.Sprintf("wx:%s", req.OpenID)

	log2.Debugf("Login by OpenID: openID=%s, username=%s, nickname=%s", req.OpenID, username, req.Nickname)

	// 查找或创建用户
	user, err := h.findOrCreateUserByName(username, req.Nickname)
	if err != nil {
		log2.Errorf("find or create user failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "用户处理失败",
		})
		return
	}

	// 生成 JWT token
	token, expiresIn, err := h.jwtService.GenerateToken(user)
	if err != nil {
		log2.Errorf("generate token failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "生成令牌失败",
		})
		return
	}

	log2.Infof("User logged in by OpenID: ID=%d, username=%s, nickname=%s", user.ID, user.Name, user.Nickname)

	c.JSON(http.StatusOK, gin.H{
		"token":       token,
		"expires_in":  expiresIn,
		"first_login": user.Nickname == "",
		"user": gin.H{
			"id":       user.ID,
			"name":     user.Name,
			"nickname": user.Nickname,
			"role":     user.Role,
			"type":     user.Type,
		},
	})
}

// findOrCreateUserByName 通过用户名查找或创建用户
func (h *WeChatHandler) findOrCreateUserByName(username string, nickname string) (*models.User, error) {
	// 先查找用户
	var user models.User
	err := h.db.Where("name = ?", username).First(&user).Error

	if err == nil {
		// 用户已存在，更新昵称（如果提供）
		if nickname != "" && user.Nickname == "" {
			user.Nickname = nickname
			if err := h.db.Save(&user).Error; err != nil {
				return nil, fmt.Errorf("update user nickname failed: %w", err)
			}
			log2.Infof("Updated user nickname: ID=%d, nickname=%s", user.ID, nickname)
		}
		return &user, nil
	}

	if err != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("query user failed: %w", err)
	}

	// 用户不存在，创建新用户
	user = models.User{
		ID:       utils.GenerateSnowflakeID(),
		Name:     username,
		Nickname: nickname, // 使用提供的昵称，可能为空
		Type:     "public_user",
		Role:     "public_user",
	}

	// 如果没有昵称，使用 OpenID 后6位作为默认昵称
	if user.Nickname == "" {
		if len(username) > 8 { // wx: + 6 chars
			user.Nickname = username[len(username)-6:]
		} else {
			user.Nickname = username
		}
	}

	if err := h.db.Create(&user).Error; err != nil {
		return nil, fmt.Errorf("create user failed: %w", err)
	}

	log2.Infof("Created new user: ID=%d, Name=%s, Nickname=%s", user.ID, user.Name, user.Nickname)

	return &user, nil
}

// Unbind 解绑微信账号（需要用户登录）
// DELETE /api/order-ease/v1/thirdparty/wechat/unbind
func (h *WeChatHandler) Unbind(c *gin.Context) {
	// 从 JWT token 获取用户 ID
	// 这个接口需要认证中间件
	// TODO: 实现解绑逻辑

	c.JSON(http.StatusOK, gin.H{
		"message": "解绑功能待实现",
	})
}
