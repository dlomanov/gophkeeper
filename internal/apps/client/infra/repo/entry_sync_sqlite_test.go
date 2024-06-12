package repo

import (
	"context"
	trmsqlx "github.com/avito-tech/go-transaction-manager/drivers/sqlx/v2"
	"github.com/dlomanov/gophkeeper/internal/apps/client/entities"
	"github.com/dlomanov/gophkeeper/internal/apps/client/migrations"
	"github.com/dlomanov/gophkeeper/internal/infra/migrator"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
	"testing"
)

type TestEntrySyncRepoSuite struct {
	suite.Suite
	db     *sqlx.DB
	logger *zap.Logger
}

func (s *TestEntrySyncRepoSuite) SetupSuite() {
	var err error
	s.logger = zaptest.NewLogger(s.T(), zaptest.Level(zap.DebugLevel))
	s.db, err = sqlx.Open("sqlite3", "file:test.db?cache=shared&mode=memory")
	require.NoError(s.T(), err, "failed to open database")
	ms, err := migrations.GetMigrations()
	require.NoError(s.T(), err, "failed to get migrations")
	err = migrator.Migrate(s.logger.Sugar(), s.db.DB, ms)
	require.NoError(s.T(), err, "failed to up migrations")
}

func (s *TestEntrySyncRepoSuite) TearDownSuite() {
	err := s.db.Close()
	require.NoError(s.T(), err, "failed to close database")
}

func TestEntrySyncRepo(t *testing.T) {
	suite.Run(t, new(TestEntrySyncRepoSuite))
}

func (s *TestEntrySyncRepoSuite) TestMethods() {
	ctx := context.Background()
	sut := NewEntrySyncRepo(s.db, trmsqlx.DefaultCtxGetter)

	entries, err := sut.GetAll(ctx)
	require.NoError(s.T(), err, "failed to get entries")
	require.Empty(s.T(), entries, "entries should be empty")
	createEntries := make(map[uuid.UUID]*entities.EntrySync, 3)
	for i := 0; i < 3; i++ {
		id := uuid.New()
		createEntries[id] = entities.NewEntrySync(id)
		err = sut.Create(ctx, *createEntries[id])
		require.NoError(s.T(), err, "failed to create entry")
	}
	entries, err = sut.GetAll(ctx)
	require.NoError(s.T(), err, "failed to get entries")
	for _, entry := range entries {
		require.Equal(s.T(), entry.ID, createEntries[entry.ID].ID)
		err = sut.Delete(ctx, createEntries[entry.ID].ID)
		require.NoError(s.T(), err, "failed to delete entry")
	}
	entries, err = sut.GetAll(ctx)
	require.NoError(s.T(), err, "failed to get entries")
	require.Empty(s.T(), entries, "entries should be empty")
}
