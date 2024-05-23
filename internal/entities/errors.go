package entities

import "github.com/dlomanov/gophkeeper/internal/core/apperrors"

var (
	ErrEntryTypeInvalid = apperrors.NewInvalid("invalid entry type")
	ErrEntryDataInvalid = apperrors.NewInvalid("invalid entry data")
	ErrUserExists       = apperrors.NewInvalid("user already exists")
	ErrUserNotFound     = apperrors.NewNotFound("user not found")
	ErrUserCredsInvalid = apperrors.NewInvalid("user credentials are invalid")
	ErrUserTokenInvalid = apperrors.NewInvalid("invalid token")
	ErrUserTokenExpired = apperrors.NewInvalid("token expired")
)
