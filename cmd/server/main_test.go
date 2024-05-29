package main_test

import (
	"context"
	"github.com/dlomanov/gophkeeper/cmd/server/config"
	"github.com/dlomanov/gophkeeper/internal/apps/server"
	sharedmd "github.com/dlomanov/gophkeeper/internal/apps/shared/md"
	pb "github.com/dlomanov/gophkeeper/internal/apps/shared/proto"
	"github.com/dlomanov/gophkeeper/internal/infra/grpcserver"
	"github.com/dlomanov/gophkeeper/internal/infra/pg/testcont"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
	"net"
	"testing"
	"time"
)

const (
	startupDelay    = 3 * time.Second
	testTimeout     = 15 * time.Second
	teardownTimeout = 10 * time.Second
	bufferSize      = 1024 * 1024
)

type AppSuite struct {
	suite.Suite
	logger          *zap.Logger
	teardownCtx     context.Context
	teardown        context.CancelFunc
	listener        *bufconn.Listener
	pgc             *postgres.PostgresContainer
	serverStoppedCh chan error
}

func TestAppSuite(t *testing.T) {
	suite.Run(t, new(AppSuite))
}

func (s *AppSuite) SetupSuite() {
	var err error

	s.logger = zaptest.NewLogger(s.T(), zaptest.Level(zap.DebugLevel))
	s.teardownCtx, s.teardown = context.WithCancel(context.Background())
	s.listener = bufconn.Listen(bufferSize)

	c := config.Parse()
	s.pgc, c.DatabaseDSN, err = testcont.RunPostgres(s.teardownCtx, c.DatabaseDSN)
	require.NoError(s.T(), err, "failed to run postgres container")

	s.serverStoppedCh = make(chan error, 1)
	go func() {
		ctx := context.WithValue(s.teardownCtx, grpcserver.ListenerKey, s.listener)
		s.serverStoppedCh <- server.Run(ctx, c)
	}()
	time.Sleep(startupDelay)
}

func (s *AppSuite) TearDownSuite() {
	timeout, cancel := context.WithTimeout(context.Background(), teardownTimeout)
	defer cancel()

	s.teardown()

	select {
	case err := <-s.serverStoppedCh:
		assert.NoError(s.T(), err, "server stopped with error")
	case <-timeout.Done():
		s.logger.Error("teardown timeout")
	}
	if err := s.pgc.Terminate(timeout); err != nil {
		s.logger.Error("failed to terminate postgres container", zap.Error(err))
	}
	if err := s.listener.Close(); err != nil {
		s.logger.Error("failed to close listener", zap.Error(err))
	}
}

func (s *AppSuite) TestAuth() {
	ctx, cancel := context.WithTimeout(s.teardownCtx, testTimeout)
	defer cancel()
	conn := s.createGRPCConn()
	defer func(conn *grpc.ClientConn) {
		if err := conn.Close(); err != nil {
			s.logger.Error("failed to close GRPC-connection", zap.Error(err))
		}
	}(conn)
	userService := pb.NewUserServiceClient(conn)

	var err error

	_, err = userService.SignIn(ctx, &pb.SignInUserRequest{
		Login:    "",
		Password: "",
	})
	require.Error(s.T(), err, "expected error")
	require.Equal(s.T(), codes.InvalidArgument, status.Code(err), "expected invalid argument error")

	_, err = userService.SignIn(ctx, &pb.SignInUserRequest{
		Login:    "testuser",
		Password: "testpassword",
	})
	require.Error(s.T(), err, "expected error")
	require.Equal(s.T(), codes.Unauthenticated, status.Code(err), "expected unauthenticated error")

	_, err = userService.SignUp(ctx, &pb.SignUpUserRequest{
		Login:    "",
		Password: "",
	})
	require.Error(s.T(), err, "expected error")
	require.Equal(s.T(), codes.InvalidArgument, status.Code(err), "expected invalid argument error")

	signUpResp, err := userService.SignUp(ctx, &pb.SignUpUserRequest{
		Login:    "testuser",
		Password: "testpassword",
	})
	require.NoError(s.T(), err, "no error expected")
	require.NotNil(s.T(), signUpResp, "expected response not nil")
	require.NotEmpty(s.T(), signUpResp.Token, "expected token not empty")

	signInResp, err := userService.SignIn(ctx, &pb.SignInUserRequest{
		Login:    "testuser",
		Password: "testpassword",
	})
	require.NoError(s.T(), err, "no error expected")
	require.NotNil(s.T(), signInResp, "expected response not nil")
	require.NotEmpty(s.T(), signInResp.Token, "expected token not empty")

	entryService := pb.NewEntryServiceClient(conn)
	ctx = metadata.AppendToOutgoingContext(ctx, sharedmd.NewTokenKV(signInResp.Token)...)
	_, err = entryService.Create(ctx, &pb.CreateEntryRequest{})
	require.NoError(s.T(), err, "expected no error while creating entry")
}

func (s *AppSuite) createGRPCConn() *grpc.ClientConn {
	conn, err := grpc.DialContext(s.teardownCtx, "bufnet",
		grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
			return s.listener.Dial()
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(s.T(), err)
	return conn
}
