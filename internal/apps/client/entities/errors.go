package entities

import "github.com/dlomanov/gophkeeper/internal/core/apperrors"

var (
	ErrEntryKeyInvalid       = apperrors.NewInvalid("invalid entry key")
	ErrEntryInvalid          = apperrors.NewInvalid("invalid entry")
	ErrEntryTypeInvalid      = apperrors.NewInvalid("invalid entry type")
	ErrEntryDataEmpty        = apperrors.NewInvalid("empty entry data")
	ErrEntryDataSizeExceeded = apperrors.NewInvalid("entry data size exceeded")
	ErrEntryExists           = apperrors.NewInvalid("entry already exists")
	ErrEntryNotFound         = apperrors.NewNotFound("entry not found")
	ErrUserExists            = apperrors.NewInvalid("user already exists")
	ErrUserCredsInvalid      = apperrors.NewInvalid("user credentials are invalid")
	ErrUserTokenNotFound     = apperrors.NewNotFound("token not found")
	ErrServerUnavailable     = apperrors.NewInternal("server unavailable")
)
