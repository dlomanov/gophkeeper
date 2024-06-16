package usecases_test

import (
	"context"
	trmsqlx "github.com/avito-tech/go-transaction-manager/drivers/sqlx/v2"
	"github.com/avito-tech/go-transaction-manager/trm/v2/manager"
	"github.com/dlomanov/gophkeeper/internal/apps/client/entities"
	"github.com/dlomanov/gophkeeper/internal/apps/client/infra/repo"
	"github.com/dlomanov/gophkeeper/internal/apps/client/infra/services/pass"
	"github.com/dlomanov/gophkeeper/internal/apps/client/migrations"
	"github.com/dlomanov/gophkeeper/internal/apps/client/usecases"
	"github.com/dlomanov/gophkeeper/internal/core"
	"github.com/dlomanov/gophkeeper/internal/infra/migrator"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
	"testing"
)

type TestUserAuthUC struct {
	suite.Suite
	logger *zap.Logger
	db     *sqlx.DB
}

func (s *TestUserAuthUC) SetupSuite() {
	var err error
	s.logger = zaptest.NewLogger(s.T(), zaptest.Level(zap.DebugLevel))
	s.db, err = sqlx.Open("sqlite3", "file:test.db?cache=shared&mode=memory")
	require.NoError(s.T(), err, "failed to open database")
	ms, err := migrations.GetMigrations()
	require.NoError(s.T(), err, "failed to get migrations")
	err = migrator.Migrate(s.logger.Sugar(), s.db.DB, ms)
	require.NoError(s.T(), err, "failed to up migrations")
}

func (s *TestUserAuthUC) TearDownSuite() {
	err := s.db.Close()
	require.NoError(s.T(), err, "failed to close database")
}

func Test(t *testing.T) {
	suite.Run(t, new(TestUserAuthUC))
}

func (s *TestUserAuthUC) TestAuth() {
	ctx := context.Background()

	trm, err := manager.New(trmsqlx.NewDefaultFactory(s.db))
	require.NoError(s.T(), err, "failed to create transaction manager")

	kvRepo := repo.NewKVPairRepo(
		s.db,
		trmsqlx.DefaultCtxGetter,
		trm)

	sut := usecases.NewUserAuthUC(&pass.Hasher{}, kvRepo, trm)
	hash, err := sut.Auth(ctx, core.Pass("password"))
	require.Len(s.T(), hash, 32, "expected 32 bytes hash")
	require.NoError(s.T(), err, "failed to auth user")

	_, err = sut.Auth(ctx, core.Pass("wrong-password"))
	require.ErrorIs(s.T(), err, entities.ErrUserMasterPassInvalid, "expected invalid password error")
	hash1, err := sut.Auth(ctx, core.Pass("password"))
	require.Len(s.T(), hash, 32, "expected 32 bytes hash")
	require.NoError(s.T(), err, "failed to auth user")
	require.Equal(s.T(), hash, hash1, "hashes should be equal")
}
