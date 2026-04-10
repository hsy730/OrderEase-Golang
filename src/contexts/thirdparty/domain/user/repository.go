package user

import (
	"orderease/contexts/thirdparty/domain/oauth"
	"orderease/models"
)

type UserBindingRepository interface {
	FindByProviderAndUserID(provider oauth.Provider, providerUserID string) (*models.UserThirdpartyBinding, error)
	FindByUserIDAndProvider(userID uint64, provider oauth.Provider) (*models.UserThirdpartyBinding, error)
	Create(binding *models.UserThirdpartyBinding) error
	Update(binding *models.UserThirdpartyBinding) error
}
