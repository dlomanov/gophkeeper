package entities_test

import (
	"github.com/dlomanov/gophkeeper/internal/entities"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewUser(t *testing.T) {
	creds := entities.HashCreds{
		Login:    "",
		PassHash: "hashedPassword",
	}
	_, err := entities.NewUser(creds)
	require.ErrorIs(t, err, entities.ErrUserCredsInvalid, "error mismatch")

	creds.Login = "testUser"
	user, err := entities.NewUser(creds)
	require.NoError(t, err, "error should be nil")
	require.NotNil(t, user, "user should not be nil")
	require.Equal(t, creds, user.HashCreds, "hash creds mismatch")
	require.NotZero(t, user.ID, "user ID should not be zero")
}

func TestTokenValid(t *testing.T) {
	tests := []struct {
		name  string
		token entities.Token
		want  bool
	}{
		{
			name:  "valid token",
			token: entities.Token("validToken"),
			want:  true,
		},
		{
			name:  "invalid token",
			token: entities.Token(""),
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, tt.token.Valid(), "token validity mismatch")
		})
	}
}

func TestCredsValid(t *testing.T) {
	tests := []struct {
		name  string
		creds entities.Creds
		want  bool
	}{
		{
			name: "valid creds",
			creds: entities.Creds{
				Login: "testUser",
				Pass:  "password",
			},
			want: true,
		},
		{
			name: "invalid creds",
			creds: entities.Creds{
				Login: "",
				Pass:  "password",
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, tt.creds.Valid(), "creds validity mismatch")
		})
	}
}
