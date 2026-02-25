package models

import (
	"testing"
	"time"

	"github.com/bwmarrin/snowflake"
	"github.com/stretchr/testify/assert"
)

func TestUserTypeConstants(t *testing.T) {
	t.Run("delivery type", func(t *testing.T) {
		assert.Equal(t, "delivery", UserTypeDelivery)
	})

	t.Run("pickup type", func(t *testing.T) {
		assert.Equal(t, "pickup", UserTypePickup)
	})
}

func TestUserRoleConstants(t *testing.T) {
	t.Run("private role", func(t *testing.T) {
		assert.Equal(t, "private_user", UserRolePrivate)
	})

	t.Run("public role", func(t *testing.T) {
		assert.Equal(t, "public_user", UserRolePublic)
	})
}

func TestUserStruct(t *testing.T) {
	now := time.Now()
	user := User{
		ID:       snowflake.ID(123),
		Name:     "testuser",
		Role:     UserRolePublic,
		Password: "hashedpassword",
		Phone:    "13800138000",
		Address:  "123 Test Street",
		Type:     UserTypeDelivery,
		Nickname: "Test User",
		CreatedAt: now,
		UpdatedAt: now,
	}

	assert.Equal(t, snowflake.ID(123), user.ID)
	assert.Equal(t, "testuser", user.Name)
	assert.Equal(t, UserRolePublic, user.Role)
	assert.Equal(t, "hashedpassword", user.Password)
	assert.Equal(t, "13800138000", user.Phone)
	assert.Equal(t, "123 Test Street", user.Address)
	assert.Equal(t, UserTypeDelivery, user.Type)
	assert.Equal(t, "Test User", user.Nickname)
}

func TestUserStructWithOrders(t *testing.T) {
	order := Order{
		ID:    snowflake.ID(456),
		Status: 1,
	}

	user := User{
		ID:     snowflake.ID(123),
		Name:   "testuser",
		Orders: []Order{order},
	}

	assert.Len(t, user.Orders, 1)
	assert.Equal(t, snowflake.ID(456), user.Orders[0].ID)
}

func TestUserStructWithThirdpartyBindings(t *testing.T) {
	binding := UserThirdpartyBinding{
		UserID: snowflake.ID(123),
	}

	user := User{
		ID:               snowflake.ID(123),
		Name:             "testuser",
		ThirdpartyBindings: []UserThirdpartyBinding{binding},
	}

	assert.Len(t, user.ThirdpartyBindings, 1)
	assert.Equal(t, snowflake.ID(123), user.ThirdpartyBindings[0].UserID)
}

func TestUserStructEmpty(t *testing.T) {
	user := User{}

	assert.Zero(t, user.ID)
	assert.Empty(t, user.Name)
	assert.Empty(t, user.Role)
	assert.Empty(t, user.Password)
	assert.Empty(t, user.Phone)
	assert.Empty(t, user.Address)
	assert.Empty(t, user.Type)
	assert.Empty(t, user.Nickname)
	assert.Nil(t, user.Orders)
	assert.Nil(t, user.ThirdpartyBindings)
}
