package shared

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"strconv"
)

type Price float64

func (p Price) String() string {
	return fmt.Sprintf("%.2f", p)
}

func (p *Price) Scan(value interface{}) error {
	switch v := value.(type) {
	case float64:
		*p = Price(v)
	case int64:
		*p = Price(float64(v))
	case []uint8:
		if f, err := strconv.ParseFloat(string(v), 64); err == nil {
			*p = Price(f)
		} else {
			return fmt.Errorf("failed to parse Price from string: %v", err)
		}
	default:
		return fmt.Errorf("unsupported Scan, storing driver.Value type %T into type *Price", value)
	}
	return nil
}

func (p Price) Value() (driver.Value, error) {
	return float64(p), nil
}

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

func (p Price) ToFloat64() float64 {
	return float64(p)
}

func NewPrice(value float64) Price {
	return Price(value)
}

func (p Price) Add(other Price) Price {
	return p + other
}

func (p Price) Multiply(quantity int) Price {
	return p * Price(quantity)
}

func (p Price) IsZero() bool {
	return p == 0
}

func (p Price) IsPositive() bool {
	return p > 0
}
