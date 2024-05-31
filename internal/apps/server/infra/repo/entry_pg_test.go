package repo_test

import (
	"context"
	trmsqlx "github.com/avito-tech/go-transaction-manager/drivers/sqlx/v2"
	"github.com/dlomanov/gophkeeper/internal/apps/server/infra/repo"
	"github.com/dlomanov/gophkeeper/internal/apps/server/migrations"
	"github.com/dlomanov/gophkeeper/internal/entities"
	"github.com/dlomanov/gophkeeper/internal/infra/pg/migrator"
	"github.com/dlomanov/gophkeeper/internal/infra/pg/testcont"
	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
	"reflect"
	"testing"
	"time"
)

type EntryTestSuit struct {
	suite.Suite
	teardownCtx context.Context
	logger      *zap.Logger
	pgc         *postgres.PostgresContainer
	db          *sqlx.DB
	teardown    func()
}

func (s *EntryTestSuit) SetupSuite() {
	s.logger = zaptest.NewLogger(s.T(), zaptest.Level(zap.DebugLevel))
	s.teardownCtx, s.teardown = context.WithCancel(context.Background())

	dsn := testcont.PostgresDSN
	s.pgc, dsn, _ = testcont.RunPostgres(s.teardownCtx, dsn)
	s.db, _ = sqlx.ConnectContext(s.teardownCtx, "pgx", dsn)

	ms, err := migrations.GetMigrations()
	require.NoError(s.T(), err, "no error expected")
	err = migrator.Migrate(s.logger.Sugar(), s.db.DB, ms)
	require.NoError(s.T(), err, "no error expected")
}

func (s *EntryTestSuit) TearDownSuite() {
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

func TestEntryRun(t *testing.T) {
	suite.Run(t, new(EntryTestSuit))
}

func (s *EntryTestSuit) TestEntryRepo() {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	user, err := entities.NewUser(entities.HashCreds{
		Login:    "test_user",
		PassHash: "test_password_hash",
	})
	require.NoError(s.T(), err, "no error expected when creating user")
	userRepo := repo.NewUserRepo(s.db, trmsqlx.DefaultCtxGetter)
	err = userRepo.Create(ctx, *user)
	require.NoError(s.T(), err, "no error expected when creating user in storage")

	entryRepo := repo.NewEntryRepo(s.db, trmsqlx.DefaultCtxGetter)
	_, err = entryRepo.Get(ctx, user.ID, uuid.New())
	require.ErrorIs(s.T(), err, entities.ErrEntryNotFound, "expected entry not found error")
	result, err := entryRepo.GetAll(ctx, user.ID)
	require.NoError(s.T(), err, "no error expected when getting entries")
	require.Empty(s.T(), result, "expected empty entries")

	entries := make([]*entities.Entry, 3)
	entries[0], err = entities.NewEntry("key1", user.ID, entities.EntryTypePassword, []byte("test_data_1"))
	entries[0].Meta = map[string]string{"key1": "value1", "key2": "value2"}
	require.NoError(s.T(), err, "no error expected when creating entry")
	entries[1], err = entities.NewEntry("key2", user.ID, entities.EntryTypeBinary, []byte("test_data_2"))
	require.NoError(s.T(), err, "no error expected when creating entry")
	entries[2], err = entities.NewEntry("key3", user.ID, entities.EntryTypeNote, []byte("test_data_3"))
	require.NoError(s.T(), err, "no error expected when creating entry")
	for _, entry := range entries {
		err = entryRepo.Create(ctx, entry)
		require.NoError(s.T(), err, "no error expected when creating entry in storage")
	}
	err = entryRepo.Create(ctx, entries[0])
	require.Error(s.T(), err, "expected entry already exists error")
	require.ErrorIs(s.T(), err, entities.ErrEntryExists, "expected entry already exists error")

	result, err = entryRepo.GetAll(ctx, user.ID)
	require.NoError(s.T(), err, "no error expected when getting entries")
	require.Equal(s.T(), 3, len(result), "expected 3 entries")
	for i, x := range result {
		s.assertEquals(s.T(), entries[i], &x)
	}

	entries[0].Meta["test_key"] = "test_value"
	err = entryRepo.Update(ctx, entries[0])
	require.NoError(s.T(), err, "no error expected when updating entry in storage")

	resultEntry, err := entryRepo.Get(ctx, user.ID, entries[0].ID)
	require.NoError(s.T(), err, "no error expected when getting entry")
	s.assertEquals(s.T(), entries[0], resultEntry)
}

func (s *EntryTestSuit) assertEquals(t *testing.T, expected *entities.Entry, actual *entities.Entry) {
	assert.Equal(t, expected.ID.String(), actual.ID.String(), "expected same entry IDs")
	assert.Equal(t, expected.UserID.String(), actual.UserID.String(), "expected same user IDs")
	assert.Equal(t, expected.Type, actual.Type, "expected same entry types")
	assert.True(t, reflect.DeepEqual(expected.Meta, actual.Meta), "expected same entry meta")
	assert.Equal(t, expected.Data, actual.Data, "expected same entry data")
	assert.Equal(t, expected.CreatedAt.Format("2006-01-02 15:04:05.000"), actual.CreatedAt.Format("2006-01-02 15:04:05.000"), "expected same entry created at")
	assert.Equal(t, expected.UpdatedAt.Format("2006-01-02 15:04:05.000"), actual.UpdatedAt.Format("2006-01-02 15:04:05.000"), "expected same entry updated at")
}
