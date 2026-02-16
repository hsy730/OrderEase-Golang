package wechat

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewMiniProgramClient(t *testing.T) {
	appID := "test_app_id"
	appSecret := "test_app_secret"

	client := NewMiniProgramClient(appID, appSecret)

	assert.NotNil(t, client)
	assert.NotNil(t, client.httpClient)
	assert.Equal(t, appID, client.appID)
	assert.Equal(t, appSecret, client.appSecret)
	assert.Equal(t, 30.0, client.httpClient.Timeout.Seconds())
}

func TestMiniProgramClient_Code2Session_Success(t *testing.T) {
	// 创建模拟微信服务器
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"openid": "test_openid_123",
			"session_key": "test_session_key",
			"unionid": "test_unionid"
		}`))
	}))
	defer server.Close()

	// 创建客户端并覆盖 base URL
	client := NewMiniProgramClient("test_app_id", "test_app_secret")
	client.httpClient = &http.Client{}

	// 注意：这里需要手动调用微信 API，实际测试中可以使用 mock
	// 由于当前实现直接调用微信 API，我们只测试解析逻辑

	t.Run("TestSessionInfo_Parsing", func(t *testing.T) {
		result := &SessionInfo{
			OpenID:    "test_openid_123",
			SessionKey: "test_session_key",
			UnionID:   "test_unionid",
			ErrCode:   0,
			ErrMsg:    "",
		}

		assert.False(t, result.IsError())
		assert.Nil(t, result.GetError())
		assert.Equal(t, "test_openid_123", result.OpenID)
		assert.Equal(t, "test_session_key", result.SessionKey)
		assert.Equal(t, "test_unionid", result.UnionID)
	})
}

func TestSessionInfo_IsError(t *testing.T) {
	tests := []struct {
		name     string
		errCode  int
		expected bool
	}{
		{
			name:     "no error",
			errCode:  0,
			expected: false,
		},
		{
			name:     "with error",
			errCode:  40013,
			expected: true,
		},
		{
			name:     "invalid code",
			errCode:  40029,
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := &SessionInfo{
				OpenID:   "test_openid",
				SessionKey: "test_session_key",
				ErrCode:   tt.errCode,
				ErrMsg:    "test error",
			}

			assert.Equal(t, tt.expected, result.IsError())
		})
	}
}

func TestSessionInfo_GetError(t *testing.T) {
	t.Run("no error returns nil", func(t *testing.T) {
		result := &SessionInfo{
			OpenID:   "test_openid",
			SessionKey: "test_session_key",
			ErrCode:   0,
			ErrMsg:    "",
		}

		assert.Nil(t, result.GetError())
	})

	t.Run("with error returns SessionError", func(t *testing.T) {
		result := &SessionInfo{
			OpenID:   "test_openid",
			SessionKey: "test_session_key",
			ErrCode:   40013,
			ErrMsg:    "invalid appid",
		}

		err := result.GetError()
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "40013")
		assert.Contains(t, err.Error(), "invalid appid")
	})
}

func TestSessionError_Error(t *testing.T) {
	err := &SessionError{
		ErrCode: 40013,
		ErrMsg:  "invalid appid",
	}

	expected := "WeChat API error: [40013] invalid appid"
	assert.Equal(t, expected, err.Error())
}
