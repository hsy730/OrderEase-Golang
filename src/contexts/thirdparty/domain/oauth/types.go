package oauth

// OAuthResult OAuth 授权结果
type OAuthResult struct {
	Provider     Provider // 平台提供者
	OpenID       string   // 用户在平台上的唯一标识
	UnionID      string   // 用户在开放平台的唯一标识（部分平台支持）
	AccessToken  string   // 访问令牌
	RefreshToken string   // 刷新令牌
	ExpiresIn    int64    // 过期时间（秒）
	RawData      map[string]interface{} // 原始数据
}

// UserInfo 第三方平台用户信息
type UserInfo struct {
	OpenID   string // 用户唯一标识
	UnionID  string // 开放平台唯一标识
	Nickname string // 昵称
	Avatar   string // 头像 URL
	Gender   int    // 性别：0-未知，1-男，2-女
	Country  string // 国家
	Province string // 省份
	City     string // 城市
}
