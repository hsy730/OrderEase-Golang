package models

import (
	"encoding/json"
	"fmt"
	"strconv"
)

type Price float64

func (p *Price) Value() float64 {
	return float64(*p)
}

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
