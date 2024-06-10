package usecases

import (
	"context"
	"fmt"
	"github.com/dlomanov/gophkeeper/internal/apps/server/entities"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

var (
	emptyToken = entities.Token("")
)

type (
	UserUC struct {
		logger   *zap.Logger
		userRepo UserRepo
		pass     PassHasher
		tokener  Tokener
	}
	UserRepo interface {
		Get(ctx context.Context, login entities.Login) (entities.User, error)
		Exists(ctx context.Context, login entities.Login) (bool, error)
		Create(ctx context.Context, user entities.User) error
	}
	PassHasher interface {
		Hash(password entities.Pass) (entities.PassHash, error)
		Compare(password entities.Pass, hash entities.PassHash) bool
	}
	Tokener interface {
		Create(id uuid.UUID) (entities.Token, error)
		GetUserID(token entities.Token) (uuid.UUID, error)
	}
)

func NewUserUC(
	logger *zap.Logger,
	userRepo UserRepo,
	pass PassHasher,
	tokener Tokener,
) *UserUC {
	return &UserUC{
		logger:   logger,
		userRepo: userRepo,
		pass:     pass,
		tokener:  tokener,
	}
}

func (uc *UserUC) SignUp(ctx context.Context, creds entities.Creds) (entities.Token, error) {
	if !creds.Valid() {
		return emptyToken, entities.ErrUserCredsInvalid
	}

	exists, err := uc.userRepo.Exists(ctx, creds.Login)
	switch {
	case err != nil:
		uc.logger.Error("failed to check user existence", zap.Error(err))
		return emptyToken, err
	case exists:
		uc.logger.Debug("user already exists", zap.String("login", string(creds.Login)))
		return emptyToken, entities.ErrUserExists
	}

	passHash, err := uc.pass.Hash(creds.Pass)
	if err != nil {
		uc.logger.Debug("failed to calculate pass hash", zap.Error(err))
		return emptyToken, err
	}
	user, err := entities.NewUser(entities.HashCreds{
		Login:    creds.Login,
		PassHash: passHash,
	})
	if err != nil {
		uc.logger.Error("failed to create user", zap.Error(err))
		return emptyToken, err
	}
	if err := uc.userRepo.Create(ctx, *user); err != nil {
		uc.logger.Error("failed to create user", zap.Error(err))
	}
	token, err := uc.tokener.Create(user.ID)
	if err != nil {
		uc.logger.Debug("failed to create token", zap.Error(err))
		return emptyToken, err
	}
	return token, nil
}

func (uc *UserUC) SignIn(
	ctx context.Context,
	creds entities.Creds,
) (entities.Token, error) {
	if !creds.Valid() {
		return emptyToken, entities.ErrUserCredsInvalid
	}

	user, err := uc.userRepo.Get(ctx, creds.Login)
	if err != nil {
		uc.logger.Debug("failed to get user", zap.Error(err))
		return emptyToken, err
	}
	if !uc.pass.Compare(creds.Pass, user.PassHash) {
		uc.logger.Debug("invalid credentials", zap.Error(err))
		return emptyToken, entities.ErrUserCredsInvalid
	}
	token, err := uc.tokener.Create(user.ID)
	if err != nil {
		uc.logger.Error("failed to create token", zap.Error(err))
		return emptyToken, err
	}

	return token, nil
}

func (uc *UserUC) GetUserID(_ context.Context, token entities.Token) (uuid.UUID, error) {
	userID, err := uc.tokener.GetUserID(token)
	if err != nil {
		uc.logger.Error("failed to get userID from token", zap.Error(err))
		return uuid.Nil, fmt.Errorf("user_usecase: failed to get userID from token: %w", err)
	}
	return userID, nil
}
