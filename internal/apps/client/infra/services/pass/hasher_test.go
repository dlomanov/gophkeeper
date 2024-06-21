package pass_test

import (
	"github.com/dlomanov/gophkeeper/internal/apps/client/infra/services/pass"
	"github.com/dlomanov/gophkeeper/internal/core"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestHasher_Hash(t *testing.T) {
	tests := []struct {
		name           string
		password       core.Pass
		salt           core.Salt
		wantHashBase64 string
	}{
		{
			name:           "correct",
			password:       core.Pass("password"),
			salt:           core.Salt("salt"),
			wantHashBase64: "S8D9UH6TpgB2gCE0HscmxXwAy1WkcCoWUBMTZVAM9HE=",
		},
		{
			name:           "correct",
			password:       core.Pass(""),
			salt:           core.Salt(""),
			wantHashBase64: "LIBL9dDMsoBR8g+wuZv36aRaZ0lmUFNgGN2IOdk3B0g=",
		},
		{
			name:           "correct",
			password:       core.Pass("password1"),
			salt:           core.Salt("salt"),
			wantHashBase64: "wETsopVSFxkjlLYAKpQ/Xbqb2dWqYtDRiqZBObKicDU=",
		},
	}

	h := pass.Hasher{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := h.Hash(tt.password, tt.salt)
			require.NoError(t, err)

			wantHash, err := core.NewPassHash(tt.wantHashBase64)
			require.NoError(t, err, "failed to create want hash")
			require.Equal(t, wantHash, got, "hashes should be equal, want %s, got %s", tt.wantHashBase64, got.Base64String())
		})
	}
}

func TestHasher_Compare(t *testing.T) {
	tests := []struct {
		name       string
		password   core.Pass
		salt       core.Salt
		hashBase64 string
		want       bool
	}{
		{
			name:       "correct",
			password:   core.Pass("password"),
			salt:       core.Salt("salt"),
			hashBase64: "S8D9UH6TpgB2gCE0HscmxXwAy1WkcCoWUBMTZVAM9HE=",
			want:       true,
		},
		{
			name:       "correct",
			password:   core.Pass("password"),
			salt:       core.Salt("salt"),
			hashBase64: "LIBL9dDMsoBR8g+wuZv36aRaZ0lmUFNgGN2IOdk3B0g=",
			want:       false,
		},
	}

	h := pass.Hasher{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash, err := core.NewPassHash(tt.hashBase64)
			require.NoError(t, err, "failed to create hash")
			got := h.Compare(tt.password, tt.salt, hash)
			require.Equal(t, tt.want, got, "hashes should be equal, want %t, got %t", tt.want, got)
		})
	}
}
