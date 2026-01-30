package user

import (
	"strings"
	"testing"

	"github.com/bwmarrin/snowflake"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
	"orderease/domain/shared/value_objects"
	"orderease/models"
)

// ==================== UserID Tests ====================

func TestNewUserID(t *testing.T) {
	userID := NewUserID()

	assert.NotEmpty(t, string(userID))
	// Should be a valid snowflake ID string format
	assert.NotEqual(t, UserID(""), userID)
}

func TestNewUserIDFromSnowflake(t *testing.T) {
	id := snowflake.ID(123456789)
	userID := NewUserIDFromSnowflake(id)

	assert.Equal(t, UserID("123456789"), userID)
}

// ==================== Constructor Tests ====================

func TestNewUser(t *testing.T) {
	tests := []struct {
		name      string
		username  string
		phone     string
		password  string
		userType  UserType
		role      UserRole
		wantErr   bool
		errMsg    string
	}{
		{
			name:     "valid user",
			username: "testuser",
			phone:    "13800138000",
			password: "abc123",
			userType: UserTypeDelivery,
			role:     UserRolePublic,
			wantErr:  false,
		},
		{
			name:     "valid user with pickup type",
			username: "testuser",
			phone:    "13800138000",
			password: "abc123",
			userType: UserTypePickup,
			role:     UserRolePrivate,
			wantErr:  false,
		},
		{
			name:     "empty phone - allowed",
			username: "testuser",
			phone:    "",
			password: "abc123",
			userType: UserTypeDelivery,
			role:     UserRolePublic,
			wantErr:  false,
		},
		{
			name:     "invalid phone - error",
			username: "testuser",
			phone:    "12345",
			password: "abc123",
			userType: UserTypeDelivery,
			role:     UserRolePublic,
			wantErr:  true,
			errMsg:   "手机号必须为11位数字且以1开头",
		},
		{
			name:     "invalid password - too short",
			username: "testuser",
			phone:    "13800138000",
			password: "ab12",
			userType: UserTypeDelivery,
			role:     UserRolePublic,
			wantErr:  true,
			errMsg:   "密码长度必须在6-20位",
		},
		{
			name:     "invalid password - only letters",
			username: "testuser",
			phone:    "13800138000",
			password: "abcdef",
			userType: UserTypeDelivery,
			role:     UserRolePublic,
			wantErr:  true,
			errMsg:   "密码必须包含字母和数字",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, err := NewUser(tt.username, tt.phone, tt.password, tt.userType, tt.role)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				assert.Nil(t, user)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, user)
				assert.NotEmpty(t, user.ID())
				assert.Equal(t, tt.username, user.Name())
				assert.Equal(t, tt.phone, user.Phone())
				assert.Equal(t, tt.userType, user.UserType())
				assert.Equal(t, tt.role, user.Role())
			}
		})
	}
}

func TestNewSimpleUser(t *testing.T) {
	tests := []struct {
		name     string
		username string
		password string
		wantErr  bool
		errMsg   string
	}{
		{
			name:     "valid simple user - 6 chars",
			username: "testuser",
			password: "abc123",
			wantErr:  false,
		},
		{
			name:     "valid simple user - all digits",
			username: "testuser",
			password: "123456",
			wantErr:  false,
		},
		{
			name:     "valid simple user - all letters",
			username: "testuser",
			password: "abcdef",
			wantErr:  false,
		},
		{
			name:     "invalid - too short",
			username: "testuser",
			password: "abc12",
			wantErr:  true,
			errMsg:   "密码必须为6位",
		},
		{
			name:     "invalid - too long",
			username: "testuser",
			password: "abcdef123",
			wantErr:  true,
			errMsg:   "密码必须为6位",
		},
		{
			name:     "invalid - contains special chars",
			username: "testuser",
			password: "abc12!",
			wantErr:  true,
			errMsg:   "密码必须为6位字母或数字",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, err := NewSimpleUser(tt.username, tt.password)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				assert.Nil(t, user)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, user)
				assert.NotEmpty(t, user.ID())
				assert.Equal(t, tt.username, user.Name())
				assert.Equal(t, "", user.Phone()) // Simple users have no phone
				assert.Equal(t, UserTypeDelivery, user.UserType())
				assert.Equal(t, UserRolePublic, user.Role())
			}
		})
	}
}

// ==================== Getter Tests ====================

func TestUser_Getters(t *testing.T) {
	user, err := NewUser("testuser", "13800138000", "abc123", UserTypeDelivery, UserRolePublic)
	assert.NoError(t, err)

	assert.NotEmpty(t, user.ID())
	assert.Equal(t, "testuser", user.Name())
	assert.Equal(t, "13800138000", user.Phone())
	assert.Equal(t, "abc123", user.Password())
	assert.Equal(t, UserTypeDelivery, user.UserType())
	assert.Equal(t, UserRolePublic, user.Role())
	assert.Equal(t, "", user.Address())
}

// ==================== Setter Tests ====================

func TestUser_Setters(t *testing.T) {
	user, _ := NewUser("testuser", "13800138000", "abc123", UserTypeDelivery, UserRolePublic)

	// Test SetName
	user.SetName("newname")
	assert.Equal(t, "newname", user.Name())

	// Test SetUserType
	user.SetUserType(UserTypePickup)
	assert.Equal(t, UserTypePickup, user.UserType())

	// Test SetRole
	user.SetRole(UserRolePrivate)
	assert.Equal(t, UserRolePrivate, user.Role())

	// Test SetAddress
	user.SetAddress("test address")
	assert.Equal(t, "test address", user.Address())
}

func TestUser_SetPhone(t *testing.T) {
	user, _ := NewUser("testuser", "13800138000", "abc123", UserTypeDelivery, UserRolePublic)

	tests := []struct {
		name    string
		phone   string
		wantErr bool
	}{
		{
			name:    "valid phone",
			phone:   "13900139000",
			wantErr: false,
		},
		{
			name:    "empty phone - allowed",
			phone:   "",
			wantErr: false,
		},
		{
			name:    "invalid phone",
			phone:   "12345",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := user.SetPhone(tt.phone)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.phone, user.Phone())
			}
		})
	}
}

func TestUser_SetPassword(t *testing.T) {
	user, _ := NewUser("testuser", "13800138000", "abc123", UserTypeDelivery, UserRolePublic)

	tests := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{
			name:     "valid password",
			password: "xyz789",
			wantErr:  false,
		},
		{
			name:     "invalid - too short",
			password: "ab12",
			wantErr:  true,
		},
		{
			name:     "invalid - only letters",
			password: "abcdef",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := user.SetPassword(tt.password)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// ==================== Business Logic Tests ====================

func TestUser_HasPhone(t *testing.T) {
	tests := []struct {
		name     string
		phone    string
		expected bool
	}{
		{
			name:     "has phone",
			phone:    "13800138000",
			expected: true,
		},
		{
			name:     "empty phone",
			phone:    "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, _ := NewUser("testuser", tt.phone, "abc123", UserTypeDelivery, UserRolePublic)
			got := user.HasPhone()
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestUser_ValidatePassword(t *testing.T) {
	user, _ := NewUser("testuser", "13800138000", "abc123", UserTypeDelivery, UserRolePublic)

	tests := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{
			name:     "valid password",
			password: "abc123",
			wantErr:  false,
		},
		{
			name:     "invalid - too short",
			password: "ab12",
			wantErr:  true,
		},
		{
			name:     "invalid - only letters",
			password: "abcdef",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := user.ValidatePassword(tt.password)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestUser_VerifyPassword(t *testing.T) {
	// Generate a real bcrypt hash for testing
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("abc123"), bcrypt.DefaultCost)
	assert.NoError(t, err)

	tests := []struct {
		name          string
		storedPass    string
		verifyPass    string
		wantErr       bool
		description   string
	}{
		{
			name:        "hashed password - correct",
			storedPass:  string(hashedPassword),
			verifyPass:  "abc123",
			wantErr:     false,
			description: "bcrypt verification should succeed",
		},
		{
			name:        "hashed password - incorrect",
			storedPass:  string(hashedPassword),
			verifyPass:  "wrongpass",
			wantErr:     true,
			description: "bcrypt verification should fail",
		},
		{
			name:        "plain password - correct",
			storedPass:  "plain123",
			verifyPass:  "plain123",
			wantErr:     false,
			description: "plain text comparison should succeed",
		},
		{
			name:        "plain password - incorrect",
			storedPass:  "plain123",
			verifyPass:  "wrongpass",
			wantErr:     true,
			description: "plain text comparison should fail",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, _ := NewUser("testuser", "13800138000", "abc123", UserTypeDelivery, UserRolePublic)
			// Directly set the stored password for testing
			user.password = value_objects.Password(tt.storedPass)

			err := user.VerifyPassword(tt.verifyPass)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// ==================== Model Conversion Tests ====================

func TestUser_ToModel(t *testing.T) {
	user, _ := NewUser("testuser", "13800138000", "plain123", UserTypeDelivery, UserRolePublic)
	user.SetAddress("test address")

	model := user.ToModel()

	assert.NotEqual(t, snowflake.ID(0), model.ID)
	assert.Equal(t, "testuser", model.Name)
	assert.Equal(t, "13800138000", model.Phone)
	// Password should be hashed
	assert.NotEqual(t, "plain123", model.Password)
	assert.True(t, strings.HasPrefix(model.Password, "$2a$") || strings.HasPrefix(model.Password, "$2b$"))
	assert.Equal(t, string(UserTypeDelivery), model.Type)
	assert.Equal(t, string(UserRolePublic), model.Role)
	assert.Equal(t, "test address", model.Address)
}

func TestUser_ToModel_WithHashedPassword(t *testing.T) {
	user, _ := NewUser("testuser", "13800138000", "abc123", UserTypeDelivery, UserRolePublic)

	// Set a pre-hashed password
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("abc123"), bcrypt.DefaultCost)
	user.password = value_objects.Password(string(hashedPassword))

	model := user.ToModel()

	// Should not re-hash an already hashed password
	assert.Equal(t, string(hashedPassword), model.Password)
}

func TestUserFromModel(t *testing.T) {
	userID := snowflake.ID(123456789)
	hashedPassword := "$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy"

	model := &models.User{
		ID:       userID,
		Name:     "testuser",
		Phone:    "13800138000",
		Password: hashedPassword,
		Type:     string(UserTypePickup),
		Role:     string(UserRolePrivate),
		Address:  "test address",
	}

	user := UserFromModel(model)

	assert.Equal(t, NewUserIDFromSnowflake(userID), user.ID())
	assert.Equal(t, "testuser", user.Name())
	assert.Equal(t, "13800138000", user.Phone())
	assert.Equal(t, hashedPassword, user.Password())
	assert.Equal(t, UserTypePickup, user.UserType())
	assert.Equal(t, UserRolePrivate, user.Role())
	assert.Equal(t, "test address", user.Address())
}
