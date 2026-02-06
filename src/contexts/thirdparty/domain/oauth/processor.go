package oauth

import "context"

// Processor OAuth 处理器接口
// 所有第三方平台（微信、支付宝等）都需要实现此接口
type Processor interface {
	// GetAuthorizeURL 获取授权 URL
	// state: 用于防止 CSRF 攻击的随机字符串
	// redirectURI: 授权完成后的回调地址
	GetAuthorizeURL(state string, redirectURI string) string

	// HandleCallback 处理授权回调
	// ctx: 上下文
	// code: 授权码
	// state: 防止 CSRF 的状态参数
	HandleCallback(ctx context.Context, code string, state string) (*OAuthResult, error)

	// GetUserInfo 获取用户详细信息
	// ctx: 上下文
	// accessToken: 访问令牌
	// openID: 用户唯一标识
	GetUserInfo(ctx context.Context, accessToken string, openID string) (*UserInfo, error)

	// GetProvider 获取平台提供者类型
	GetProvider() Provider
}
