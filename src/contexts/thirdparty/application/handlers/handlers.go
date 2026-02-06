package handlers

import (
	"orderease/contexts/thirdparty/infrastructure/config"

	"gorm.io/gorm"
)

// Handler 第三方平台处理器集合
type Handler struct {
	WeChat *WeChatHandler
	// 未来添加:
	// Alipay *AlipayHandler
}

// NewHandler 创建第三方平台处理器
func NewHandler(db *gorm.DB) (*Handler, error) {
	handler := &Handler{}

	// 初始化微信处理器
	wechatConfig := config.LoadWeChatConfig()
	if wechatConfig.IsEnabled() {
		wechatHandler, err := NewWeChatHandler(db)
		if err != nil {
			return nil, err
		}
		handler.WeChat = wechatHandler
	}

	return handler, nil
}
