package value_objects

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewWeakPassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantErr  bool
		errMsg   string
	}{
		{
			name:     "valid - letters and digits",
			password: "abc123",
			wantErr:  false,
		},
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
			name:     "valid - 20 characters",
			password: "abcdefghijklmn12345",
			wantErr:  false,
		},
		{
			name:     "valid - uppercase letters",
			password: "ABCDEF",
			wantErr:  false,
		},
		{
			name:     "valid - with special chars (like strict password)",
			password: "Abc123!@",
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
			name:     "only special chars - no letter or digit",
			password: "!@#$%^",
			wantErr:  true,
			errMsg:   "密码必须包含字母或数字",
		},
		{
			name:     "empty string",
			password: "",
			wantErr:  true,
			errMsg:   "密码长度必须在6-20位",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewWeakPassword(tt.password)
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
			name:     "valid strict password - 20 chars",
			password: "MyStr0ng!Pass12345!",
			wantErr:  false,
		},
		{
			name:     "too short - less than 8 chars",
			password: "Abc123!",
			wantErr:  true,
			errMsg:   "密码长度必须在8-20位",
		},
		{
			name:     "too long - more than 20 chars",
			password: "MyStr0ng!Pass12345!TooLong",
			wantErr:  true,
			errMsg:   "密码长度必须在8-20位",
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

func TestStrictPasswordPassesWeakValidation(t *testing.T) {
	strictPasswords := []string{
		"Abc123!@",
		"MyStr0ng!Pass",
		"Test123!@#",
	}

	for _, pwd := range strictPasswords {
		t.Run("strict password: "+pwd, func(t *testing.T) {
			_, err := NewWeakPassword(pwd)
			assert.NoError(t, err, "强密码应该能通过弱密码校验")
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
			name:      "valid weak password - letters and digits",
			password:  Password("abc123"),
			wantValid: true,
		},
		{
			name:      "valid weak password - all letters",
			password:  Password("abcdef"),
			wantValid: true,
		},
		{
			name:      "valid weak password - all digits",
			password:  Password("123456"),
			wantValid: true,
		},
		{
			name:      "valid - strict password passes weak validation",
			password:  Password("Abc123!@"),
			wantValid: true,
		},
		{
			name:      "invalid - too short",
			password:  Password("ab12"),
			wantValid: false,
		},
		{
			name:      "invalid - only special chars",
			password:  Password("!@#$%^"),
			wantValid: false,
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
			name:      "invalid - weak password not strict",
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
