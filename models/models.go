package models

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"
)

type Product struct {
	ID          uint      `gorm:"primarykey" json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Price       float64   `json:"price"`
	Stock       int       `json:"stock"`
	ImageURL    string    `json:"image_url"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Price float64

func (p *Price) UnmarshalJSON(data []byte) error {
	var value interface{}
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}

	switch v := value.(type) {
	case float64:
		*p = Price(v)
	case string:
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			*p = Price(f)
		} else {
			return fmt.Errorf("invalid price format: %s", v)
		}
	default:
		return fmt.Errorf("invalid price type: %T", value)
	}

	return nil
}

type Order struct {
	ID         uint        `gorm:"primarykey" json:"id"`
	UserID     uint        `json:"user_id"`
	TotalPrice Price       `json:"total_price"`
	Status     string      `json:"status"`
	Remark     string      `json:"remark"`
	Items      []OrderItem `gorm:"foreignKey:OrderID" json:"items"`
	CreatedAt  time.Time   `json:"created_at"`
	UpdatedAt  time.Time   `json:"updated_at"`
}

type OrderItem struct {
	ID        uint    `gorm:"primarykey" json:"id"`
	OrderID   uint    `json:"order_id"`
	ProductID uint    `json:"product_id"`
	Quantity  int     `json:"quantity"`
	Price     Price   `json:"price"`
	Product   Product `gorm:"foreignKey:ProductID" json:"product"`
}
