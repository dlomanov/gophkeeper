package main_test

import (
	"context"
	"github.com/dlomanov/gophkeeper/cmd/server/config"
	"github.com/dlomanov/gophkeeper/internal/apps/server"
	sharedmd "github.com/dlomanov/gophkeeper/internal/apps/shared/md"
	pb "github.com/dlomanov/gophkeeper/internal/apps/shared/proto"
	"github.com/dlomanov/gophkeeper/internal/infra/grpcserver"
	"github.com/dlomanov/gophkeeper/internal/infra/pg/testcont"
	"github.com/google/uuid"
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
	"reflect"
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

	s.logger = zaptest.NewLogger(s.T(), zaptest.Level(zap.InfoLevel))
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

	// 1 Auth
	_, err = userService.SignIn(ctx, &pb.SignInUserRequest{
		Login:    "",
		Password: "",
	})
	require.Error(s.T(), err, "expected error")
	require.Equal(s.T(), codes.InvalidArgument, status.Code(err), "expected invalid argument code")

	_, err = userService.SignIn(ctx, &pb.SignInUserRequest{
		Login:    "testuser",
		Password: "testpassword",
	})
	require.Error(s.T(), err, "expected error")
	require.Equal(s.T(), codes.Unauthenticated, status.Code(err), "expected unauthenticated code")

	_, err = userService.SignUp(ctx, &pb.SignUpUserRequest{
		Login:    "",
		Password: "",
	})
	require.Error(s.T(), err, "expected error")
	require.Equal(s.T(), codes.InvalidArgument, status.Code(err), "expected invalid argument code")

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

	// 2 Entries
	entryService := pb.NewEntryServiceClient(conn)
	_, err = entryService.GetAll(ctx, &pb.GetAllEntriesRequest{})
	require.Error(s.T(), err, "expected error")
	require.Equal(s.T(), codes.Unauthenticated, status.Code(err), "expected unauthenticated code")

	ctx = metadata.AppendToOutgoingContext(ctx, sharedmd.NewTokenKV(signInResp.Token)...)
	getAll, err := entryService.GetAll(ctx, &pb.GetAllEntriesRequest{})
	require.NoError(s.T(), err, "no error expected")
	require.NotNil(s.T(), getAll, "expected response not nil")
	require.Empty(s.T(), getAll.Entries, "expected empty entries response")

	// 2.1 Entries: create and get
	createEntries := make([]*pb.CreateEntryRequest, 3)
	createEntries[0] = &pb.CreateEntryRequest{
		Key:  "key1",
		Type: pb.EntryType_ENTRY_TYPE_NOTE,
		Meta: map[string]string{"description": "test_note_1"},
		Data: []byte("test_data_1"),
	}
	createEntries[1] = &pb.CreateEntryRequest{
		Key:  "key2",
		Type: pb.EntryType_ENTRY_TYPE_BINARY,
		Meta: map[string]string{"description": "test_binary_2"},
		Data: []byte("test_data_2"),
	}
	createEntries[2] = &pb.CreateEntryRequest{
		Key:  "key3",
		Type: pb.EntryType_ENTRY_TYPE_PASSWORD,
		Meta: map[string]string{"description": "test_password_3"},
		Data: []byte("test_data_3"),
	}
	for _, entry := range createEntries {
		created, err := entryService.Create(ctx, entry)
		require.NoError(s.T(), err, "no error expected")
		assert.NotEqual(s.T(), created.Id, uuid.Nil.String(), "expected not empty ID")
		assert.NotEmpty(s.T(), created.CreatedAt, "expected not empty created_at")
		assert.NotEmpty(s.T(), created.UpdatedAt, "expected not empty updated_at")
		assert.Equal(s.T(), created.CreatedAt, created.UpdatedAt, "expected created_at == updated_at")
	}
	getAll, err = entryService.GetAll(ctx, &pb.GetAllEntriesRequest{})
	require.NoError(s.T(), err, "no error expected")
	require.NotNil(s.T(), getAll, "expected response not nil")
	require.NotEmpty(s.T(), getAll.Entries, "expected not empty entries response")
	for i, entry := range getAll.Entries {
		assert.Equal(s.T(), createEntries[i].Key, entry.Key, "entry key mismatch")
		assert.Equal(s.T(), createEntries[i].Type, entry.Type, "entry type mismatch")
		assert.True(s.T(), reflect.DeepEqual(createEntries[i].Meta, entry.Meta), "entry meta mismatch")
		assert.Equal(s.T(), createEntries[i].Data, entry.Data, "entry data mismatch")
	}
	get, err := entryService.Get(ctx, &pb.GetEntryRequest{Id: getAll.Entries[0].Id})
	require.NoError(s.T(), err, "no error expected")
	require.NotNil(s.T(), get, "expected response not nil")
	require.NotNil(s.T(), get.Entry, "expected not empty entry response")
	require.NotEmpty(s.T(), get.Entry.Key)
	assert.Equal(s.T(), getAll.Entries[0].Key, get.Entry.Key, "entry key mismatch")
	assert.Equal(s.T(), getAll.Entries[0].Type, get.Entry.Type, "entry type mismatch")
	assert.True(s.T(), reflect.DeepEqual(getAll.Entries[0].Meta, get.Entry.Meta), "entry meta mismatch")
	assert.Equal(s.T(), getAll.Entries[0].Data, get.Entry.Data, "entry data mismatch")

	// 2.2 Entries: delete
	entries := getAll.Entries
	deleted, err := entryService.Delete(ctx, &pb.DeleteEntryRequest{Id: entries[0].Id})
	require.NoError(s.T(), err, "no error expected")
	assert.Equal(s.T(), entries[0].Id, deleted.Id, "deleted entry id mismatch")
	assert.Equal(s.T(), entries[0].CreatedAt, deleted.CreatedAt, "deleted entry created_at mismatch")
	assert.Equal(s.T(), entries[0].UpdatedAt, deleted.UpdatedAt, "deleted entry updated_at mismatch")
	entries = entries[1:]
	getAll, err = entryService.GetAll(ctx, &pb.GetAllEntriesRequest{})
	require.NoError(s.T(), err, "no error expected")
	require.NotNil(s.T(), getAll, "expected response not nil")
	require.NotEmpty(s.T(), getAll.Entries, "expected not empty entries response")
	for i, entry := range getAll.Entries {
		assert.NotEmpty(s.T(), entry.Key, "expected not empty key")
		assert.Equal(s.T(), entries[i].Key, entry.Key, "entry key mismatch")
		assert.Equal(s.T(), entries[i].Type, entry.Type, "entry type mismatch")
		assert.True(s.T(), reflect.DeepEqual(entries[i].Meta, entry.Meta), "entry meta mismatch")
		assert.Equal(s.T(), entries[i].Data, entry.Data, "entry data")
	}

	// 2.3 Entries: update
	entries[0].Meta["updated_key"] = "updated_value"
	entries[0].Data = []byte("updated_data")
	updated, err := entryService.Update(ctx, &pb.UpdateEntryRequest{
		Id:   entries[0].Id,
		Meta: entries[0].Meta,
		Data: entries[0].Data,
	})
	require.NoError(s.T(), err, "no error expected on update")
	assert.Equal(s.T(), entries[0].Id, updated.Id, "updated entry id mismatch")
	assert.Equal(s.T(), entries[0].CreatedAt, updated.CreatedAt, "updated entry created_at mismatch")
	assert.Greater(s.T(), updated.UpdatedAt, entries[0].UpdatedAt, "updated_at should changed")
	_, err = entryService.Get(ctx, &pb.GetEntryRequest{Id: uuid.New().String()})
	require.Error(s.T(), err, "expected error on unknown entry")
	require.Equalf(s.T(), codes.NotFound, status.Code(err), "expected not found code, got %v", status.Code(err))
	get, err = entryService.Get(ctx, &pb.GetEntryRequest{Id: entries[0].Id})
	require.NoError(s.T(), err, "no error expected on get")
	require.NotNil(s.T(), get, "expected response not nil")
	require.NotNil(s.T(), get.Entry, "expected entry not nil")
	assert.Equal(s.T(), entries[0].Id, get.Entry.Id, "get entry id mismatch")
	assert.Equal(s.T(), entries[0].Key, get.Entry.Key, "get entry key mismatch")
	assert.Equal(s.T(), entries[0].Type, get.Entry.Type, "get entry type mismatch")
	assert.True(s.T(), reflect.DeepEqual(entries[0].Meta, get.Entry.Meta), "get entry meta mismatch")
	assert.Equal(s.T(), entries[0].Data, get.Entry.Data, "get entry data mismatch")
	assert.Equal(s.T(), entries[0].CreatedAt, get.Entry.CreatedAt, "get entry created_at mismatch")
	assert.Equal(s.T(), updated.UpdatedAt, get.Entry.UpdatedAt, "get entry updated_at mismatch")
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
