package models

import (
	"testing"
	"time"

	"github.com/bwmarrin/snowflake"
	"github.com/stretchr/testify/assert"
)

func TestHashShopPassword(t *testing.T) {
	shop := &Shop{
		ID:            snowflake.ID(123),
		Name:          "Test Shop",
		OwnerUsername: "shopowner",
		OwnerPassword: "plainpassword",
	}

	err := HashShopPassword(shop)

	assert.NoError(t, err)
	assert.NotEqual(t, "plainpassword", shop.OwnerPassword)
	assert.Len(t, shop.OwnerPassword, 60)
}

func TestHashShopPassword_Empty(t *testing.T) {
	shop := &Shop{
		ID:            snowflake.ID(123),
		Name:          "Test Shop",
		OwnerUsername: "shopowner",
		OwnerPassword: "",
	}

	err := HashShopPassword(shop)

	assert.NoError(t, err)
	assert.NotEmpty(t, shop.OwnerPassword)
}

func TestHashShopPassword_VerifyWithBcrypt(t *testing.T) {
	originalPassword := "mypassword123"
	shop := &Shop{
		ID:            snowflake.ID(123),
		Name:          "Test Shop",
		OwnerUsername: "shopowner",
		OwnerPassword: originalPassword,
	}

	err := HashShopPassword(shop)
	assert.NoError(t, err)

	hashedPassword := shop.OwnerPassword

	shop2 := &Shop{
		OwnerPassword: originalPassword,
	}
	err = HashShopPassword(shop2)
	assert.NoError(t, err)

	assert.NotEqual(t, hashedPassword, shop2.OwnerPassword)
}

func TestHashShopPassword_DifferentShops(t *testing.T) {
	shop1 := &Shop{
		ID:            snowflake.ID(123),
		OwnerPassword: "password1",
	}
	shop2 := &Shop{
		ID:            snowflake.ID(456),
		OwnerPassword: "password2",
	}

	err := HashShopPassword(shop1)
	assert.NoError(t, err)

	err = HashShopPassword(shop2)
	assert.NoError(t, err)

	assert.NotEqual(t, shop1.OwnerPassword, shop2.OwnerPassword)
}

func TestHashShopPassword_SamePassword(t *testing.T) {
	samePassword := "samepassword123"

	shop1 := &Shop{
		ID:            snowflake.ID(123),
		OwnerPassword: samePassword,
	}
	shop2 := &Shop{
		ID:            snowflake.ID(456),
		OwnerPassword: samePassword,
	}

	err := HashShopPassword(shop1)
	assert.NoError(t, err)

	err = HashShopPassword(shop2)
	assert.NoError(t, err)

	assert.NotEqual(t, shop1.OwnerPassword, shop2.OwnerPassword)
	assert.NotEqual(t, shop1.OwnerPassword, samePassword)
	assert.NotEqual(t, shop2.OwnerPassword, samePassword)
}

func TestHashShopPassword_PreservesOtherFields(t *testing.T) {
	now := time.Now()
	shop := &Shop{
		ID:            snowflake.ID(123),
		Name:          "Test Shop",
		OwnerUsername: "shopowner",
		OwnerPassword: "plainpassword",
		ContactPhone:  "13800138000",
		ContactEmail:  "shop@example.com",
		Address:       "123 Shop Street",
		ImageURL:      "https://example.com/shop.jpg",
		Description:   "A test shop",
		CreatedAt:     now,
		UpdatedAt:     now,
		ValidUntil:    now.AddDate(1, 0, 0),
	}

	originalName := shop.Name
	originalUsername := shop.OwnerUsername
	originalPhone := shop.ContactPhone
	originalEmail := shop.ContactEmail
	originalAddress := shop.Address

	err := HashShopPassword(shop)

	assert.NoError(t, err)
	assert.Equal(t, originalName, shop.Name)
	assert.Equal(t, originalUsername, shop.OwnerUsername)
	assert.Equal(t, originalPhone, shop.ContactPhone)
	assert.Equal(t, originalEmail, shop.ContactEmail)
	assert.Equal(t, originalAddress, shop.Address)
}

func TestHashShopPassword_WithSpecialCharacters(t *testing.T) {
	passwords := []string{
		"with!@#$%^&*()special",
		"with spaces and\ttabs",
		"中文密码测试",
		"verylongpasswordverylongpasswordverylongpasswordverylongpassword",
	}

	for _, pwd := range passwords {
		t.Run("password with special chars", func(t *testing.T) {
			shop := &Shop{
				ID:            snowflake.ID(123),
				OwnerPassword: pwd,
			}

			err := HashShopPassword(shop)
			assert.NoError(t, err)
			assert.NotEqual(t, pwd, shop.OwnerPassword)
		})
	}
}
