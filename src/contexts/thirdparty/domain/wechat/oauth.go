package wechat

import (
	"context"
	"fmt"
	"net/url"
	"orderease/contexts/thirdparty/domain/oauth"
)

// Service 微信 OAuth 服务
type Service struct {
	config     *Config
	httpClient HTTPClient
}

// HTTPClient HTTP 客户端接口
type HTTPClient interface {
	GetAccessToken(ctx context.Context, code string) (*AccessTokenResponse, error)
	GetUserInfo(ctx context.Context, accessToken, openID string) (*UserInfo, error)
}

// NewService 创建微信 OAuth 服务
func NewService(config *Config, httpClient HTTPClient) (*Service, error) {
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid wechat config: %w", err)
	}

	return &Service{
		config:     config,
		httpClient: httpClient,
	}, nil
}

// GetAuthorizeURL 获取微信授权 URL
func (s *Service) GetAuthorizeURL(state string, redirectURI string) string {
	authURL := fmt.Sprintf(
		"https://%s/connect/oauth2/authorize?appid=%s&redirect_uri=%s&response_type=code&scope=%s&state=%s#wechat_redirect",
		s.config.GetAuthDomain(),
		s.config.AppID,
		url.QueryEscape(redirectURI),
		s.config.Scope,
		state,
	)
	return authURL
}

// HandleCallback 处理微信授权回调
func (s *Service) HandleCallback(ctx context.Context, code string, state string) (*oauth.OAuthResult, error) {
	// 1. 使用 code 换取 access_token
	tokenResp, err := s.httpClient.GetAccessToken(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("get access token failed: %w", err)
	}

	// 2. 构建授权结果
	result := &oauth.OAuthResult{
		Provider:     oauth.ProviderWeChat,
		OpenID:       tokenResp.OpenID,
		UnionID:      tokenResp.UnionID,
		AccessToken:  tokenResp.AccessToken,
		RefreshToken: tokenResp.RefreshToken,
		ExpiresIn:    tokenResp.ExpiresIn,
		RawData:      make(map[string]interface{}),
	}

	// 将原始数据存储到 RawData
	result.RawData["scope"] = tokenResp.Scope

	// 3. 如果 scope 是 snsapi_userinfo，获取用户详细信息
	if s.config.IsUserInfoScope() && tokenResp.AccessToken != "" {
		userInfo, err := s.httpClient.GetUserInfo(ctx, tokenResp.AccessToken, tokenResp.OpenID)
		if err == nil {
			result.RawData["nickname"] = userInfo.Nickname
			result.RawData["headimgurl"] = userInfo.HeadImgURL
			result.RawData["sex"] = userInfo.Sex
			result.RawData["province"] = userInfo.Province
			result.RawData["city"] = userInfo.City
			result.RawData["country"] = userInfo.Country
		}
		// 获取用户信息失败不影响授权流程
	}

	return result, nil
}

// GetUserInfo 获取用户信息
func (s *Service) GetUserInfo(ctx context.Context, accessToken string, openID string) (*oauth.UserInfo, error) {
	wechatUserInfo, err := s.httpClient.GetUserInfo(ctx, accessToken, openID)
	if err != nil {
		return nil, fmt.Errorf("get user info failed: %w", err)
	}

	return &oauth.UserInfo{
		OpenID:   wechatUserInfo.OpenID,
		UnionID:  wechatUserInfo.UnionID,
		Nickname: wechatUserInfo.Nickname,
		Avatar:   wechatUserInfo.HeadImgURL,
		Gender:   wechatUserInfo.Sex,
		Country:  wechatUserInfo.Country,
		Province: wechatUserInfo.Province,
		City:     wechatUserInfo.City,
	}, nil
}

// GetProvider 获取平台提供者
func (s *Service) GetProvider() oauth.Provider {
	return oauth.ProviderWeChat
}
