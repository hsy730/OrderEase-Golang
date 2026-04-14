package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/bwmarrin/snowflake"
)

// SnowflakeString 是一个包装类型，用于将 snowflake.ID 序列化为字符串
type SnowflakeString snowflake.ID

// MarshalJSON 将 SnowflakeString 序列化为 JSON 字符串
func (s SnowflakeString) MarshalJSON() ([]byte, error) {
	return json.Marshal(strconv.FormatInt(int64(s), 10))
}

// UnmarshalJSON 从 JSON 字符串或数字反序列化为 SnowflakeString
func (s *SnowflakeString) UnmarshalJSON(data []byte) error {
	// 尝试先作为字符串解析
	var str string
	if err := json.Unmarshal(data, &str); err == nil {
		id, err := strconv.ParseInt(str, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid snowflake string: %s", str)
		}
		*s = SnowflakeString(id)
		return nil
	}

	// 尝试作为数字解析
	var num int64
	if err := json.Unmarshal(data, &num); err != nil {
		return fmt.Errorf("snowflake ID must be string or number, got: %s", string(data))
	}
	*s = SnowflakeString(num)
	return nil
}

// String 返回字符串表示
func (s SnowflakeString) String() string {
	return strconv.FormatInt(int64(s), 10)
}

// Int64 返回 int64 表示
func (s SnowflakeString) Int64() int64 {
	return int64(s)
}

// ToSnowflakeID 转换为 snowflake.ID
func (s SnowflakeString) ToSnowflakeID() snowflake.ID {
	return snowflake.ID(s)
}

// FromSnowflakeID 从 snowflake.ID 创建 SnowflakeString
func FromSnowflakeID(id snowflake.ID) SnowflakeString {
	return SnowflakeString(id)
}

// Value 实现 driver.Valuer 接口，用于数据库操作
func (s SnowflakeString) Value() (driver.Value, error) {
	return int64(s), nil
}

// Scan 实现 sql.Scanner 接口，用于数据库操作
func (s *SnowflakeString) Scan(value interface{}) error {
	switch v := value.(type) {
	case int64:
		*s = SnowflakeString(v)
		return nil
	case int:
		*s = SnowflakeString(int64(v))
		return nil
	case uint64:
		*s = SnowflakeString(int64(v))
		return nil
	case []byte:
		id, err := strconv.ParseInt(string(v), 10, 64)
		if err != nil {
			return err
		}
		*s = SnowflakeString(id)
		return nil
	default:
		return fmt.Errorf("cannot scan %T into SnowflakeString", value)
	}
}
