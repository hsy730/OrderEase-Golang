package shop

import (
	"errors"
	"time"
)

type Tag struct {
	ID          int
	ShopID      uint64
	Name        string
	Description string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func NewTag(shopID uint64, name, description string) (*Tag, error) {
	if shopID == 0 {
		return nil, errors.New("店铺ID不能为空")
	}

	if name == "" {
		return nil, errors.New("标签名称不能为空")
	}

	now := time.Now()

	return &Tag{
		ID:          0,
		ShopID:      shopID,
		Name:        name,
		Description: description,
		CreatedAt:   now,
		UpdatedAt:   now,
	}, nil
}

func (t *Tag) Update(name, description string) error {
	if name != "" {
		t.Name = name
	}
	if description != "" {
		t.Description = description
	}
	t.UpdatedAt = time.Now()
	return nil
}
