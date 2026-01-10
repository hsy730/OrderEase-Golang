package shared

import (
	"database/sql/driver"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewPrice(t *testing.T) {
	tests := []struct {
		name  string
		value float64
		want  Price
	}{
		{"positive price", 100.5, Price(100.5)},
		{"zero price", 0, Price(0)},
		{"negative price", -50.25, Price(-50.25)},
		{"large price", 999999.99, Price(999999.99)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewPrice(tt.value)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestPrice_Add(t *testing.T) {
	tests := []struct {
		name  string
		p     Price
		other Price
		want  Price
	}{
		{"add positive", Price(100), Price(50), Price(150)},
		{"add zero", Price(100), Price(0), Price(100)},
		{"add negative", Price(100), Price(-30), Price(70)},
		{"add decimals", Price(100.5), Price(50.25), Price(150.75)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.p.Add(tt.other)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestPrice_Multiply(t *testing.T) {
	tests := []struct {
		name     string
		p        Price
		quantity int
		want     Price
	}{
		{"multiply positive", Price(100), 2, Price(200)},
		{"multiply by zero", Price(100), 0, Price(0)},
		{"multiply by one", Price(100), 1, Price(100)},
		{"multiply negative", Price(100), -3, Price(-300)},
		{"multiply decimals", Price(10.5), 3, Price(31.5)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.p.Multiply(tt.quantity)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestPrice_IsZero(t *testing.T) {
	tests := []struct {
		name string
		p    Price
		want bool
	}{
		{"zero price", Price(0), true},
		{"positive price", Price(100), false},
		{"negative price", Price(-50), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.p.IsZero()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestPrice_IsPositive(t *testing.T) {
	tests := []struct {
		name string
		p    Price
		want bool
	}{
		{"positive price", Price(100), true},
		{"zero price", Price(0), false},
		{"negative price", Price(-50), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.p.IsPositive()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestPrice_ToFloat64(t *testing.T) {
	tests := []struct {
		name  string
		p     Price
		want  float64
	}{
		{"positive", Price(100.5), 100.5},
		{"zero", Price(0), 0},
		{"negative", Price(-50.25), -50.25},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.p.ToFloat64()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestPrice_String(t *testing.T) {
	tests := []struct {
		name string
		p    Price
		want string
	}{
		{"integer", Price(100), "100.00"},
		{"decimal", Price(100.5), "100.50"},
		{"two decimals", Price(100.56), "100.56"},
		{"more decimals", Price(100.567), "100.57"},
		{"zero", Price(0), "0.00"},
		{"negative", Price(-50.5), "-50.50"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.p.String()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestPrice_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		data    string
		want    Price
		wantErr bool
		errMsg  string
	}{
		// 有效类型
		{"valid float64", `100.5`, Price(100.5), false, ""},
		{"valid float32", `100.5`, Price(100.5), false, ""},
		{"valid string", `"100.5"`, Price(100.5), false, ""},
		{"valid int", `100`, Price(100), false, ""},
		{"valid int64", `100`, Price(100), false, ""},
		{"valid int32", `100`, Price(100), false, ""},
		{"zero float64", `0`, Price(0), false, ""},
		{"zero string", `"0"`, Price(0), false, ""},
		// 无效类型
		{"invalid type - bool", `true`, Price(0), true, "invalid price type"},
		{"invalid type - object", `{}`, Price(0), true, "invalid price type"},
		{"invalid type - array", `[]`, Price(0), true, "invalid price type"},
		{"invalid type - null", `null`, Price(0), true, "invalid price type"},
		// 无效字符串格式
		{"invalid string - letters", `"abc"`, Price(0), true, "invalid price format"},
		{"invalid string - mixed", `"100abc"`, Price(0), true, "invalid price format"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var p Price
			data := []byte(tt.data)
			err := json.Unmarshal(data, &p)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, p)
			}
		})
	}
}

func TestPrice_Scan(t *testing.T) {
	tests := []struct {
		name    string
		value   interface{}
		want    Price
		wantErr bool
		errMsg  string
	}{
		// 有效类型
		{"float64", float64(100.5), Price(100.5), false, ""},
		{"int64", int64(100), Price(100), false, ""},
		{"[]uint8 valid", []uint8("100.5"), Price(100.5), false, ""},
		{"[]uint8 zero", []uint8("0"), Price(0), false, ""},
		{"[]uint8 negative", []uint8("-50.25"), Price(-50.25), false, ""},
		// 无效类型
		{"unsupported type - string", "100.5", Price(0), true, "unsupported Scan"},
		{"unsupported type - bool", true, Price(0), true, "unsupported Scan"},
		{"unsupported type - nil", nil, Price(0), true, "unsupported Scan"},
		// 无效字符串格式
		{"[]uint8 invalid format", []uint8("abc"), Price(0), true, "failed to parse Price"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var p Price
			err := p.Scan(tt.value)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, p)
			}
		})
	}
}

func TestPrice_Value(t *testing.T) {
	tests := []struct {
		name    string
		p       Price
		want    driver.Value
		wantErr bool
	}{
		{"positive", Price(100.5), float64(100.5), false},
		{"zero", Price(0), float64(0), false},
		{"negative", Price(-50.25), float64(-50.25), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.p.Value()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestPrice_JSONMarshalUnmarshal(t *testing.T) {
	tests := []struct {
		name  string
		price Price
	}{
		{"positive", Price(100.5)},
		{"zero", Price(0)},
		{"negative", Price(-50.25)},
		{"large", Price(999999.99)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Marshal
			data, err := json.Marshal(tt.price)
			assert.NoError(t, err)

			// Unmarshal
			var got Price
			err = json.Unmarshal(data, &got)
			assert.NoError(t, err)
			assert.Equal(t, tt.price, got)
		})
	}
}
