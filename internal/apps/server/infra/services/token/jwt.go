package token

import (
	"fmt"
	"github.com/dlomanov/gophkeeper/internal/apps/server/entities"
	"github.com/dlomanov/gophkeeper/internal/apps/server/usecases"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"time"
)

var (
	_ usecases.Tokener = (*JWTTokener)(nil)

	method = jwt.SigningMethodHS256
)

type (
	JWTTokener struct {
		secret  []byte
		expires time.Duration
	}
	Claims struct {
		jwt.RegisteredClaims
		UserID string `json:"user_id"`
	}
)

func NewJWT(secret []byte, expires time.Duration) JWTTokener {
	return JWTTokener{
		secret:  secret,
		expires: expires,
	}
}

func (t JWTTokener) Create(id uuid.UUID) (entities.Token, error) {
	token := jwt.NewWithClaims(method, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(t.expires)),
		},
		UserID: id.String(),
	})
	tokenString, err := token.SignedString(t.secret)
	if err != nil {
		return "", err
	}
	return entities.Token(tokenString), nil
}

func (t JWTTokener) GetUserID(token entities.Token) (uuid.UUID, error) {
	c := new(Claims)

	value, err := jwt.ParseWithClaims(string(token), c, func(token *jwt.Token) (any, error) {
		if m, ok := token.Method.(*jwt.SigningMethodHMAC); !ok || m.Name != method.Name {
			return nil, fmt.Errorf("%w unexpected signing method: %v", entities.ErrUserTokenInvalid, token.Header["alg"])
		}
		return t.secret, nil
	})
	if err != nil {
		return uuid.Nil, err
	}
	if !value.Valid {
		return uuid.Nil, entities.ErrUserTokenInvalid
	}

	expires := c.ExpiresAt.UTC()
	now := time.Now().UTC()
	if expires.Compare(now) == -1 {
		return uuid.Nil, entities.ErrUserTokenExpired
	}

	id, err := uuid.Parse(c.UserID)
	if err != nil {
		return uuid.Nil, err
	}

	return id, nil
}
