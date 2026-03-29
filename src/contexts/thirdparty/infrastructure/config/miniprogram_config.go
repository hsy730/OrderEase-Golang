package config

import (
	"os"
	"strconv"

	"github.com/spf13/viper"
)

// MiniProgramConfig 微信小程序配置
type MiniProgramConfig struct {
	Enabled   bool   // 是否启用小程序登录
	AppID     string // 小程序 AppID
	AppSecret string // 小程序 AppSecret
}

// LoadMiniProgramConfig 加载小程序配置（优先从环境变量读取）
func LoadMiniProgramConfig() *MiniProgramConfig {
	config := &MiniProgramConfig{}

	// 优先从环境变量读取
	if enabledStr := os.Getenv("WECHAT_MINIPROGRAM_ENABLED"); enabledStr != "" {
		config.Enabled = enabledStr == "true" || enabledStr == "1"
	} else {
		config.Enabled = viper.GetBool("thirdparty.wechat.miniprogram.enabled")
	}

	if appID := os.Getenv("WECHAT_MINIPROGRAM_APP_ID"); appID != "" {
		config.AppID = appID
	} else {
		config.AppID = viper.GetString("thirdparty.wechat.miniprogram.app_id")
	}

	if appSecret := os.Getenv("WECHAT_MINIPROGRAM_APP_SECRET"); appSecret != "" {
		config.AppSecret = appSecret
	} else {
		config.AppSecret = viper.GetString("thirdparty.wechat.miniprogram.app_secret")
	}

	return config
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

// GetEnvBool 辅助函数：从环境变量读取布尔值
func GetEnvBool(key string, defaultValue bool) bool {
	if val := os.Getenv(key); val != "" {
		b, err := strconv.ParseBool(val)
		if err == nil {
			return b
		}
	}
	return defaultValue
}
