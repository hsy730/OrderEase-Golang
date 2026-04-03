package handlers

import (
	"fmt"
	"net/http"
	"orderease/contexts/thirdparty/domain/oauth"
	thirdpartyuser "orderease/contexts/thirdparty/domain/user"
	wechatservice "orderease/contexts/thirdparty/domain/wechat"
	"orderease/contexts/thirdparty/infrastructure/config"
	"orderease/contexts/thirdparty/infrastructure/external/wechat"
	"orderease/contexts/thirdparty/infrastructure/persistence/repositories"
	"orderease/utils/log2"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// WeChatHandler 微信 OAuth 处理器
type WeChatHandler struct {
	oauthService *wechatservice.Service
	stateRepo    *repositories.OAuthStateRepository
	userService  *thirdpartyuser.Service
	jwtService   *JWTService
	config       *config.WeChatConfig
}

// NewWeChatHandler 创建微信 OAuth 处理器
func NewWeChatHandler(db *gorm.DB) (*WeChatHandler, error) {
	wechatConfig := config.LoadWeChatConfig()
	if err := wechatConfig.Validate(); err != nil {
		return nil, fmt.Errorf("invalid wechat config: %w", err)
	}

	apiClient := wechat.NewClient(wechatConfig.AppID, wechatConfig.AppSecret)

	oauthService, err := wechatservice.NewService(&wechatservice.Config{
		AppID:       wechatConfig.AppID,
		AppSecret:   wechatConfig.AppSecret,
		RedirectURI: wechatConfig.RedirectURI,
		Scope:       wechatConfig.Scope,
	}, apiClient)
	if err != nil {
		return nil, fmt.Errorf("create oauth service failed: %w", err)
	}

	stateRepo := repositories.NewOAuthStateRepository(db)
	bindingRepo := repositories.NewUserThirdpartyBindingRepository(db)
	userService := thirdpartyuser.NewService(db, bindingRepo)
	jwtService := NewJWTService()

	return &WeChatHandler{
		oauthService: oauthService,
		stateRepo:    stateRepo,
		userService:  userService,
		jwtService:   jwtService,
		config:       wechatConfig,
	}, nil
}

// Authorize 获取微信授权 URL
// GET /api/order-ease/v1/thirdparty/wechat/authorize
func (h *WeChatHandler) Authorize(c *gin.Context) {
	state := h.stateRepo.Generate()

	if err := h.stateRepo.Save(state, oauth.ProviderWeChat, 600); err != nil {
		log2.Errorf("save oauth state failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "生成授权状态失败",
		})
		return
	}

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

	if err := h.stateRepo.VerifyAndDelete(state); err != nil {
		log2.Errorf("verify state failed: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的授权状态",
		})
		return
	}

	result, err := h.oauthService.HandleCallback(c.Request.Context(), code, state)
	if err != nil {
		log2.Errorf("handle callback failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "获取授权信息失败",
		})
		return
	}

	log2.Debugf("OAuth result: openID=%s, unionID=%s", result.OpenID, result.UnionID)

	userModel, err := h.userService.FindOrCreateByOpenID(result)
	if err != nil {
		log2.Errorf("find or create user failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "用户处理失败",
		})
		return
	}

	log2.Debugf("User: ID=%d, Name=%s", userModel.ID, userModel.Name)

	token, _, err := h.jwtService.GenerateToken(userModel)
	if err != nil {
		log2.Errorf("generate token failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "生成令牌失败",
		})
		return
	}

	frontendURL := fmt.Sprintf("/order-ease-iui/wechat-login-callback?token=%s", token)

	c.Redirect(http.StatusFound, frontendURL)
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
		Nickname string `json:"nickname"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		log2.Errorf("参数绑定失败: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的参数",
		})
		return
	}

	username := fmt.Sprintf("wx:%s", req.OpenID)

	log2.Debugf("Login by OpenID: openID=%s, username=%s, nickname=%s", req.OpenID, username, req.Nickname)

	userModel, err := h.userService.FindOrCreateByName(username, req.Nickname)
	if err != nil {
		log2.Errorf("find or create user failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "用户处理失败",
		})
		return
	}

	token, expiresIn, err := h.jwtService.GenerateToken(userModel)
	if err != nil {
		log2.Errorf("generate token failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "生成令牌失败",
		})
		return
	}

	log2.Infof("User logged in by OpenID: ID=%d, username=%s, nickname=%s", userModel.ID, username, req.Nickname)

	c.JSON(http.StatusOK, gin.H{
		"token":       token,
		"expires_in":  expiresIn,
		"first_login": userModel.Nickname == "",
		"user": gin.H{
			"id":       userModel.ID,
			"name":     userModel.Name,
			"nickname": userModel.Nickname,
			"role":     userModel.Role,
			"type":     userModel.Type,
		},
	})
}

// Unbind 解绑微信账号（需要用户登录）
// DELETE /api/order-ease/v1/thirdparty/wechat/unbind
func (h *WeChatHandler) Unbind(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "解绑功能待实现",
	})
}
