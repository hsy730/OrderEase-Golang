package handlers

import (
	"fmt"
	"net/http"
	"orderease/contexts/ordercontext/application/services"
	"orderease/contexts/thirdparty/domain/oauth"
	"orderease/contexts/thirdparty/domain/user"
	"orderease/contexts/thirdparty/infrastructure/config"
	"orderease/contexts/thirdparty/infrastructure/external/wechat"
	"orderease/models"
	"orderease/utils"
	"orderease/utils/log2"
	"time"

	"github.com/gin-gonic/gin"
)

// MiniProgramLoginRequest 小程序登录请求（新版微信授权流程）
type MiniProgramLoginRequest struct {
	Code      string `json:"code" binding:"required"`
	Silent    bool   `json:"silent"` // 静默登录标识
	Nickname  string `json:"nickname"`
	AvatarURL string `json:"avatar_url"`
}

// MiniProgramAuthHandler 小程序认证处理器
type MiniProgramAuthHandler struct {
	miniProgramClient *wechat.MiniProgramClient
	config            *config.MiniProgramConfig
	authService       *services.MiniProgramAuthService
	bindingRepo       user.UserBindingRepository
}

// NewMiniProgramAuthHandler 创建小程序认证处理器
func NewMiniProgramAuthHandler(authService *services.MiniProgramAuthService, bindingRepo user.UserBindingRepository) (*MiniProgramAuthHandler, error) {
	miniConfig := config.LoadMiniProgramConfig()
	if err := miniConfig.Validate(); err != nil {
		return nil, fmt.Errorf("invalid miniprogram config: %w", err)
	}

	miniClient := wechat.NewMiniProgramClient(miniConfig.AppID, miniConfig.AppSecret)

	return &MiniProgramAuthHandler{
		miniProgramClient: miniClient,
		config:            miniConfig,
		authService:       authService,
		bindingRepo:       bindingRepo,
	}, nil
}

// WeChatMiniProgramLogin 微信小程序登录（新版授权流程）
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

	log2.Debugf("微信小程序登录请求: code=%s, silent=%v, nickname=%s, avatar_url=%s", req.Code, req.Silent, req.Nickname, req.AvatarURL)

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

	var nickname, avatarURL string
	if !req.Silent {
		nickname = req.Nickname
		avatarURL = req.AvatarURL
	}

	result, err := h.authService.FindOrCreateUser(sessionInfo, nickname, avatarURL)
	if err != nil {
		log2.Errorf("查找或创建用户失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "登录失败",
			"error":   "用户处理失败",
		})
		return
	}

	token, expiredAt, err := utils.GenerateToken(uint64(result.User.ID), result.User.Name)
	if err != nil {
		log2.Errorf("生成 token 失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "登录失败",
			"error":   "生成令牌失败",
		})
		return
	}

	log2.Infof("微信小程序登录成功: ID=%d, OpenID=%s, isNewUser=%v", result.User.ID, sessionInfo.OpenID, result.IsNewUser)

	responseNickname := result.User.Nickname
	responseAvatarURL := h.extractAvatarURL(result.User, sessionInfo.OpenID)

	if responseNickname == "" || responseAvatarURL == "" {
		binding, err := h.bindingRepo.FindByUserIDAndProvider(uint64(result.User.ID), oauth.ProviderWeChat)
		if err == nil && binding != nil {
			if responseNickname == "" && binding.Nickname != "" {
				responseNickname = binding.Nickname
			}
			if responseAvatarURL == "" && binding.AvatarURL != "" {
				responseAvatarURL = binding.AvatarURL
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "登录成功",
		"data": gin.H{
			"user": gin.H{
				"id":         result.User.ID,
				"name":       result.User.Name,
				"nickname":   responseNickname,
				"avatar":     responseAvatarURL,
				"role":       result.User.Role,
				"type":       result.User.Type,
				"created_at": result.User.CreatedAt.Format(time.RFC3339),
			},
			"token":       token,
			"expiredAt":   expiredAt.Unix(),
			"first_login": result.IsNewUser,
		},
	})
}

func (h *MiniProgramAuthHandler) extractAvatarURL(userModel *models.User, openID string) string {
	binding, err := h.bindingRepo.FindByUserIDAndProvider(uint64(userModel.ID), oauth.ProviderWeChat)
	if err == nil && binding != nil && binding.AvatarURL != "" {
		return binding.AvatarURL
	}
	return ""
}
