package config

import (
	"github.com/spf13/viper"
)

// MiniProgramConfig 微信小程序配置
type MiniProgramConfig struct {
	Enabled   bool   // 是否启用小程序登录
	AppID     string // 小程序 AppID
	AppSecret string // 小程序 AppSecret
}

// LoadMiniProgramConfig 加载小程序配置
func LoadMiniProgramConfig() *MiniProgramConfig {
	return &MiniProgramConfig{
		Enabled:   viper.GetBool("thirdparty.wechat.miniprogram.enabled"),
		AppID:     viper.GetString("thirdparty.wechat.miniprogram.app_id"),
		AppSecret: viper.GetString("thirdparty.wechat.miniprogram.app_secret"),
	}
}

// Validate 验证配置
func (c *MiniProgramConfig) Validate() error {
	if !c.Enabled {
		return nil
	}
	if c.AppID == "" {
		return &ConfigError{Field: "miniprogram.app_id", Message: "小程序 app_id is required when enabled"}
	}
	if c.AppSecret == "" {
		return &ConfigError{Field: "miniprogram.app_secret", Message: "小程序 app_secret is required when enabled"}
	}
	return nil
}

// IsEnabled 检查是否启用
func (c *MiniProgramConfig) IsEnabled() bool {
	return c.Enabled && c.AppID != "" && c.AppSecret != ""
}
