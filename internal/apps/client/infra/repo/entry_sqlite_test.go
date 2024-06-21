package repo

import (
	"context"
	trmsqlx "github.com/avito-tech/go-transaction-manager/drivers/sqlx/v2"
	"github.com/dlomanov/gophkeeper/internal/apps/client/entities"
	"github.com/dlomanov/gophkeeper/internal/apps/client/migrations"
	"github.com/dlomanov/gophkeeper/internal/core"
	"github.com/dlomanov/gophkeeper/internal/infra/migrator"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
	"reflect"
	"testing"
)

type TestEntryRepoSuite struct {
	suite.Suite
	db     *sqlx.DB
	logger *zap.Logger
}

func (s *TestEntryRepoSuite) SetupSuite() {
	var err error
	s.logger = zaptest.NewLogger(s.T(), zaptest.Level(zap.DebugLevel))
	s.db, err = sqlx.Open("sqlite3", "file:test.db?cache=shared&mode=memory")
	require.NoError(s.T(), err, "failed to open database")
	ms, err := migrations.GetMigrations()
	require.NoError(s.T(), err, "failed to get migrations")
	err = migrator.Migrate(s.logger.Sugar(), s.db.DB, ms)
	require.NoError(s.T(), err, "failed to up migrations")
}

func (s *TestEntryRepoSuite) TearDownSuite() {
	err := s.db.Close()
	require.NoError(s.T(), err, "failed to close database")
}

func TestEntryRepo(t *testing.T) {
	suite.Run(t, new(TestEntryRepoSuite))
}

func (s *TestEntryRepoSuite) TestMethods() {
	ctx := context.Background()
	sut := NewEntryRepo(s.db, trmsqlx.DefaultCtxGetter)

	// get all
	entries, err := sut.GetAll(ctx)
	require.NoError(s.T(), err, "failed to get entries")
	require.Empty(s.T(), entries, "entries should be empty")

	// create + get all
	createEntries := make(map[uuid.UUID]*entities.Entry, 3)
	var entry *entities.Entry
	entry, err = entities.NewEntry("key1", core.EntryTypeBinary, []byte("data1"))
	require.NoError(s.T(), err, "failed to create entry")
	createEntries[entry.ID] = entry
	entry, err = entities.NewEntry("key2", core.EntryTypePassword, []byte("data2"))
	require.NoError(s.T(), err, "failed to create entry")
	createEntries[entry.ID] = entry
	entry, err = entities.NewEntry("key3", core.EntryTypeCard, []byte("data3"))
	require.NoError(s.T(), err, "failed to create entry")
	createEntries[entry.ID] = entry
	for _, v := range createEntries {
		err = sut.Create(ctx, *v)
		require.NoError(s.T(), err, "failed to create entry")
	}
	entries, err = sut.GetAll(ctx)
	require.NoError(s.T(), err, "failed to get entries")
	require.Len(s.T(), entries, len(createEntries), "entries should be equal")
	for _, v := range entries {
		entry, ok := createEntries[v.ID]
		require.True(s.T(), ok, "entry not found")
		require.Equal(s.T(), v.ID, entry.ID)
		require.Equal(s.T(), v.Key, entry.Key)
		require.Equal(s.T(), v.Type, entry.Type)
		require.Equal(s.T(), v.Data, entry.Data)
		require.True(s.T(), reflect.DeepEqual(v.Meta, entry.Meta))
		require.Equal(s.T(), v.Version, entry.Version)
		require.Equal(s.T(), v.GlobalVersion, entry.GlobalVersion)
		require.Equal(s.T(), v.CreatedAt, entry.CreatedAt)
		require.Equal(s.T(), v.UpdatedAt, entry.UpdatedAt)
	}

	// update + get
	err = entries[0].Update(entities.UpdateEntryMeta(nil), entities.UpdateEntryData([]byte("data1_new")))
	require.NoError(s.T(), err, "failed to update entry")
	err = sut.Update(ctx, entries[0])
	require.NoError(s.T(), err, "failed to update entry")
	getEntry, err := sut.Get(ctx, entries[0].ID)
	require.NoError(s.T(), err, "failed to get entry")
	require.Equal(s.T(), entries[0].ID, getEntry.ID)
	require.Equal(s.T(), entries[0].Key, getEntry.Key)
	require.Equal(s.T(), entries[0].Type, getEntry.Type)
	require.Equal(s.T(), entries[0].Data, getEntry.Data)
	require.True(s.T(), reflect.DeepEqual(entries[0].Meta, getEntry.Meta))
	require.Equal(s.T(), entries[0].Version, getEntry.Version)
	require.Equal(s.T(), entries[0].GlobalVersion, getEntry.GlobalVersion)
	require.Equal(s.T(), entries[0].CreatedAt, getEntry.CreatedAt)
	require.Equal(s.T(), entries[0].UpdatedAt, getEntry.UpdatedAt)

	// delete + get
	err = sut.Delete(ctx, entries[0].ID)
	require.NoError(s.T(), err, "failed to delete entry")
	_, err = sut.Get(ctx, entries[0].ID)
	require.ErrorIs(s.T(), err, entities.ErrEntryNotFound, "expected entry not found error")

	// get versions
	versions, err := sut.GetVersions(ctx)
	require.NoError(s.T(), err, "failed to get entry versions")
	require.Len(s.T(), versions, len(createEntries)-1, "expected 2 versions")
	for _, v := range versions {
		entry, ok := createEntries[v.ID]
		require.True(s.T(), ok, "entry not found")
		require.Equal(s.T(), v.ID, entry.ID)
		require.Equal(s.T(), v.Version, entry.GlobalVersion)
	}
}
