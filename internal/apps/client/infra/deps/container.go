package deps

import (
	"context"
	"crypto/x509"
	"errors"
	"fmt"
	"github.com/dlomanov/gophkeeper/internal/apps/client/config"
	"github.com/dlomanov/gophkeeper/internal/apps/client/usecases"
	pb "github.com/dlomanov/gophkeeper/internal/apps/shared/proto"
	"github.com/patrickmn/go-cache"
	"go.uber.org/multierr"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type Container struct {
	Logger  *zap.Logger
	Config  *config.Config
	Conn    *grpc.ClientConn
	UserUC  *usecases.UserUC
	Cache   *cache.Cache
	EntryUC *usecases.EntryUC
}

func NewContainer(
	ctx context.Context,
	logger *zap.Logger,
	config *config.Config,
) (*Container, error) {
	conn, err := createGRPCConn(ctx, config)
	if err != nil {
		return nil, err
	}

	// services
	userClient := pb.NewUserServiceClient(conn)
	entryClient := pb.NewEntryServiceClient(conn)
	cch := cache.New(cache.NoExpiration, cache.NoExpiration)

	// use-cases
	userUC := usecases.NewUserUC(logger, cch, userClient)
	entryUC := usecases.NewEntriesUC(logger, cch, entryClient)

	return &Container{
		Logger:  logger,
		Config:  config,
		Conn:    conn,
		Cache:   cch,
		UserUC:  userUC,
		EntryUC: entryUC,
	}, nil
}

func (c *Container) Close() (merr error) {
	if err := c.Conn.Close(); err != nil {
		merr = multierr.Append(merr, fmt.Errorf("container: failed to close GRPC-connection: %w", err))
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
