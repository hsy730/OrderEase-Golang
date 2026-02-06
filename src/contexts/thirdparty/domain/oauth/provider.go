package oauth

// Provider 第三方平台提供者类型
type Provider string

const (
	ProviderWeChat  Provider = "wechat"
	ProviderAlipay  Provider = "alipay"
	ProviderDingTalk Provider = "dingtalk"
)

// IsValid 检查平台提供者是否有效
func (p Provider) IsValid() bool {
	switch p {
	case ProviderWeChat, ProviderAlipay, ProviderDingTalk:
		return true
	}
	return false
}

// String 返回平台提供者的字符串表示
func (p Provider) String() string {
	return string(p)
}
