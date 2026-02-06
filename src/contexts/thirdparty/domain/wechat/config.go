package wechat

import (
	"fmt"
	"strings"
)

// Config 微信 OAuth 配置
type Config struct {
	AppID       string // 微信公众号 AppID
	AppSecret   string // 微信公众号 AppSecret
	RedirectURI string // 授权回调地址
	Scope       string // 授权作用域：snsapi_base | snsapi_userinfo
}

// Validate 验证配置是否有效
func (c *Config) Validate() error {
	if c.AppID == "" {
		return fmt.Errorf("wechat AppID is required")
	}
	if c.AppSecret == "" {
		return fmt.Errorf("wechat AppSecret is required")
	}
	if c.RedirectURI == "" {
		return fmt.Errorf("wechat RedirectURI is required")
	}
	if c.Scope == "" {
		c.Scope = "snsapi_base" // 默认使用静默授权
	}
	// 验证 scope 是否有效
	if c.Scope != "snsapi_base" && c.Scope != "snsapi_userinfo" {
		return fmt.Errorf("invalid wechat scope: %s", c.Scope)
	}
	return nil
}

// IsUserInfoScope 是否为获取用户信息作用域
func (c *Config) IsUserInfoScope() bool {
	return c.Scope == "snsapi_userinfo"
}

// GetAuthDomain 获取授权域名
// 根据配置的 redirect_uri 判断使用哪个授权域名
func (c *Config) GetAuthDomain() string {
	// 开放平台授权域名
	openDomain := "open.weixin.qq.com"

	// 如果 redirect_uri 包含特定域名，可以切换到其他授权域名
	if strings.Contains(c.RedirectURI, "sandbox") {
		return "open.weixin.qq.com" // 沙箱环境
	}

	return openDomain
}
