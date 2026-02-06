package wechat

// AccessTokenResponse 获取 access_token 的响应
type AccessTokenResponse struct {
	AccessToken  string `json:"access_token"`
	ExpiresIn    int64  `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	OpenID       string `json:"openid"`
	Scope        string `json:"scope"`
	UnionID      string `json:"unionid,omitempty"`
	ErrCode      int    `json:"errcode,omitempty"`
	ErrMsg       string `json:"errmsg,omitempty"`
}

// IsError 检查响应是否包含错误
func (r *AccessTokenResponse) IsError() bool {
	return r.ErrCode != 0
}

// GetError 获取错误信息
func (r *AccessTokenResponse) GetError() *WeChatError {
	if !r.IsError() {
		return nil
	}
	return &WeChatError{
		ErrCode: WeChatErrorCode(r.ErrCode),
		ErrMsg:  r.ErrMsg,
	}
}

// UserInfo 微信用户信息
type UserInfo struct {
	OpenID     string   `json:"openid"`
	Nickname   string   `json:"nickname"`
	Sex        int      `json:"sex"`        // 1-男，2-女，0-未知
	Province   string   `json:"province"`
	City       string   `json:"city"`
	Country    string   `json:"country"`
	HeadImgURL string   `json:"headimgurl"`
	Privilege  []string `json:"privilege"`
	UnionID    string   `json:"unionid,omitempty"`
}

// GetGenderText 获取性别文本
func (u *UserInfo) GetGenderText() string {
	switch u.Sex {
	case 1:
		return "男"
	case 2:
		return "女"
	default:
		return "未知"
	}
}
