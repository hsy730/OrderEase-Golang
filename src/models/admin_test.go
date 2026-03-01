package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestAdminStruct(t *testing.T) {
	now := time.Now()
	admin := Admin{
		ID:        1,
		Username:  "admin",
		Password:  "hashedpassword",
		CreatedAt: now,
		UpdatedAt: now,
	}

	assert.Equal(t, uint64(1), admin.ID)
	assert.Equal(t, "admin", admin.Username)
	assert.Equal(t, "hashedpassword", admin.Password)
	assert.Equal(t, now, admin.CreatedAt)
	assert.Equal(t, now, admin.UpdatedAt)
}

func TestAdminStructEmpty(t *testing.T) {
	admin := Admin{}

	assert.Zero(t, admin.ID)
	assert.Empty(t, admin.Username)
	assert.Empty(t, admin.Password)
	assert.True(t, admin.CreatedAt.IsZero())
	assert.True(t, admin.UpdatedAt.IsZero())
}

func TestAdmin_HashPassword(t *testing.T) {
	admin := Admin{
		Username: "testadmin",
		Password: "plainpassword",
	}

	err := admin.HashPassword()

	assert.NoError(t, err)
	assert.NotEqual(t, "plainpassword", admin.Password)
	assert.Len(t, admin.Password, 60)
}

func TestAdmin_HashPassword_Empty(t *testing.T) {
	admin := Admin{
		Username: "testadmin",
		Password: "",
	}

	err := admin.HashPassword()

	assert.NoError(t, err)
	assert.NotEmpty(t, admin.Password)
}

func TestAdmin_CheckPassword_Correct(t *testing.T) {
	admin := Admin{
		Username: "testadmin",
		Password: "plainpassword",
	}

	err := admin.HashPassword()
	assert.NoError(t, err)

	result := admin.CheckPassword("plainpassword")
	assert.True(t, result)
}

func TestAdmin_CheckPassword_Incorrect(t *testing.T) {
	admin := Admin{
		Username: "testadmin",
		Password: "plainpassword",
	}

	err := admin.HashPassword()
	assert.NoError(t, err)

	result := admin.CheckPassword("wrongpassword")
	assert.False(t, result)
}

func TestAdmin_CheckPassword_Empty(t *testing.T) {
	admin := Admin{
		Username: "testadmin",
		Password: "plainpassword",
	}

	err := admin.HashPassword()
	assert.NoError(t, err)

	result := admin.CheckPassword("")
	assert.False(t, result)
}

func TestAdmin_PasswordRoundTrip(t *testing.T) {
	passwords := []string{
		"simple",
		"with spaces",
		"with!@#$%^&*()special",
		"中文密码",
		"verylongpasswordverylongpasswordverylongpasswordverylongpassword",
	}

	for _, pwd := range passwords {
		t.Run("password: "+pwd, func(t *testing.T) {
			admin := Admin{Password: pwd}
			err := admin.HashPassword()
			assert.NoError(t, err)
			assert.True(t, admin.CheckPassword(pwd))
		})
	}
}

func TestUserInfoStruct(t *testing.T) {
	userInfo := UserInfo{
		UserID:   123,
		Username: "testuser",
		IsAdmin:  true,
	}

	assert.Equal(t, uint64(123), userInfo.UserID)
	assert.Equal(t, "testuser", userInfo.Username)
	assert.True(t, userInfo.IsAdmin)
}

func TestUserInfoStructEmpty(t *testing.T) {
	userInfo := UserInfo{}

	assert.Zero(t, userInfo.UserID)
	assert.Empty(t, userInfo.Username)
	assert.False(t, userInfo.IsAdmin)
}

func TestUserInfo_IsAdminFalse(t *testing.T) {
	userInfo := UserInfo{
		UserID:   456,
		Username: "regularuser",
		IsAdmin:  false,
	}

	assert.Equal(t, uint64(456), userInfo.UserID)
	assert.Equal(t, "regularuser", userInfo.Username)
	assert.False(t, userInfo.IsAdmin)
}
