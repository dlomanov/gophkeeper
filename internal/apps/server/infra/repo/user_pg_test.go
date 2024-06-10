package repo_test

import (
	"context"
	trmsqlx "github.com/avito-tech/go-transaction-manager/drivers/sqlx/v2"
	"github.com/dlomanov/gophkeeper/internal/apps/server/entities"
	"github.com/dlomanov/gophkeeper/internal/apps/server/infra/repo"
	"github.com/dlomanov/gophkeeper/internal/apps/server/migrations"
	"github.com/dlomanov/gophkeeper/internal/infra/pg/migrator"
	"github.com/dlomanov/gophkeeper/internal/infra/pg/testcont"
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

type UserTestSuit struct {
	suite.Suite
	teardownCtx context.Context
	logger      *zap.Logger
	pgc         *postgres.PostgresContainer
	db          *sqlx.DB
	teardown    func()
}

func (s *UserTestSuit) SetupSuite() {
	var err error
	s.logger = zaptest.NewLogger(s.T(), zaptest.Level(zap.DebugLevel))
	s.teardownCtx, s.teardown = context.WithCancel(context.Background())

	dsn := testcont.PostgresDSN
	s.pgc, dsn, err = testcont.RunPostgres(s.teardownCtx, dsn)
	require.NoError(s.T(), err, "no error expected")
	s.db, err = sqlx.ConnectContext(s.teardownCtx, "pgx", dsn)
	require.NoError(s.T(), err)

	ms, err := migrations.GetMigrations()
	require.NoError(s.T(), err, "no error expected")
	err = migrator.Migrate(s.logger.Sugar(), s.db.DB, ms)
	require.NoError(s.T(), err, "no error expected")
}

func (s *UserTestSuit) TearDownSuite() {
	s.teardown()

	if err := s.db.Close(); err != nil {
		s.logger.Error("failed to close postgres db", zap.Error(err))
	}

	timeout, cancel := context.WithTimeout(context.Background(), testcont.TeardownTimeout)
	defer cancel()
	if err := s.pgc.Terminate(timeout); err != nil {
		s.logger.Error("failed to terminate postgres container", zap.Error(err))
	}
}

func TestUserRun(t *testing.T) {
	suite.Run(t, new(UserTestSuit))
}

func (s *UserTestSuit) TestUserRepo() {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	r := repo.NewUserRepo(s.db, trmsqlx.DefaultCtxGetter)
	login := entities.Login("testUser")

	_, err := r.Get(ctx, login)
	require.ErrorIs(s.T(), err, entities.ErrUserNotFound, "expected user not found error")

	exists, err := r.Exists(ctx, login)
	require.NoError(s.T(), err, "no error expected")
	require.False(s.T(), exists, "expected user not found")

	user := must(s.T(), func() (*entities.User, error) {
		return entities.NewUser(entities.HashCreds{
			Login:    login,
			PassHash: "hash",
		})
	})

	err = r.Create(ctx, *user)
	require.NoError(s.T(), err, "no error expected")

	err = r.Create(ctx, *user)
	require.ErrorIs(s.T(), err, entities.ErrUserExists, "expected user already exists error")

	exists, err = r.Exists(ctx, login)
	require.NoError(s.T(), err, "no error expected")
	require.True(s.T(), exists, "expected user found")

	user1, err := r.Get(ctx, login)
	require.NoError(s.T(), err, "no error expected")

	require.Equal(s.T(), user.ID, user1.ID, "expected same user IDs")
	require.Equal(s.T(), user.HashCreds, user1.HashCreds, "expected same user creds")
	require.Equal(s.T(), user.CreatedAt.Format("2006-01-02 15:04:05.000"), user1.CreatedAt.Format("2006-01-02 15:04:05.000"), "expected same user created at")
	require.Equal(s.T(), user.UpdatedAt.Format("2006-01-02 15:04:05.000"), user1.UpdatedAt.Format("2006-01-02 15:04:05.000"), "expected same user updated at")
}

func must[T any](t *testing.T, fn func() (T, error)) T {
	v, err := fn()
	require.NoError(t, err, "must not return error")
	return v
}
