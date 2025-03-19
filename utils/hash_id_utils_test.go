package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEncryptOrderID(t *testing.T) {
	id := uint64(123456)
	encrypted, err := EncryptOrderID(id)
	if err != nil {
		t.Errorf("EncryptOrderID failed: %v", err)
	}
	if encrypted == "" {
		t.Errorf("EncryptOrderID returned empty string")
	}
}

func TestDecryptOrderID(t *testing.T) {
	id := uint64(123456)
	encrypted, err := EncryptOrderID(id)
	if err != nil {
		t.Errorf("EncryptOrderID failed: %v", err)
	}
	decrypted, err := DecryptOrderID(encrypted)
	if err != nil {
		t.Errorf("DecryptOrderID failed: %v", err)
	}
	if int64(id) != decrypted {
		t.Errorf("DecryptOrderID returned incorrect ID: got %d, want %d", decrypted, id)
	}
}

func TestDecryptOrderIDIsRight(t *testing.T) {
	id := uint64(123456)

	encrypted, _ := EncryptOrderID(id)

	decrypted, _ := DecryptOrderID(encrypted)
	assert.Equal(t, int64(123456), decrypted)
}
