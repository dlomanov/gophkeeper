package deps

import (
	"errors"
	"fmt"
	trmsqlx "github.com/avito-tech/go-transaction-manager/drivers/sqlx/v2"
	"github.com/avito-tech/go-transaction-manager/trm/v2/manager"
	"github.com/dlomanov/gophkeeper/internal/apps/server/config"
	"github.com/dlomanov/gophkeeper/internal/apps/server/infra/pass"
	"github.com/dlomanov/gophkeeper/internal/apps/server/infra/repo"
	"github.com/dlomanov/gophkeeper/internal/apps/server/infra/services/diff"
	"github.com/dlomanov/gophkeeper/internal/apps/server/infra/services/token"
	"github.com/dlomanov/gophkeeper/internal/apps/server/migrations"
	"github.com/dlomanov/gophkeeper/internal/apps/server/usecases"
	"github.com/dlomanov/gophkeeper/internal/infra/encrypto"
	"github.com/dlomanov/gophkeeper/internal/infra/migrator"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
	"io"
)

var _ io.Closer = (*Container)(nil)

type Container struct {
	Logger  *zap.Logger
	Config  *config.Config
	DB      *sqlx.DB
	Tx      *manager.Manager
	UserUC  *usecases.UserUC
	EntryUC *usecases.EntryUC
}

func NewContainer(
	logger *zap.Logger,
	config *config.Config,
) (*Container, error) {
	db, err := sqlx.Connect("pgx", config.DatabaseDSN)
	if err != nil {
		return nil, fmt.Errorf("container: failed to connect to database: %w", err)
	}
	if err = upMigrations(logger, db); err != nil {
		return nil, err
	}
	trm, err := manager.New(trmsqlx.NewDefaultFactory(db))
	if err != nil {
		return nil, fmt.Errorf("container: failed to create transaction manager: %w", err)
	}

	// repos
	getter := trmsqlx.DefaultCtxGetter
	userRepo := repo.NewUserRepo(db, getter)

	// services
	hasher := pass.NewHasher(config.PassHashCost)
	tokener := token.NewJWT(config.TokenSecretKey, config.TokenExpires)
	merger := diff.NewEntry()
	encrypter, err := encrypto.NewEncrypter(config.DataSecretKey)
	if err != nil {
		return nil, fmt.Errorf("container: failed to create encrypter: %w", err)
	}

	// usecases
	userUC := usecases.NewUserUC(logger, userRepo, hasher, tokener)
	entryUC := usecases.NewEntryUC(
		logger,
		repo.NewEntryRepo(db, getter),
		merger,
		encrypter,
		trm)

	return &Container{
		Logger:  logger,
		Config:  config,
		DB:      db,
		Tx:      trm,
		UserUC:  userUC,
		EntryUC: entryUC,
	}, nil
}

func (c Container) Close() error {
	var errs []error
	if err := c.DB.Close(); err != nil {
		errs = append(errs, err)
	}
	return errors.Join(errs...)
}

func upMigrations(logger *zap.Logger, db *sqlx.DB) error {
	ms, err := migrations.GetMigrations()
	if err != nil {
		return fmt.Errorf("container: failed to get migrations: %w", err)
	}
	if err = migrator.Migrate(logger.Sugar(), db.DB, ms); err != nil {
		return fmt.Errorf("container: failed to up migrations: %w", err)
	}
	return nil
}
