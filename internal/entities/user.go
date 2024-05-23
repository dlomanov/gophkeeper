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

func NewUser(creds HashCreds) User {

	now := time.Now().UTC()
	return User{
		ID:        uuid.New(),
		HashCreds: creds,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func (c Creds) Valid() bool {
	return c.Login != "" && c.Pass != ""
}

func (t Token) Valid() bool {
	return t != ""
}
