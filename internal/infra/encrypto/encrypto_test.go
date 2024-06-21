package encrypto_test

import (
	"crypto/rand"
	"github.com/dlomanov/gophkeeper/internal/infra/encrypto"
	"github.com/stretchr/testify/require"
	"io"
	"testing"
)

func TestCrypto(t *testing.T) {
	key := make([]byte, 32)
	_, err := io.ReadFull(rand.Reader, key)
	require.NoError(t, err, "no error expected")

	data := []byte("testdata")
	encrypted, err := encrypto.Encrypt(key, data)
	require.NoError(t, err, "no error expected")

	decrypted, err := encrypto.Decrypt(key, encrypted)
	require.NoError(t, err, "no error expected")

	require.Equal(t, data, decrypted, "data should be equal")
}

func TestEncrypt_validation(t *testing.T) {
	tests := []struct {
		name    string
		key     []byte
		data    []byte
		wantErr bool
	}{
		{
			name:    "key and data are nil",
			key:     nil,
			data:    nil,
			wantErr: true,
		},
		{
			name:    "key with invalid length and data is nil",
			key:     []byte{0x00},
			data:    nil,
			wantErr: true,
		},
		{
			name:    "key with valid length and data is nil",
			key:     make([]byte, 16),
			data:    nil,
			wantErr: true,
		},
		{
			name:    "key and data are valid",
			key:     make([]byte, 16),
			data:    make([]byte, 1),
			wantErr: false,
		},
		{
			name:    "key and data are valid",
			key:     make([]byte, 24),
			data:    make([]byte, 1),
			wantErr: false,
		},
		{
			name:    "key and data are valid",
			key:     make([]byte, 32),
			data:    make([]byte, 1),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := encrypto.Encrypt(tt.key, tt.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("Encrypt() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestDecrypt_validation(t *testing.T) {
	tests := []struct {
		name    string
		key     []byte
		data    []byte
		wantErr bool
	}{
		{
			name:    "key and data are nil",
			key:     nil,
			data:    nil,
			wantErr: true,
		},
		{
			name:    "key with invalid length and data is nil",
			key:     []byte{0x00},
			data:    nil,
			wantErr: true,
		},
		{
			name:    "key with valid length and data is nil",
			key:     make([]byte, 16),
			data:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := encrypto.Decrypt(tt.key, tt.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("Encrypt() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}
