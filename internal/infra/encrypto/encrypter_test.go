package encrypto_test

import (
	"github.com/dlomanov/gophkeeper/internal/infra/encrypto"
	"testing"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name    string
		key     []byte
		wantErr bool
	}{
		{
			name:    "1",
			key:     nil,
			wantErr: true,
		},
		{
			name:    "2",
			key:     []byte{0x00},
			wantErr: true,
		},
		{
			name:    "3",
			key:     make([]byte, 16),
			wantErr: false,
		},
		{
			name:    "4",
			key:     make([]byte, 24),
			wantErr: false,
		},
		{
			name:    "5",
			key:     make([]byte, 32),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := encrypto.NewEncrypter(tt.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("Encrypt() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}
