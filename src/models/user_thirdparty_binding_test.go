package models

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/bwmarrin/snowflake"
	"github.com/stretchr/testify/assert"
)

func TestUserThirdpartyBindingStruct(t *testing.T) {
	now := time.Now()
	userID := snowflake.ID(123)

	binding := UserThirdpartyBinding{
		ID:             1,
		UserID:         userID,
		Provider:       "wechat",
		ProviderUserID: "wx123456",
		UnionID:        "union123",
		Nickname:       "测试用户",
		AvatarURL:      "https://example.com/avatar.jpg",
		Gender:         1,
		Country:        "中国",
		Province:       "北京",
		City:           "北京市",
		Metadata:       Metadata{"access_token": "token123"},
		IsActive:       true,
		LastLoginAt:    &now,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	assert.Equal(t, uint(1), binding.ID)
	assert.Equal(t, userID, binding.UserID)
	assert.Equal(t, "wechat", binding.Provider)
	assert.Equal(t, "wx123456", binding.ProviderUserID)
	assert.Equal(t, "union123", binding.UnionID)
	assert.Equal(t, "测试用户", binding.Nickname)
	assert.Equal(t, "https://example.com/avatar.jpg", binding.AvatarURL)
	assert.Equal(t, 1, binding.Gender)
	assert.Equal(t, "中国", binding.Country)
	assert.Equal(t, "北京", binding.Province)
	assert.Equal(t, "北京市", binding.City)
	assert.True(t, binding.IsActive)
	assert.Equal(t, now, *binding.LastLoginAt)
}

func TestUserThirdpartyBindingStructEmpty(t *testing.T) {
	binding := UserThirdpartyBinding{}

	assert.Zero(t, binding.ID)
	assert.Zero(t, binding.UserID)
	assert.Empty(t, binding.Provider)
	assert.Empty(t, binding.ProviderUserID)
	assert.Empty(t, binding.UnionID)
	assert.Empty(t, binding.Nickname)
	assert.Empty(t, binding.AvatarURL)
	assert.Zero(t, binding.Gender)
	assert.Empty(t, binding.Country)
	assert.Empty(t, binding.Province)
	assert.Empty(t, binding.City)
	assert.False(t, binding.IsActive)
	assert.Nil(t, binding.LastLoginAt)
}

func TestUserThirdpartyBinding_TableName(t *testing.T) {
	binding := UserThirdpartyBinding{}
	assert.Equal(t, "user_thirdparty_bindings", binding.TableName())
}

func TestUserThirdpartyBinding_IsActiveBinding(t *testing.T) {
	t.Run("active binding", func(t *testing.T) {
		binding := UserThirdpartyBinding{IsActive: true}
		assert.True(t, binding.IsActiveBinding())
	})

	t.Run("inactive binding", func(t *testing.T) {
		binding := UserThirdpartyBinding{IsActive: false}
		assert.False(t, binding.IsActiveBinding())
	})
}

func TestUserThirdpartyBinding_UpdateLastLogin(t *testing.T) {
	binding := UserThirdpartyBinding{}

	assert.Nil(t, binding.LastLoginAt)

	binding.UpdateLastLogin()

	assert.NotNil(t, binding.LastLoginAt)
	assert.False(t, binding.LastLoginAt.IsZero())
}

func TestUserThirdpartyBinding_UpdateLastLogin_MultipleTimes(t *testing.T) {
	binding := UserThirdpartyBinding{}

	binding.UpdateLastLogin()
	firstLogin := *binding.LastLoginAt

	time.Sleep(10 * time.Millisecond)

	binding.UpdateLastLogin()
	secondLogin := *binding.LastLoginAt

	assert.True(t, secondLogin.After(firstLogin) || secondLogin.Equal(firstLogin))
}

func TestMetadata_Scan(t *testing.T) {
	t.Run("scan nil value", func(t *testing.T) {
		var m Metadata
		err := m.Scan(nil)

		assert.NoError(t, err)
		assert.NotNil(t, m)
	})

	t.Run("scan valid JSON", func(t *testing.T) {
		jsonData := `{"access_token":"token123","refresh_token":"refresh123"}`
		var m Metadata
		err := m.Scan([]byte(jsonData))

		assert.NoError(t, err)
		assert.Equal(t, "token123", m["access_token"])
		assert.Equal(t, "refresh123", m["refresh_token"])
	})

	t.Run("scan invalid type", func(t *testing.T) {
		var m Metadata
		err := m.Scan("not a byte slice")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to unmarshal")
	})

	t.Run("scan empty JSON object", func(t *testing.T) {
		var m Metadata
		err := m.Scan([]byte("{}"))

		assert.NoError(t, err)
		assert.NotNil(t, m)
	})
}

func TestMetadata_Value(t *testing.T) {
	t.Run("empty metadata returns nil", func(t *testing.T) {
		m := Metadata{}
		value, err := m.Value()

		assert.NoError(t, err)
		assert.Nil(t, value)
	})

	t.Run("nil metadata returns nil", func(t *testing.T) {
		var m Metadata
		value, err := m.Value()

		assert.NoError(t, err)
		assert.Nil(t, value)
	})

	t.Run("metadata with values returns JSON", func(t *testing.T) {
		m := Metadata{
			"access_token":  "token123",
			"refresh_token": "refresh123",
		}
		value, err := m.Value()

		assert.NoError(t, err)
		assert.NotNil(t, value)

		var decoded map[string]interface{}
		err = json.Unmarshal(value.([]byte), &decoded)
		assert.NoError(t, err)
		assert.Equal(t, "token123", decoded["access_token"])
		assert.Equal(t, "refresh123", decoded["refresh_token"])
	})
}

func TestMetadata_GetAccessToken(t *testing.T) {
	t.Run("token exists", func(t *testing.T) {
		m := Metadata{"access_token": "token123"}
		assert.Equal(t, "token123", m.GetAccessToken())
	})

	t.Run("token not exists", func(t *testing.T) {
		m := Metadata{}
		assert.Equal(t, "", m.GetAccessToken())
	})

	t.Run("token is not string", func(t *testing.T) {
		m := Metadata{"access_token": 12345}
		assert.Equal(t, "", m.GetAccessToken())
	})
}

func TestMetadata_GetRefreshToken(t *testing.T) {
	t.Run("token exists", func(t *testing.T) {
		m := Metadata{"refresh_token": "refresh123"}
		assert.Equal(t, "refresh123", m.GetRefreshToken())
	})

	t.Run("token not exists", func(t *testing.T) {
		m := Metadata{}
		assert.Equal(t, "", m.GetRefreshToken())
	})

	t.Run("token is not string", func(t *testing.T) {
		m := Metadata{"refresh_token": 12345}
		assert.Equal(t, "", m.GetRefreshToken())
	})
}

func TestMetadata_SetAccessToken(t *testing.T) {
	t.Run("set on existing metadata", func(t *testing.T) {
		m := Metadata{}
		m.SetAccessToken("newtoken")

		assert.Equal(t, "newtoken", m["access_token"])
	})

	t.Run("overwrite existing token", func(t *testing.T) {
		m := Metadata{"access_token": "oldtoken"}
		m.SetAccessToken("newtoken")

		assert.Equal(t, "newtoken", m["access_token"])
	})
}

func TestMetadata_SetRefreshToken(t *testing.T) {
	t.Run("set on existing metadata", func(t *testing.T) {
		m := Metadata{}
		m.SetRefreshToken("newrefresh")

		assert.Equal(t, "newrefresh", m["refresh_token"])
	})

	t.Run("overwrite existing token", func(t *testing.T) {
		m := Metadata{"refresh_token": "oldrefresh"}
		m.SetRefreshToken("newrefresh")

		assert.Equal(t, "newrefresh", m["refresh_token"])
	})
}

func TestMetadata_ScanValueRoundtrip(t *testing.T) {
	original := Metadata{
		"access_token":  "token123",
		"refresh_token": "refresh123",
		"expires_in":    7200,
	}

	value, err := original.Value()
	assert.NoError(t, err)

	var scanned Metadata
	err = scanned.Scan(value)
	assert.NoError(t, err)

	assert.Equal(t, original["access_token"], scanned["access_token"])
	assert.Equal(t, original["refresh_token"], scanned["refresh_token"])
}

func TestMetadata_ComplexData(t *testing.T) {
	m := Metadata{
		"access_token": "token123",
		"user_info": map[string]interface{}{
			"nickname": "测试用户",
			"avatar":   "https://example.com/avatar.jpg",
		},
		"scopes": []string{"user:read", "user:write"},
	}

	value, err := m.Value()
	assert.NoError(t, err)

	var scanned Metadata
	err = scanned.Scan(value)
	assert.NoError(t, err)

	assert.Equal(t, "token123", scanned.GetAccessToken())
}
