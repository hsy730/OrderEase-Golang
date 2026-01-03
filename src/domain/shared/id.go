package shared

import (
	"github.com/bwmarrin/snowflake"
)

type ID snowflake.ID

func NewID() ID {
	return ID(snowflake.ID(0))
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
