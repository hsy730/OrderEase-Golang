package value_objects

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewPassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantErr  bool
		errMsg   string
	}{
		{
			name:     "valid password - letters and digits",
			password: "abc123",
			wantErr:  false,
		},
		{
			name:     "valid password - 20 characters",
			password: "abcdefghijklmn12345",
			wantErr:  false,
		},
		{
			name:     "valid password - with special chars",
			password: "abc123!@#",
			wantErr:  false,
		},
		{
			name:     "too short - less than 6 chars",
			password: "ab12",
			wantErr:  true,
			errMsg:   "密码长度必须在6-20位",
		},
		{
			name:     "too long - more than 20 chars",
			password: "abcdefghijklmnopqrstuvwxyz123456",
			wantErr:  true,
			errMsg:   "密码长度必须在6-20位",
		},
		{
			name:     "only letters",
			password: "abcdef",
			wantErr:  true,
			errMsg:   "密码必须包含字母和数字",
		},
		{
			name:     "only digits",
			password: "123456",
			wantErr:  true,
			errMsg:   "密码必须包含字母和数字",
		},
		{
			name:     "empty string",
			password: "",
			wantErr:  true,
			errMsg:   "密码长度必须在6-20位",
		},
		{
			name:     "chinese characters",
			password: "中文密码123",
			wantErr:  true,
			errMsg:   "密码必须包含字母和数字",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewPassword(tt.password)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				assert.Equal(t, Password(""), got)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, Password(tt.password), got)
			}
		})
	}
}

func TestNewStrictPassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantErr  bool
		errMsg   string
	}{
		{
			name:     "valid strict password",
			password: "Abc123!@",
			wantErr:  false,
		},
		{
			name:     "valid strict password - longer",
			password: "MyStr0ng!Pass",
			wantErr:  false,
		},
		{
			name:     "too short - less than 8 chars",
			password: "Abc123!",
			wantErr:  true,
			errMsg:   "密码长度至少为8位",
		},
		{
			name:     "missing digit",
			password: "Abcdefg!",
			wantErr:  true,
			errMsg:   "密码必须包含数字",
		},
		{
			name:     "missing lowercase",
			password: "ABC123!@",
			wantErr:  true,
			errMsg:   "密码必须包含小写字母",
		},
		{
			name:     "missing uppercase",
			password: "abc123!@",
			wantErr:  true,
			errMsg:   "密码必须包含大写字母",
		},
		{
			name:     "missing special char",
			password: "Abc12345",
			wantErr:  true,
			errMsg:   "密码必须包含特殊字符",
		},
		{
			name:     "only digits and lowercase",
			password: "abc12345",
			wantErr:  true,
			errMsg:   "密码必须包含大写字母",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewStrictPassword(tt.password)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				assert.Equal(t, Password(""), got)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, Password(tt.password), got)
			}
		})
	}
}

func TestNewSimplePassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantErr  bool
		errMsg   string
	}{
		{
			name:     "valid - all letters",
			password: "abcdef",
			wantErr:  false,
		},
		{
			name:     "valid - all digits",
			password: "123456",
			wantErr:  false,
		},
		{
			name:     "valid - mixed letters and digits",
			password: "abc123",
			wantErr:  false,
		},
		{
			name:     "valid - uppercase letters",
			password: "ABCDEF",
			wantErr:  false,
		},
		{
			name:     "too short - less than 6 chars",
			password: "abc12",
			wantErr:  true,
			errMsg:   "密码必须为6位",
		},
		{
			name:     "too long - more than 6 chars",
			password: "abcdef123",
			wantErr:  true,
			errMsg:   "密码必须为6位",
		},
		{
			name:     "contains special characters",
			password: "abc12!",
			wantErr:  true,
			errMsg:   "密码必须为6位字母或数字",
		},
		{
			name:     "contains chinese characters",
			password: "中文密码",
			wantErr:  true,
			errMsg:   "密码必须为6位",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewSimplePassword(tt.password)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				assert.Equal(t, Password(""), got)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, Password(tt.password), got)
			}
		})
	}
}

func TestPassword_String(t *testing.T) {
	password := Password("test123")
	assert.Equal(t, "test123", password.String())
}

func TestPassword_Hash(t *testing.T) {
	tests := []struct {
		name       string
		password   Password
		wantHashed bool
	}{
		{
			name:       "already hashed bcrypt password",
			password:   Password("$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy"),
			wantHashed: true,
		},
		{
			name:       "plain password",
			password:   Password("test123"),
			wantHashed: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash, err := tt.password.Hash()
			if tt.wantHashed {
				assert.NoError(t, err)
				assert.Equal(t, string(tt.password), hash)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, "", hash)
			}
		})
	}
}

func TestPassword_IsValid(t *testing.T) {
	tests := []struct {
		name      string
		password  Password
		wantValid bool
	}{
		{
			name:      "valid password",
			password:  Password("abc123"),
			wantValid: true,
		},
		{
			name:      "invalid - too short",
			password:  Password("ab12"),
			wantValid: false,
		},
		{
			name:      "invalid - only letters",
			password:  Password("abcdef"),
			wantValid: false,
		},
		{
			name:      "valid with special chars",
			password:  Password("abc123!@#"),
			wantValid: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.password.IsValid()
			assert.Equal(t, tt.wantValid, got)
		})
	}
}

func TestPassword_IsStrictValid(t *testing.T) {
	tests := []struct {
		name      string
		password  Password
		wantValid bool
	}{
		{
			name:      "valid strict password",
			password:  Password("Abc123!@"),
			wantValid: true,
		},
		{
			name:      "invalid - missing uppercase",
			password:  Password("abc123!@"),
			wantValid: false,
		},
		{
			name:      "invalid - missing special char",
			password:  Password("Abc12345"),
			wantValid: false,
		},
		{
			name:      "valid simple password but not strict",
			password:  Password("abc123"),
			wantValid: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.password.IsStrictValid()
			assert.Equal(t, tt.wantValid, got)
		})
	}
}
