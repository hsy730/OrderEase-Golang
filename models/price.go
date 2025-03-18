package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"strconv"
)

type Price float64

// GORM's Scanner interface for mapping database value to custom type
func (p *Price) Scan(value interface{}) error {
	switch v := value.(type) {
	case float64:
		*p = Price(v)
	case int64:
		*p = Price(float64(v))
	default:
		return fmt.Errorf("unsupported Scan, storing driver.Value type %T into type *Price", value)
	}
	return nil
}

// GORM's Valuer interface for serializing custom type to database value
func (p Price) Value() (driver.Value, error) {
	return float64(p), nil
}

// Custom unmarshal method for JSON, if you're also using JSON
func (p *Price) UnmarshalJSON(data []byte) error {
	var value interface{}
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}

	switch v := value.(type) {
	case float64:
		*p = Price(v)
	case float32:
		*p = Price(float64(v))
	case string:
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			*p = Price(f)
		} else {
			return fmt.Errorf("invalid price format: %s", v)
		}
	case int:
		*p = Price(float64(v))
	case int64:
		*p = Price(float64(v))
	case int32:
		*p = Price(float64(v))
	default:
		return fmt.Errorf("invalid price type: %T", value)
	}
	return nil
}
