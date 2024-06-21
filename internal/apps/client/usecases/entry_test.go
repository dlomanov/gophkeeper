package usecases_test

import (
	"context"
	trmsqlx "github.com/avito-tech/go-transaction-manager/drivers/sqlx/v2"
	"github.com/avito-tech/go-transaction-manager/trm/v2/manager"
	"github.com/dlomanov/gophkeeper/internal/apps/client/entities"
	"github.com/dlomanov/gophkeeper/internal/apps/client/infra/repo"
	"github.com/dlomanov/gophkeeper/internal/apps/client/infra/services/marshal"
	"github.com/dlomanov/gophkeeper/internal/apps/client/infra/services/mem"
	"github.com/dlomanov/gophkeeper/internal/apps/client/migrations"
	"github.com/dlomanov/gophkeeper/internal/apps/client/usecases"
	pb "github.com/dlomanov/gophkeeper/internal/apps/shared/proto"
	"github.com/dlomanov/gophkeeper/internal/apps/shared/proto/mocks"
	"github.com/dlomanov/gophkeeper/internal/core"
	"github.com/dlomanov/gophkeeper/internal/infra/encrypto"
	"github.com/dlomanov/gophkeeper/internal/infra/migrator"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
	"testing"
	"time"
)

type TestEntryUC struct {
	suite.Suite
	logger *zap.Logger
	db     *sqlx.DB
}

func (s *TestEntryUC) SetupSuite() {
	var err error
	s.logger = zaptest.NewLogger(s.T(), zaptest.Level(zap.DebugLevel))
	s.db, err = sqlx.Open("sqlite3", "file:test.db?cache=shared&mode=memory")
	require.NoError(s.T(), err, "failed to open database")
	ms, err := migrations.GetMigrations()
	require.NoError(s.T(), err, "failed to get migrations")
	err = migrator.Migrate(s.logger.Sugar(), s.db.DB, ms)
	require.NoError(s.T(), err, "failed to up migrations")
}

func (s *TestEntryUC) TearDownSuite() {
	err := s.db.Close()
	require.NoError(s.T(), err, "failed to close database")
}

func TestEntryUCRun(t *testing.T) {
	suite.Run(t, new(TestEntryUC))
}

func (s *TestEntryUC) TestMethods() {
	ctx := context.Background()
	ctrl := gomock.NewController(s.T())
	defer ctrl.Finish()

	client := mocks.NewMockEntryServiceClient(ctrl)
	client.EXPECT().Create(gomock.Any(), gomock.Any()).AnyTimes().Return(nil, nil)
	client.EXPECT().Update(gomock.Any(), gomock.Any()).AnyTimes().Return(nil, nil)
	client.EXPECT().Delete(gomock.Any(), gomock.Any()).AnyTimes().Return(nil, nil)
	client.EXPECT().GetDiff(gomock.Any(), gomock.Any()).AnyTimes().Return(&pb.GetEntriesDiffResponse{}, nil)

	encrypter, err := encrypto.NewEncrypter([]byte("1234567890123456"))
	require.NoError(s.T(), err, "failed to create encrypter")
	entryRepo := repo.NewEntryRepo(s.db, trmsqlx.DefaultCtxGetter)
	entrySyncRepo := repo.NewEntrySyncRepo(s.db, trmsqlx.DefaultCtxGetter)
	memcache := mem.NewCache()
	trm, err := manager.New(trmsqlx.NewDefaultFactory(s.db))
	require.NoError(s.T(), err, "failed to create transaction manager")
	sut := usecases.NewEntriesUC(
		s.logger,
		client,
		entryRepo,
		entrySyncRepo,
		encrypter,
		marshal.EntryMarshaler{},
		memcache,
		trm,
	)

	// create
	createEntries := map[string]entities.CreateEntryRequest{
		"key1": {
			Key:  "key1",
			Type: core.EntryTypePassword,
			Meta: map[string]string{"description": "description1"},
			Data: entities.EntryDataPassword{
				Login:    "login1",
				Password: "password1",
			},
		},
		"key2": {
			Key:  "key2",
			Type: core.EntryTypeNote,
			Meta: map[string]string{"description": "description2"},
			Data: entities.EntryDataNote("note2"),
		},
		"key3": {
			Key:  "key3",
			Type: core.EntryTypeCard,
			Meta: map[string]string{"description": "description3"},
			Data: entities.EntryDataCard{
				Number:  "number3",
				Expires: "expires3",
				Cvc:     "cvc3",
				Owner:   "owner3 ",
			},
		},
		"key4": {
			Key:  "key4",
			Type: core.EntryTypeBinary,
			Meta: map[string]string{
				"filename":    "filename4",
				"description": "description4",
			},
			Data: entities.EntryDataBinary("binary4"),
		},
	}
	for _, entry := range createEntries {
		created, err := sut.Create(ctx, entry)
		time.Sleep(time.Millisecond * 300)
		require.NoError(s.T(), err, "failed to create entry")
		require.NotEmpty(s.T(), created.ID, "entry ID should not be empty")
	}
	getAll, err := sut.GetAll(ctx)
	require.NoError(s.T(), err, "failed to get all createEntries")
	require.Len(s.T(), getAll.Entries, len(createEntries), "wrong createEntries count")
	entries := make(map[uuid.UUID]entities.GetEntryResponse)
	updateIdx := 0
	for i, entry := range getAll.Entries {
		require.NotEmpty(s.T(), entry.ID, "entry ID should not be empty")
		require.Equal(s.T(), entry.Key, createEntries[entry.Key].Key, "entry key mismatch")
		require.Equal(s.T(), entry.Type, createEntries[entry.Key].Type, "entry type mismatch")
		require.Equal(s.T(), entry.Data, createEntries[entry.Key].Data, "entry data mismatch")
		require.Equal(s.T(), entry.Meta, createEntries[entry.Key].Meta, "entry meta mismatch")
		entries[entry.ID] = entry
		if entry.Type == core.EntryTypePassword {
			updateIdx = i
		}
	}

	// update
	updateEntry := getAll.Entries[updateIdx]
	require.Equal(s.T(), core.EntryTypePassword, updateEntry.Type, "entry type mismatch")
	updateEntry.Meta["description"] = "updated description"
	updateEntry.Data = entities.EntryDataPassword{
		Login:    "updated login",
		Password: "updated password",
	}
	entries[updateEntry.ID] = updateEntry
	err = sut.Update(ctx, entities.UpdateEntryRequest{
		ID:   updateEntry.ID,
		Meta: updateEntry.Meta,
		Data: updateEntry.Data,
	})
	require.NoError(s.T(), err, "failed to update entry")
	getAll, err = sut.GetAll(ctx)
	require.NoError(s.T(), err, "failed to get all createEntries")
	require.Len(s.T(), getAll.Entries, len(entries), "wrong createEntries count")
	for _, entry := range getAll.Entries {
		require.NotEmpty(s.T(), entry.ID, "entry ID should not be empty")
		require.Equal(s.T(), entry.ID, entries[entry.ID].ID, "entry id mismatch")
		require.Equal(s.T(), entry.Key, entries[entry.ID].Key, "entry key mismatch")
		require.Equal(s.T(), entry.Type, entries[entry.ID].Type, "entry type mismatch")
		require.Equal(s.T(), entry.Data, entries[entry.ID].Data, "entry data mismatch")
		require.Equal(s.T(), entry.Meta, entries[entry.ID].Meta, "entry meta mismatch")
	}

	// delete
	err = sut.Delete(ctx, entities.DeleteEntryRequest{ID: updateEntry.ID})
	require.NoError(s.T(), err, "failed to delete entry")
	getAll, err = sut.GetAll(ctx)
	require.NoError(s.T(), err, "failed to get all createEntries")
	require.Len(s.T(), getAll.Entries, len(entries)-1, "unexpected entries count after deletion")

	// sync
	memcache.SetString("token", "token-value")
	err = sut.Sync(ctx)
	require.NoError(s.T(), err, "failed to sync entries")
}
