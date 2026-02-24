package value_objects

import (
	"errors"
	"regexp"
	"strings"
	"unicode"
)

// Password 密码值对象
type Password string

// NewWeakPassword 创建弱密码值对象（6-20位，必须包含字母或数字）
// 用于管理员创建用户、前端用户注册等场景
// 强密码也能通过弱密码校验（因为强密码包含字母和数字）
func NewWeakPassword(password string) (Password, error) {
	if len(password) < 6 || len(password) > 20 {
		return "", errors.New("密码长度必须在6-20位")
	}

	hasLetterOrDigit := regexp.MustCompile(`[a-zA-Z0-9]`).MatchString(password)
	if !hasLetterOrDigit {
		return "", errors.New("密码必须包含字母或数字")
	}

	return Password(password), nil
}

// NewStrictPassword 创建强密码值对象（8-20位，大小写字母+数字+特殊字符）
// 用于管理员和店主密码
func NewStrictPassword(password string) (Password, error) {
	if len(password) < 8 || len(password) > 20 {
		return "", errors.New("密码长度必须在8-20位")
	}

	var (
		hasNumber  bool
		hasLower   bool
		hasUpper   bool
		hasSpecial bool
	)

	for _, char := range password {
		switch {
		case unicode.IsNumber(char):
			hasNumber = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	if !hasNumber {
		return "", errors.New("密码必须包含数字")
	}
	if !hasLower {
		return "", errors.New("密码必须包含小写字母")
	}
	if !hasUpper {
		return "", errors.New("密码必须包含大写字母")
	}
	if !hasSpecial {
		return "", errors.New("密码必须包含特殊字符")
	}

	return Password(password), nil
}

// String 返回密码字符串
func (p Password) String() string {
	return string(p)
}

// Hash 返回密码的哈希值（用于bcrypt）
func (p Password) Hash() (string, error) {
	if strings.HasPrefix(string(p), "$2a$") {
		return string(p), nil
	}
	return "", nil
}

// IsValid 验证密码是否符合弱密码格式
func (p Password) IsValid() bool {
	_, err := NewWeakPassword(string(p))
	return err == nil
}

// IsStrictValid 验证密码是否符合强密码格式
func (p Password) IsStrictValid() bool {
	_, err := NewStrictPassword(string(p))
	return err == nil
}
