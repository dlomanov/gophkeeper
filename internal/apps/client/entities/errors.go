package entities

import "github.com/dlomanov/gophkeeper/internal/core/apperrors"

var (
	ErrEntryKeyInvalid       = apperrors.NewInvalid("invalid entry key")
	ErrEntryInvalid          = apperrors.NewInvalid("invalid entry")
	ErrEntryIDInvalid        = apperrors.NewInvalid("invalid entry ID")
	ErrEntryTypeInvalid      = apperrors.NewInvalid("invalid entry type")
	ErrEntryDataEmpty        = apperrors.NewInvalid("empty entry data")
	ErrEntryDataSizeExceeded = apperrors.NewInvalid("entry data size exceeded")
	ErrEntryExists           = apperrors.NewInvalid("entry already exists")
	ErrEntryNotFound         = apperrors.NewNotFound("entry not found")
	ErrUserExists            = apperrors.NewInvalid("user already exists")
	ErrUserCredsInvalid      = apperrors.NewInvalid("user credentials are invalid")
	ErrUserLoginInvalid      = apperrors.NewInvalid("user login is invalid")
	ErrUserPasswordInvalid   = apperrors.NewInvalid("user password is invalid")
	ErrUserTokenNotFound     = apperrors.NewNotFound("token not found")
	ErrUserTokenInvalid      = apperrors.NewInvalid("token is invalid")
	ErrUserMasterPassInvalid = apperrors.NewInvalid("master password is invalid")
	ErrKVPairNotFound        = apperrors.NewNotFound("key-value pair not found")
	ErrServerUnavailable     = apperrors.NewInternal("server unavailable")
	ErrEntryDataTypeInvalid  = apperrors.NewInvalid("invalid entry data type")
)
