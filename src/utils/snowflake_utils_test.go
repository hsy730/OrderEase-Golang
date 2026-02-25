package utils

import (
	"strconv"
	"testing"
	"time"
	"sync"

	"github.com/bwmarrin/snowflake"
	"github.com/stretchr/testify/assert"
)

func TestGenerateSnowflakeID(t *testing.T) {
	// Test that generated IDs are unique
	ids := make(map[snowflake.ID]bool)

	for i := 0; i < 100; i++ {
		id := GenerateSnowflakeID()
		assert.NotZero(t, id)
		assert.False(t, ids[id])
		ids[id] = true
	}
}

func TestGenerateSnowflakeID_UniquenessInConcurrent(t *testing.T) {
	// Test that IDs are unique in concurrent scenarios
	ids := make(map[snowflake.ID]bool)
	var wg sync.WaitGroup
	var mu sync.Mutex

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 10; j++ {
				id := GenerateSnowflakeID()
				mu.Lock()
				exists := ids[id]
				ids[id] = true
				mu.Unlock()
				assert.False(t, exists)
			}
		}()
	}

	wg.Wait()
	assert.Len(t, ids, 100)
}

func TestGenerateSnowflakeID_Format(t *testing.T) {
	// Test that generated IDs are valid snowflake IDs
	id := GenerateSnowflakeID()

	// Snowflake ID should be positive
	assert.Greater(t, id, snowflake.ID(0))

	// Test string representation
	idStr := id.String()
	assert.NotEmpty(t, idStr)

	// Should be able to parse back
	parsed, err := strconv.ParseInt(idStr, 10, 64)
	assert.NoError(t, err)
	assert.Equal(t, id, snowflake.ID(parsed))
}

func TestGenerateSnowflakeID_Timestamp(t *testing.T) {
	// Test that generated IDs have reasonable timestamps
	id1 := GenerateSnowflakeID()
	time.Sleep(10 * time.Millisecond)
	id2 := GenerateSnowflakeID()

	// IDs should be different
	assert.NotEqual(t, id1, id2)

	// Second ID should be greater (since time has passed)
	assert.Greater(t, id2, id1)
}

func TestStringToSnowflakeID(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		expect  snowflake.ID
		wantErr bool
	}{
		{
			name:   "valid ID",
			input:  "123456789",
			expect:  123456789,
			wantErr: false,
		},
		{
			name:   "zero",
			input:  "0",
			expect:  0,
			wantErr: false,
		},
		{
			name:   "large positive ID",
			input:  "9223372036854775807",
			expect:  9223372036854775807,
			wantErr: false,
		},
		{
			name:   "invalid - empty",
			input:  "",
			expect:  0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, err := StringToSnowflakeID(tt.input)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Zero(t, id)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expect, id)
			}
		})
	}
}

func TestStringToSnowflakeID_Roundtrip(t *testing.T) {
	// Test that StringToSnowflakeID(GenerateSnowflakeID().String()) works
	originalID := GenerateSnowflakeID()
	idStr := originalID.String()

	parsedID, err := StringToSnowflakeID(idStr)
	assert.NoError(t, err)
	assert.Equal(t, originalID, parsedID)
}
