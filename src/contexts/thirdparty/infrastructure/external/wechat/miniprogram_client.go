package wechat

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// MiniProgramClient 微信小程序 API 客户端
type MiniProgramClient struct {
	httpClient *http.Client
	appID      string
	appSecret  string
}

// NewMiniProgramClient 创建小程序客户端
func NewMiniProgramClient(appID, appSecret string) *MiniProgramClient {
	return &MiniProgramClient{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		appID:     appID,
		appSecret: appSecret,
	}
}

// Code2Session 通过 code 换取 openid 和 session_key
// 文档: https://developers.weixin.qq.com/miniprogram/dev/OpenAPIDoc/user-login/code2Session.html
func (c *MiniProgramClient) Code2Session(ctx context.Context, code string) (*SessionInfo, error) {
	url := fmt.Sprintf(
		"https://api.weixin.qq.com/sns/jscode2session?appid=%s&secret=%s&js_code=%s&grant_type=authorization_code",
		c.appID,
		c.appSecret,
		code,
	)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request failed: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("send request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body failed: %w", err)
	}

	var result SessionInfo
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("unmarshal response failed: %w", err)
	}

	if result.IsError() {
		return nil, result.GetError()
	}

	return &result, nil
}

// SessionInfo 小程序 session 信息
type SessionInfo struct {
	OpenID     string `json:"openid"`      // 用户唯一标识
	SessionKey  string `json:"session_key"` // 会话密钥
	UnionID     string `json:"unionid,omitempty"` // 在开放平台下的唯一标识符
	ErrCode     int    `json:"errcode,omitempty"`
	ErrMsg      string `json:"errmsg,omitempty"`
}

// IsError 检查是否为错误响应
func (s *SessionInfo) IsError() bool {
	return s.ErrCode != 0
}

// GetError 获取错误信息
func (s *SessionInfo) GetError() *SessionError {
	if !s.IsError() {
		return nil
	}
	return &SessionError{
		ErrCode: s.ErrCode,
		ErrMsg:  s.ErrMsg,
	}
}

// SessionError 微信小程序错误
type SessionError struct {
	ErrCode int
	ErrMsg  string
}

// Error 实现 error 接口
func (e *SessionError) Error() string {
	return fmt.Sprintf("WeChat API error: [%d] %s", e.ErrCode, e.ErrMsg)
}

// DecryptUserInfo 解密用户信息（可选，用于验证加密数据）
// 注意：小程序端已通过 getUserProfile 获取用户信息，后端可以不做解密
func (c *MiniProgramClient) DecryptUserInfo(encryptedData, iv, sessionKey string) (map[string]interface{}, error) {
	// 如果需要验证，可以实现此方法
	// 使用 AES-128-CBC 解密
	return nil, nil
}
