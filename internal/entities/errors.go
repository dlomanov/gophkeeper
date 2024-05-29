package entities

import "github.com/dlomanov/gophkeeper/internal/core/apperrors"

var (
	ErrEntryKeyInvalid       = apperrors.NewInvalid("invalid entry key")
	ErrEntryIsNil            = apperrors.NewInvalid("entry is nil")
	ErrEntryIDInvalid        = apperrors.NewInvalid("invalid entry ID")
	ErrEntryTypeInvalid      = apperrors.NewInvalid("invalid entry type")
	ErrEntryDataEmpty        = apperrors.NewInvalid("empty entry data")
	ErrEntryDataSizeExceeded = apperrors.NewInvalid("entry data size exceeded")
	ErrEntryExists           = apperrors.NewInvalid("entry already exists")
	ErrEntryNotFound         = apperrors.NewNotFound("entry not found")
	ErrUserIDInvalid         = apperrors.NewInvalid("user ID is invalid")
	ErrUserExists            = apperrors.NewInvalid("user already exists")
	ErrUserNotFound          = apperrors.NewNotFound("user not found")
	ErrUserCredsInvalid      = apperrors.NewInvalid("user credentials are invalid")
	ErrUserTokenInvalid      = apperrors.NewInvalid("invalid token")
	ErrUserTokenExpired      = apperrors.NewInvalid("token expired")
)
