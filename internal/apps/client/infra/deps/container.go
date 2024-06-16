package deps

import (
	"context"
	"crypto/x509"
	"errors"
	"fmt"
	_ "github.com/CovenantSQL/go-sqlite3-encrypt"
	trmsqlx "github.com/avito-tech/go-transaction-manager/drivers/sqlx/v2"
	"github.com/avito-tech/go-transaction-manager/trm/v2/manager"
	"github.com/dlomanov/gophkeeper/internal/apps/client/config"
	"github.com/dlomanov/gophkeeper/internal/apps/client/entities"
	"github.com/dlomanov/gophkeeper/internal/apps/client/infra/repo"
	"github.com/dlomanov/gophkeeper/internal/apps/client/infra/services/mem"
	"github.com/dlomanov/gophkeeper/internal/apps/client/infra/services/pass"
	"github.com/dlomanov/gophkeeper/internal/apps/client/migrations"
	"github.com/dlomanov/gophkeeper/internal/apps/client/usecases"
	pb "github.com/dlomanov/gophkeeper/internal/apps/shared/proto"
	"github.com/dlomanov/gophkeeper/internal/core"
	"github.com/dlomanov/gophkeeper/internal/infra/encrypto"
	"github.com/dlomanov/gophkeeper/internal/infra/migrator"
	"github.com/jmoiron/sqlx"
	"go.uber.org/multierr"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"strings"
	"sync/atomic"
	"time"
)

type Container struct {
	registered atomic.Bool
	Logger     *zap.Logger
	Config     *config.Config
	DB         *sqlx.DB
	Conn       *grpc.ClientConn
	Tx         *manager.Manager
	Memcache   *mem.Cache
	Memstorage *mem.Storage
	UserUC     *usecases.UserUC
	EntryUC    *usecases.EntryUC
}

func NewContainer(
	logger *zap.Logger,
	config *config.Config,
) (*Container, error) {
	memcache := mem.NewCache()
	return &Container{
		Logger:     logger,
		Config:     config,
		Memcache:   memcache,
		Memstorage: nil,
		DB:         nil,
		Conn:       nil,
		Tx:         nil,
		UserUC:     nil,
		EntryUC:    nil,
	}, nil
}

func (c *Container) Register(ctx context.Context, password core.Pass) error {
	db, err := c.createDB(password)
	if err != nil {
		return fmt.Errorf("container: failed to create db: %w", err)
	}
	getter := trmsqlx.DefaultCtxGetter
	trm, err := manager.New(trmsqlx.NewDefaultFactory(db))
	if err != nil {
		return fmt.Errorf("container: failed to create transaction manager: %w", err)
	}

	// auth
	kvRepo := repo.NewKVPairRepo(db, getter, trm)
	memstorage := mem.NewStorage(kvRepo)
	if err = memstorage.Load(ctx, c.Memcache); err != nil {
		return fmt.Errorf("container: failed to load memcache: %w", err)
	}
	userAuthUC := usecases.NewUserAuthUC(&pass.Hasher{}, kvRepo, trm)
	hash, err := userAuthUC.Auth(ctx, password)
	if err != nil {
		return fmt.Errorf("container: failed to auth user: %w, %w", entities.ErrUserMasterPassInvalid, err)
	}

	// grpc
	conn, err := createGRPCConn(ctx, c.Config)
	if err != nil {
		return fmt.Errorf("container: failed to create grpc connection: %w", err)
	}

	// repos
	entryRepo := repo.NewEntryRepo(db, getter)
	entrySyncRepo := repo.NewEntrySyncRepo(db, getter)

	// services
	userClient := pb.NewUserServiceClient(conn)
	entryClient := pb.NewEntryServiceClient(conn)
	encrypter, err := encrypto.NewEncrypter(hash)
	if err != nil {
		return fmt.Errorf("container: failed to create encrypter: %w", err)
	}

	// use-cases
	userUC := usecases.NewUserUC(c.Logger, c.Memcache, userClient)
	entryUC := usecases.NewEntriesUC(
		c.Logger,
		entryClient,
		entryRepo,
		entrySyncRepo,
		encrypter,
		c.Memcache,
		trm,
	)

	c.registered.Store(true)
	c.DB = db
	c.Conn = conn
	c.Tx = trm
	c.Memstorage = memstorage
	c.UserUC = userUC
	c.EntryUC = entryUC
	return nil
}

func (c *Container) Close() (merr error) {
	if !c.registered.Load() {
		c.Logger.Error("container: dependencies are not registered when closing")
		return
	}
	timeoutCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := c.Conn.Close(); err != nil {
		merr = multierr.Append(merr, fmt.Errorf("container: failed to close GRPC-connection: %w", err))
	}
	if err := c.Memstorage.Flush(timeoutCtx, c.Memcache); err != nil {
		merr = multierr.Append(merr, fmt.Errorf("container: failed to flush memcache: %w", err))
	}
	if err := c.DB.Close(); err != nil {
		merr = multierr.Append(merr, fmt.Errorf("container: failed to close database: %w", err))
	}

	return merr
}

func createGRPCConn(ctx context.Context, conf *config.Config) (*grpc.ClientConn, error) {
	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(conf.Cert) {
		return nil, errors.New("container: failed to append cert to pool")
	}
	creds := credentials.NewClientTLSFromCert(certPool, "")
	conn, err := grpc.DialContext(ctx, conf.Address, grpc.WithTransportCredentials(creds))
	if err != nil {
		return nil, fmt.Errorf("container: failed to dial: %w", err)
	}
	return conn, nil
}

func (c *Container) createDB(password core.Pass) (*sqlx.DB, error) {
	dsn := c.Config.DSN
	if strings.HasSuffix(dsn, ".db") {
		dsn += "?"
	} else {
		dsn += "&"
	}
	dsn += "_crypto_key=" + string(password)
	db, err := sqlx.Connect("sqlite3", dsn)
	if err != nil {
		return nil, fmt.Errorf("container: failed to open database: %w", err)
	}

	ms, err := migrations.GetMigrations()
	if err != nil {
		return nil, fmt.Errorf("container: failed to get migrations: %w", err)
	}
	if err = migrator.Migrate(c.Logger.Sugar(), db.DB, ms); err != nil {
		return nil, fmt.Errorf("container: failed to up migrations: %w", err)
	}

	return db, nil
}
