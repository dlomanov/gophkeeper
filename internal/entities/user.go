package entities

import (
	"time"

	"github.com/google/uuid"
)

type (
	Creds struct {
		Login Login
		Pass  Pass
	}
	HashCreds struct {
		Login    Login
		PassHash PassHash
	}
	Login     string
	Pass      string
	PassHash  string
	UserToken struct {
		UserID uuid.UUID
		Token  Token
	}
	User struct {
		ID uuid.UUID
		HashCreds
		CreatedAt time.Time
		UpdatedAt time.Time
	}
	Token string
)

func NewUser(creds HashCreds) (*User, error) {
	if !creds.Valid() {
		return nil, ErrUserCredsInvalid
	}

	now := time.Now().UTC()
	return &User{
		ID:        uuid.New(),
		HashCreds: creds,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

func (c Creds) Valid() bool {
	return c.Login != "" && c.Pass != ""
}

func (c HashCreds) Valid() bool {
	return c.Login != "" && c.PassHash != ""
}

func (t Token) Valid() bool {
	return t != ""
}
