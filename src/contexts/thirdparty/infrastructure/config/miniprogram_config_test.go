package config

import (
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestLoadMiniProgramConfig(t *testing.T) {
	// 设置测试配置
	viper.Set("thirdparty.wechat.miniprogram.enabled", true)
	viper.Set("thirdparty.wechat.miniprogram.app_id", "test_app_id")
	viper.Set("thirdparty.wechat.miniprogram.app_secret", "test_secret")

	config := LoadMiniProgramConfig()

	assert.NotNil(t, config)
	assert.True(t, config.Enabled)
	assert.Equal(t, "test_app_id", config.AppID)
	assert.Equal(t, "test_secret", config.AppSecret)
}

func TestMiniProgramConfig_Validate_Disabled(t *testing.T) {
	config := &MiniProgramConfig{
		Enabled:   false,
		AppID:     "",
		AppSecret: "",
	}

	err := config.Validate()
	assert.NoError(t, err)
}

func TestMiniProgramConfig_Validate_EnabledNoAppID(t *testing.T) {
	config := &MiniProgramConfig{
		Enabled:   true,
		AppID:     "",
		AppSecret: "test_secret",
	}

	err := config.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "app_id")
	assert.Contains(t, err.Error(), "required")
}

func TestMiniProgramConfig_Validate_EnabledNoAppSecret(t *testing.T) {
	config := &MiniProgramConfig{
		Enabled:   true,
		AppID:     "test_app_id",
		AppSecret: "",
	}

	err := config.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "app_secret")
	assert.Contains(t, err.Error(), "required")
}

func TestMiniProgramConfig_Validate_EnabledWithAllFields(t *testing.T) {
	config := &MiniProgramConfig{
		Enabled:   true,
		AppID:     "test_app_id",
		AppSecret: "test_secret",
	}

	err := config.Validate()
	assert.NoError(t, err)
}

func TestMiniProgramConfig_IsEnabled(t *testing.T) {
	t.Run("enabled with all fields", func(t *testing.T) {
		config := &MiniProgramConfig{
			Enabled:   true,
			AppID:     "test_app_id",
			AppSecret: "test_secret",
		}
		assert.True(t, config.IsEnabled())
	})

	t.Run("enabled but missing appid", func(t *testing.T) {
		config := &MiniProgramConfig{
			Enabled:   true,
			AppID:     "",
			AppSecret: "test_secret",
		}
		assert.False(t, config.IsEnabled())
	})

	t.Run("enabled but missing secret", func(t *testing.T) {
		config := &MiniProgramConfig{
			Enabled:   true,
			AppID:     "test_app_id",
			AppSecret: "",
		}
		assert.False(t, config.IsEnabled())
	})
}

func TestMiniProgramConfig_IsEnabled_FalseWhenDisabled(t *testing.T) {
	config := &MiniProgramConfig{
		Enabled:   false,
		AppID:     "test_app_id",
		AppSecret: "test_secret",
	}
	assert.False(t, config.IsEnabled())
}
