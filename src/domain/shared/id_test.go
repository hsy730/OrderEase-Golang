package shared

import (
	"orderease/utils"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewID(t *testing.T) {
	id := NewID()
	assert.True(t, id.IsZero(), "NewID should return zero ID")
}

func TestID_IsZero(t *testing.T) {
	tests := []struct {
		name string
		id   ID
		want bool
	}{
		{"zero ID", ID(0), true},
		{"non-zero ID", ID(123), false},
		{"large ID", ID(9999999999), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.id.IsZero()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestParseIDFromString(t *testing.T) {
	tests := []struct {
		name     string
		s        string
		wantErr  bool
		validate func(*testing.T, ID)
	}{
		{
			name:    "valid numeric string",
			s:       "123456",
			wantErr: false,
			validate: func(t *testing.T, id ID) {
				assert.False(t, id.IsZero())
			},
		},
		{
			name:    "valid large number string",
			s:       "12345678901234",
			wantErr: false,
			validate: func(t *testing.T, id ID) {
				assert.False(t, id.IsZero())
			},
		},
		{
			name:    "empty string",
			s:       "",
			wantErr: true,
		},
		{
			name:    "invalid string - letters",
			s:       "abc123",
			wantErr: true,
		},
		{
			name:    "invalid string - special chars",
			s:       "12-34",
			wantErr: true,
		},
		{
			name:    "zero string",
			s:       "0",
			wantErr: false,
			validate: func(t *testing.T, id ID) {
				assert.True(t, id.IsZero())
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseIDFromString(tt.s)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tt.validate != nil {
					tt.validate(t, got)
				}
			}
		})
	}
}

func TestParseIDFromUint64(t *testing.T) {
	tests := []struct {
		name string
		u    uint64
		want ID
	}{
		{"zero", 0, ID(0)},
		{"positive", 123, ID(123)},
		{"large number", 12345678901234, ID(12345678901234)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseIDFromUint64(tt.u)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestID_ToUint64(t *testing.T) {
	tests := []struct {
		name string
		id   ID
		want uint64
	}{
		{"zero", ID(0), 0},
		{"positive", ID(123), 123},
		{"large", ID(12345678901234), 12345678901234},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.id.ToUint64()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestID_Value(t *testing.T) {
	tests := []struct {
		name string
		id   ID
		want uint64
	}{
		{"zero", ID(0), 0},
		{"positive", ID(123), 123},
		{"large", ID(999999), 999999},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.id.Value()
			assert.Equal(t, tt.want, uint64(got))
		})
	}
}

func TestID_String(t *testing.T) {
	tests := []struct {
		name     string
		id       ID
		notEmpty bool
	}{
		{"zero ID", ID(0), true},
		{"positive ID", ID(123), true},
		{"large ID", ID(12345678901234), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.id.String()
			if tt.notEmpty {
				assert.NotEmpty(t, got)
			}
		})
	}
}

func TestID_Comparisons(t *testing.T) {
	tests := []struct {
		name      string
		id1       ID
		id2       ID
		wantEqual bool
		wantLess  bool
	}{
		{"equal", ID(100), ID(100), true, false},
		{"less than", ID(50), ID(100), false, true},
		{"greater than", ID(100), ID(50), false, false},
		{"zero vs positive", ID(0), ID(1), false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test equality
			if tt.wantEqual {
				assert.Equal(t, tt.id1, tt.id2)
			} else {
				assert.NotEqual(t, tt.id1, tt.id2)
			}

			// Test less than
			if tt.wantLess {
				assert.Less(t, tt.id1, tt.id2)
			}

			// Test conversion consistency
			assert.Equal(t, tt.id1, ParseIDFromUint64(tt.id1.ToUint64()))
		})
	}
}

func TestID_GenerateSnowflakeID(t *testing.T) {
	id := ID(utils.GenerateSnowflakeID())
	assert.False(t, id.IsZero(), "Generated Snowflake ID should not be zero")
}
