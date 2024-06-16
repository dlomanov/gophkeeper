package entities

import (
	"github.com/dlomanov/gophkeeper/internal/core"
	"time"

	"github.com/google/uuid"
)

type (
	Creds struct {
		Login Login
		Pass  core.Pass
	}
	HashCreds struct {
		Login    Login
		PassHash core.PassHash
	}
	Login     string
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
	return len(c.Login) != 0 && len(c.Pass) != 0
}

func (c HashCreds) Valid() bool {
	return len(c.Login) != 0 && len(c.PassHash) != 0
}

func (t Token) Valid() bool {
	return t != ""
}
