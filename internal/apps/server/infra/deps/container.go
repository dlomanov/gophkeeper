package deps

import (
	"errors"
	trmsqlx "github.com/avito-tech/go-transaction-manager/drivers/sqlx/v2"
	"github.com/avito-tech/go-transaction-manager/trm/v2/manager"
	"github.com/dlomanov/gophkeeper/internal/apps/server/config"
	"github.com/dlomanov/gophkeeper/internal/apps/server/infra/repo"
	"github.com/dlomanov/gophkeeper/internal/apps/server/infra/services/pass"
	"github.com/dlomanov/gophkeeper/internal/apps/server/infra/services/token"
	"github.com/dlomanov/gophkeeper/internal/apps/server/usecases"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
	"io"
)

var _ io.Closer = (*Container)(nil)

type Container struct {
	Logger *zap.Logger
	Config *config.Config
	DB     *sqlx.DB
	Tx     *manager.Manager
	UserUC *usecases.UserUC
}

func NewContainer(
	logger *zap.Logger,
	config *config.Config,
) (*Container, error) {
	db, err := sqlx.Connect("pgx", config.DatabaseDSN)
	if err != nil {
		logger.Error("failed to connect to database", zap.Error(err))
		return nil, err
	}

	trm, err := manager.New(trmsqlx.NewDefaultFactory(db))
	if err != nil {
		return nil, err
	}

	// repos
	getter := trmsqlx.DefaultCtxGetter
	userRepo := repo.NewUserRepo(db, getter)

	// services
	hasher := pass.NewHasher(config.PassHashCost)
	tokener := token.NewJWT([]byte(config.TokenSecretKey), config.TokenExpires)

	// usecases
	userUC := usecases.NewUserUC(logger, userRepo, hasher, tokener)

	return &Container{
		Logger: logger,
		Config: config,
		DB:     db,
		Tx:     trm,
		UserUC: userUC,
	}, nil
}

func (c Container) Close() error {
	var errs []error
	if err := c.DB.Close(); err != nil {
		errs = append(errs, err)
	}
	return errors.Join(errs...)
}
