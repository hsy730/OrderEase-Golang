package wechat

import "fmt"

// WeChatErrorCode 微信错误码
type WeChatErrorCode int

const (
	ErrCodeOK                WeChatErrorCode = 0     // 成功
	ErrCodeInvalidAppID      WeChatErrorCode = 40013 // 无效 AppID
	ErrCodeInvalidSecret     WeChatErrorCode = 40125 // 无效 AppSecret
	ErrCodeInvalidCode       WeChatErrorCode = 40029 // 无效 code
	ErrCodeCodeExpired       WeChatErrorCode = 40163 // code 已过期
	ErrCodeInvalidGrant      WeChatErrorCode = 40030 // 无效 grant_type
	ErrCodeRedirectMismatch  WeChatErrorCode = 40163 // 重定向 URI 不匹配
)

// WeChatError 微信 API 错误
type WeChatError struct {
	ErrCode WeChatErrorCode
	ErrMsg  string
}

// Error 实现 error 接口
func (e *WeChatError) Error() string {
	return fmt.Sprintf("WeChat API error: [%d] %s", e.ErrCode, e.ErrMsg)
}

// IsInvalidCode 检查是否为无效 code 错误
func (e *WeChatError) IsInvalidCode() bool {
	return e.ErrCode == ErrCodeInvalidCode || e.ErrCode == ErrCodeCodeExpired
}

// IsConfigError 检查是否为配置错误
func (e *WeChatError) IsConfigError() bool {
	return e.ErrCode == ErrCodeInvalidAppID || e.ErrCode == ErrCodeInvalidSecret
}

// IsRedirectMismatch 检查是否为重定向 URI 不匹配
func (e *WeChatError) IsRedirectMismatch() bool {
	return e.ErrCode == ErrCodeRedirectMismatch
}
