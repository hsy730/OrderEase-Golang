package shared

import (
	"encoding/json"
	"fmt"

	"github.com/bwmarrin/snowflake"
	"orderease/utils"
)

type ID snowflake.ID

func NewID() ID {
	return ID(utils.GenerateSnowflakeID())
}

func (id ID) Value() snowflake.ID {
	return snowflake.ID(id)
}

func (id ID) String() string {
	return snowflake.ID(id).String()
}

func (id ID) IsZero() bool {
	return id == 0
}

func ParseIDFromString(s string) (ID, error) {
	id, err := snowflake.ParseString(s)
	return ID(id), err
}

func ParseIDFromUint64(u uint64) ID {
	return ID(u)
}

func (id ID) ToUint64() uint64 {
	return uint64(id)
}

// UnmarshalJSON implements json.Unmarshaler interface
// Supports parsing from string, float64 (number in JSON), and int
func (id *ID) UnmarshalJSON(data []byte) error {
	// Handle null
	if len(data) == 0 || string(data) == "null" {
		*id = ID(0)
		return nil
	}

	// Try to parse as string (quoted)
	if len(data) >= 2 && data[0] == '"' && data[len(data)-1] == '"' {
		var s string
		if err := json.Unmarshal(data, &s); err != nil {
			return err
		}
		parsed, err := snowflake.ParseString(s)
		if err != nil {
			return fmt.Errorf("failed to parse ID from string %q: %w", s, err)
		}
		*id = ID(parsed)
		return nil
	}

	// Try to parse as float64 (JSON numbers are float64)
	var f float64
	if err := json.Unmarshal(data, &f); err != nil {
		return err
	}
	*id = ID(uint64(f))
	return nil
}

// MarshalJSON implements json.Marshaler interface
func (id ID) MarshalJSON() ([]byte, error) {
	return json.Marshal(id.String())
}

// UnmarshalText implements encoding.TextUnmarshaler interface
// Supports parsing from string format
func (id *ID) UnmarshalText(data []byte) error {
	parsed, err := snowflake.ParseString(string(data))
	if err != nil {
		return fmt.Errorf("failed to parse ID from text %q: %w", string(data), err)
	}
	*id = ID(parsed)
	return nil
}

// MarshalText implements encoding.TextMarshaler interface
func (id ID) MarshalText() ([]byte, error) {
	return []byte(id.String()), nil
}

// Scan implements sql.Scanner interface for database integration
func (id *ID) Scan(value interface{}) error {
	if value == nil {
		*id = ID(0)
		return nil
	}

	switch v := value.(type) {
	case []byte:
		parsed, err := snowflake.ParseString(string(v))
		if err != nil {
			return err
		}
		*id = ID(parsed)
	case string:
		parsed, err := snowflake.ParseString(v)
		if err != nil {
			return err
		}
		*id = ID(parsed)
	case int64:
		*id = ID(v)
	case uint64:
		*id = ID(v)
	case float64:
		*id = ID(uint64(v))
	default:
		return fmt.Errorf("unsupported type for ID: %T", value)
	}
	return nil
}
