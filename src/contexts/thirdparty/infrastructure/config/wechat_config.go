package config

import (
	"github.com/spf13/viper"
)

// WeChatConfig 微信配置
type WeChatConfig struct {
	Enabled     bool   // 是否启用微信登录
	AppID       string // 微信公众号 AppID
	AppSecret   string // 微信公众号 AppSecret
	RedirectURI string // 授权回调地址
	Scope       string // 授权作用域：snsapi_base | snsapi_userinfo
}

// LoadWeChatConfig 从配置文件加载微信配置
func LoadWeChatConfig() *WeChatConfig {
	return &WeChatConfig{
		Enabled:     viper.GetBool("thirdparty.wechat.enabled"),
		AppID:       viper.GetString("thirdparty.wechat.app_id"),
		AppSecret:   viper.GetString("thirdparty.wechat.app_secret"),
		RedirectURI: viper.GetString("thirdparty.wechat.redirect_uri"),
		Scope:       viper.GetString("thirdparty.wechat.scope"),
	}
}

// IsEnabled 检查是否启用微信登录
func (c *WeChatConfig) IsEnabled() bool {
	return c.Enabled && c.AppID != "" && c.AppSecret != ""
}

// Validate 验证配置
func (c *WeChatConfig) Validate() error {
	if !c.Enabled {
		return nil
	}
	if c.AppID == "" {
		return &ConfigError{Field: "app_id", Message: "wechat app_id is required when enabled"}
	}
	if c.AppSecret == "" {
		return &ConfigError{Field: "app_secret", Message: "wechat app_secret is required when enabled"}
	}
	if c.RedirectURI == "" {
		return &ConfigError{Field: "redirect_uri", Message: "wechat redirect_uri is required when enabled"}
	}
	if c.Scope == "" {
		c.Scope = "snsapi_base" // 默认静默授权
	}
	if c.Scope != "snsapi_base" && c.Scope != "snsapi_userinfo" {
		return &ConfigError{Field: "scope", Message: "wechat scope must be snsapi_base or snsapi_userinfo"}
	}
	return nil
}

// ConfigError 配置错误
type ConfigError struct {
	Field   string
	Message string
}

func (e *ConfigError) Error() string {
	return e.Message
}
