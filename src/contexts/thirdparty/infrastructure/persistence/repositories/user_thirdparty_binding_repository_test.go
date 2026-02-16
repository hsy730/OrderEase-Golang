package repositories

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"orderease/contexts/thirdparty/domain/oauth"
)

// ==================== Constructor Tests ====================

func TestNewUserThirdpartyBindingRepository(t *testing.T) {
	repo := NewUserThirdpartyBindingRepository(nil)

	assert.NotNil(t, repo)
}

// ==================== Provider Enum Tests ====================

func TestProviderWeChat(t *testing.T) {
	assert.Equal(t, "wechat", oauth.ProviderWeChat.String())
}

func TestProvider(t *testing.T) {
	provider := oauth.Provider("alipay")
	assert.Equal(t, "alipay", provider.String())
}

// ==================== Integration Tests (Skip) ====================

func TestUserThirdpartyBindingRepository_Create(t *testing.T) {
	t.Skip("Integration test - requires real database connection")
}

func TestUserThirdpartyBindingRepository_Update(t *testing.T) {
	t.Skip("Integration test - requires real database connection")
}

func TestUserThirdpartyBindingRepository_UpdateLastLogin(t *testing.T) {
	t.Skip("Integration test - requires real database connection")
}

func TestUserThirdpartyBindingRepository_Deactivate(t *testing.T) {
	t.Skip("Integration test - requires real database connection")
}

func TestUserThirdpartyBindingRepository_Delete(t *testing.T) {
	t.Skip("Integration test - requires real database connection")
}

func TestUserThirdpartyBindingRepository_ListActive(t *testing.T) {
	t.Skip("Integration test - requires real database connection")
}

func TestUserThirdpartyBindingRepository_CountByProvider(t *testing.T) {
	t.Skip("Integration test - requires real database connection")
}
