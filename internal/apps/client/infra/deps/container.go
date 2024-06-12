package deps

import (
	"context"
	"crypto/x509"
	"errors"
	"fmt"
	trmsqlx "github.com/avito-tech/go-transaction-manager/drivers/sqlx/v2"
	"github.com/avito-tech/go-transaction-manager/trm/v2/manager"
	"github.com/dlomanov/gophkeeper/internal/apps/client/config"
	"github.com/dlomanov/gophkeeper/internal/apps/client/infra/repo"
	"github.com/dlomanov/gophkeeper/internal/apps/client/infra/services/mem"
	"github.com/dlomanov/gophkeeper/internal/apps/client/usecases"
	pb "github.com/dlomanov/gophkeeper/internal/apps/shared/proto"
	"github.com/dlomanov/gophkeeper/internal/infra/encrypto"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"go.uber.org/multierr"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"time"
)

type Container struct {
	Logger   *zap.Logger
	Config   *config.Config
	DB       *sqlx.DB
	Conn     *grpc.ClientConn
	Tx       *manager.Manager
	Memcache *mem.Cache
	UserUC   *usecases.UserUC
	EntryUC  *usecases.EntryUC
}

func NewContainer(
	ctx context.Context,
	logger *zap.Logger,
	config *config.Config,
) (*Container, error) {
	db, err := sqlx.Connect("sqlite3", config.DSN)
	if err != nil {
		return nil, fmt.Errorf("container: failed to open database: %w", err)
	}

	conn, err := createGRPCConn(ctx, config)
	if err != nil {
		return nil, err
	}

	trm, err := manager.New(trmsqlx.NewDefaultFactory(db))
	if err != nil {
		return nil, err
	}

	// repos
	getter := trmsqlx.DefaultCtxGetter
	kvRepo := repo.NewKVPairRepo(db)
	entryRepo := repo.NewEntryRepo(db, getter)
	entrySyncRepo := repo.NewEntrySyncRepo(db, getter)

	// services
	userClient := pb.NewUserServiceClient(conn)
	entryClient := pb.NewEntryServiceClient(conn)
	memcache := mem.NewCache(kvRepo)
	encrypter, err := encrypto.NewEncrypter([]byte("1234567890123456")) // TODO: it should be master-key
	if err != nil {
		return nil, fmt.Errorf("container: failed to create encrypter: %w", err)
	}

	// use-cases
	userUC := usecases.NewUserUC(logger, memcache, userClient)
	entryUC := usecases.NewEntriesUC(
		logger,
		entryClient,
		entryRepo,
		entrySyncRepo,
		encrypter,
		memcache,
		trm,
	)

	return &Container{
		Logger:   logger,
		Config:   config,
		DB:       db,
		Conn:     conn,
		Tx:       trm,
		Memcache: memcache,
		UserUC:   userUC,
		EntryUC:  entryUC,
	}, nil
}

func (c *Container) Close() (merr error) {
	timeoutCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := c.Conn.Close(); err != nil {
		merr = multierr.Append(merr, fmt.Errorf("container: failed to close GRPC-connection: %w", err))
	}
	if err := c.Memcache.Flush(timeoutCtx); err != nil {
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
