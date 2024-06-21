package usecases

import (
	"context"
	"errors"
	"fmt"
	"github.com/avito-tech/go-transaction-manager/trm/v2/manager"
	"github.com/dlomanov/gophkeeper/internal/apps/client/entities"
	"github.com/dlomanov/gophkeeper/internal/core"
)

const (
	storageKeyUserPassHash = "user_pass_hash"
	storageKeyUserSalt     = "user_salt"
)

type (
	UserAuthUC struct {
		hasher  Hasher
		storage Storage
		tx      *manager.Manager
	}
	Hasher interface {
		Hash(password core.Pass, salt core.Salt) (core.PassHash, error)
		Compare(pass core.Pass, salt core.Salt, hash core.PassHash) bool
		GenerateSalt() (core.Salt, error)
	}
	Storage interface {
		Get(ctx context.Context, key string) (string, error)
		Set(ctx context.Context, key string, value string) error
	}
)

func NewUserAuthUC(
	hasher Hasher,
	storage Storage,
	tx *manager.Manager,
) *UserAuthUC {
	return &UserAuthUC{
		hasher:  hasher,
		storage: storage,
		tx:      tx,
	}
}

func (uc *UserAuthUC) Auth(ctx context.Context, pass core.Pass) (hash core.PassHash, err error) {
	hashBase64, err := uc.storage.Get(ctx, storageKeyUserPassHash)
	switch {
	case errors.Is(err, entities.ErrKVPairNotFound):
		return uc.register(ctx, pass)
	case err != nil:
		return nil, fmt.Errorf("user_pass: failed to get hash: %w", err)
	}
	hash, err = core.NewPassHash(hashBase64)
	if err != nil {
		return nil, fmt.Errorf("user_pass: failed to create hash from base64: %w", err)
	}

	saltBase64, err := uc.storage.Get(ctx, storageKeyUserSalt)
	switch {
	case errors.Is(err, entities.ErrKVPairNotFound):
		return nil, fmt.Errorf("user_pass: salt not found: %w", err)
	case err != nil:
		return nil, fmt.Errorf("user_pass: failed to get salt: %w", err)
	}
	salt, err := core.NewSalt(saltBase64)
	if err != nil {
		return nil, fmt.Errorf("user_pass: failed to create salt from base64: %w", err)
	}

	if !uc.hasher.Compare(pass, salt, hash) {
		return nil, entities.ErrUserMasterPassInvalid
	}
	return hash, nil
}

func (uc *UserAuthUC) register(ctx context.Context, pass core.Pass) (core.PassHash, error) {
	salt, err := uc.hasher.GenerateSalt()
	if err != nil {
		return nil, fmt.Errorf("user_pass: failed to generate salt: %w", err)
	}
	hash, err := uc.hasher.Hash(pass, salt)
	if err != nil {
		return nil, fmt.Errorf("user_pass: failed to hash pass: %w", err)
	}

	if err = uc.tx.Do(ctx, func(ctx context.Context) error {
		if err := uc.storage.Set(ctx, storageKeyUserPassHash, hash.Base64String()); err != nil {
			return fmt.Errorf("user_pass: failed to set hash: %w", err)
		}
		if err := uc.storage.Set(ctx, storageKeyUserSalt, salt.Base64String()); err != nil {
			return fmt.Errorf("user_pass: failed to set salt: %w", err)
		}
		return nil
	}); err != nil {
		return nil, err
	}

	return hash, nil
}
