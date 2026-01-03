package user

import (
	"orderease/domain/shared"
)

type UserRepository interface {
	Save(user *User) error
	FindByID(id shared.ID) (*User, error)
	FindByName(name string) (*User, error)
	FindAll(page, pageSize int) ([]User, int64, error)
	Delete(id shared.ID) error
	Update(user *User) error
	Exists(id shared.ID) (bool, error)
}
