package repo

import (
	"context"
	trmsqlx "github.com/avito-tech/go-transaction-manager/drivers/sqlx/v2"
	"github.com/dlomanov/gophkeeper/internal/apps/server/migrations"
	"github.com/dlomanov/gophkeeper/internal/entities"
	"github.com/dlomanov/gophkeeper/internal/infra/migrator"
	"github.com/dlomanov/gophkeeper/internal/infra/testing/container"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
	"testing"
	"time"
)

const (
	teardownTimeout = 10 * time.Second
	dsnDefault      = "host=localhost port=5432 user=postgres password=1 dbname=gophkeeper sslmode=disable"
)

type TestSuit struct {
	suite.Suite
	teardownCtx context.Context
	logger      *zap.Logger
	pgc         *postgres.PostgresContainer
	db          *sqlx.DB
	teardown    func()
}

func TestRun(t *testing.T) {
	suite.Run(t, new(TestSuit))
}

func (s *TestSuit) SetupSuite() {
	var err error
	s.logger = zaptest.NewLogger(s.T(), zaptest.Level(zap.DebugLevel))
	s.teardownCtx, s.teardown = context.WithCancel(context.Background())

	dsn := dsnDefault
	s.pgc, dsn, err = container.RunPostgres(s.teardownCtx, dsn)
	s.db, err = sqlx.ConnectContext(s.teardownCtx, "pgx", dsn)
	require.NoError(s.T(), err)

	ms, err := migrations.GetMigrations()
	require.NoError(s.T(), err, "no error expected")
	err = migrator.Migrate(s.logger.Sugar(), s.db.DB, ms)
	require.NoError(s.T(), err, "no error expected")
}

func (s *TestSuit) TearDownSuite() {
	s.teardown()

	if err := s.db.Close(); err != nil {
		s.logger.Error("failed to close postgres db", zap.Error(err))
	}

	timeout, cancel := context.WithTimeout(context.Background(), teardownTimeout)
	defer cancel()
	if err := s.pgc.Terminate(timeout); err != nil {
		s.logger.Error("failed to terminate postgres container", zap.Error(err))
	}
}

func (s *TestSuit) TestUserRepo() {
	repo := NewUserRepo(s.db, trmsqlx.DefaultCtxGetter)
	login := entities.Login("testUser")

	_, err := repo.Get(s.teardownCtx, login)
	require.ErrorIs(s.T(), err, entities.ErrUserNotFound, "expected user not found error")

	exists, err := repo.Exists(s.teardownCtx, login)
	require.NoError(s.T(), err, "no error expected")
	require.False(s.T(), exists, "expected user not found")

	user := must(s.T(), func() (*entities.User, error) {
		return entities.NewUser(entities.HashCreds{
			Login:    login,
			PassHash: "hash",
		})
	})

	err = repo.Create(s.teardownCtx, *user)
	require.NoError(s.T(), err, "no error expected")

	err = repo.Create(s.teardownCtx, *user)
	require.ErrorIs(s.T(), err, entities.ErrUserExists, "expected user already exists error")

	exists, err = repo.Exists(s.teardownCtx, login)
	require.NoError(s.T(), err, "no error expected")
	require.True(s.T(), exists, "expected user found")

	user1, err := repo.Get(s.teardownCtx, login)
	require.NoError(s.T(), err, "no error expected")

	require.Equal(s.T(), user.ID, user1.ID, "expected same user IDs")
	require.Equal(s.T(), user.HashCreds, user1.HashCreds, "expected same user creds")
	require.Equal(s.T(), user.CreatedAt.Format("2006-01-02 15:04:05"), user1.CreatedAt.Format("2006-01-02 15:04:05"), "expected same user created at")
	require.Equal(s.T(), user.UpdatedAt.Format("2006-01-02 15:04:05"), user1.UpdatedAt.Format("2006-01-02 15:04:05"), "expected same user updated at")
}

func must[T any](t *testing.T, fn func() (T, error)) T {
	v, err := fn()
	require.NoError(t, err, "must not return error")
	return v
}
