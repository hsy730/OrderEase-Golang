package tag

import "orderease/models"

type Repository interface {
	GetByID(id string) (*models.Tag, error)
	Create(tag *models.Tag) error
	Update(tag *models.Tag) error
	Delete(tag *models.Tag) error
	CountAssociatedProducts(tagID string) (int64, error)
}
