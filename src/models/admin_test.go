package models

import (
	"testing"

	"golang.org/x/crypto/bcrypt"
	"github.com/stretchr/testify/assert"
)

func TestAdmin_HashPassword(t *testing.T) {
	admin := &Admin{
		Username: "test_admin",
		Password: "plain_password",
	}

	// Hash the password
	err := admin.HashPassword()
	assert.NoError(t, err)

	// Password should be hashed (bcrypt format starts with $2a$ or $2b$)
	assert.True(t, len(admin.Password) > 0)
	assert.True(t, len(admin.Password) != len("plain_password"))
	assert.True(t, len(admin.Password) >= 60) // bcrypt hash is typically 60 chars

	// Verify hash can be used to check the password
	err = bcrypt.CompareHashAndPassword([]byte(admin.Password), []byte("plain_password"))
	assert.NoError(t, err)

	// bcrypt format should start with $2a$ or $2b$
	assert.True(t, len(admin.Password) > 4)
	assert.True(t, admin.Password[:4] == "$2a$" || admin.Password[:4] == "$2b$")
}

func TestAdmin_HashPassword_Empty(t *testing.T) {
	// bcrypt.GenerateFromPassword should always succeed, even with empty password
	admin := &Admin{Password: ""}
	err := admin.HashPassword()
	assert.NoError(t, err)

	// Empty password should still be hashed
	assert.True(t, len(admin.Password) > 0)
	assert.True(t, admin.Password[:4] == "$2a$" || admin.Password[:4] == "$2b$")
}

func TestAdmin_HashPassword_AlreadyHashed(t *testing.T) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("original_password"), bcrypt.DefaultCost)
	assert.NoError(t, err)

	admin := &Admin{
		Username: "test_admin",
		Password: string(hashedPassword),
	}

	// Hashing an already hashed password should still work (it will re-hash)
	err = admin.HashPassword()
	assert.NoError(t, err)

	// The password should be re-hashed
	assert.NotEqual(t, string(hashedPassword), admin.Password)
}

func TestAdmin_CheckPassword(t *testing.T) {
	admin := &Admin{
		Username: "test_admin",
	}
	admin.Password = "test123"

	// Hash the password first
	admin.HashPassword()

	tests := []struct {
		name     string
		password string
		expected bool
	}{
		{"correct password", "test123", true},
		{"wrong password", "wrongpass", false},
		{"empty password", "", false},
		{"password with extra spaces", "test123 ", false},
		{"case sensitive password", "Test123", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := admin.CheckPassword(tt.password)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestAdmin_CheckPassword_WithBcryptHash(t *testing.T) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("secret123"), bcrypt.DefaultCost)
	assert.NoError(t, err)

	admin := &Admin{
		Username: "test_admin",
		Password: string(hashedPassword),
	}

	// Correct password should pass
	assert.True(t, admin.CheckPassword("secret123"))

	// Wrong password should fail
	assert.False(t, admin.CheckPassword("wrongpass"))

	// Empty password should fail
	assert.False(t, admin.CheckPassword(""))

	// Different case should fail (bcrypt handles password comparison internally)
	assert.False(t, admin.CheckPassword("Secret123"))
}

func TestAdmin_CheckPassword_ComplexPassword(t *testing.T) {
	// Test with a complex password that meets strict requirements
	complexPassword := "MyStr0ng!Pass"
	admin := &Admin{
		Username: "test_admin",
	}
	admin.Password = complexPassword
	admin.HashPassword()

	assert.True(t, admin.CheckPassword(complexPassword))
	assert.False(t, admin.CheckPassword("MyStr0ng!pas")) // wrong case
}

func TestAdmin_BasicPasswordHashingAndChecking(t *testing.T) {
	// Complete flow: set password, hash, then verify
	admin := &Admin{
		Username: "test_admin",
		Password: "initial_password",
	}

	// Hash the password
	err := admin.HashPassword()
	assert.NoError(t, err)

	// Store the hashed password
	hashedPassword := admin.Password

	// Check if original password matches
	assert.True(t, admin.CheckPassword("initial_password"))

	// Recreate a new admin with hashed password
	admin2 := &Admin{
		Username: "test_admin",
		Password: hashedPassword,
	}

	// The password should still be valid
	assert.True(t, admin2.CheckPassword("initial_password"))
	assert.False(t, admin2.CheckPassword("wrong_password"))
}
