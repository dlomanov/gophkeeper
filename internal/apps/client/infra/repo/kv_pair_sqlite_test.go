package repo

import (
	"context"
	trmsqlx "github.com/avito-tech/go-transaction-manager/drivers/sqlx/v2"
	"github.com/avito-tech/go-transaction-manager/trm/v2/manager"
	"github.com/dlomanov/gophkeeper/internal/apps/client/entities"
	"github.com/dlomanov/gophkeeper/internal/apps/client/migrations"
	"github.com/dlomanov/gophkeeper/internal/infra/migrator"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
	"testing"
)

type KVPairRepoTestSuit struct {
	suite.Suite
	db     *sqlx.DB
	logger *zap.Logger
}

func (s *KVPairRepoTestSuit) SetupSuite() {
	var err error
	s.logger = zaptest.NewLogger(s.T(), zaptest.Level(zap.DebugLevel))
	s.db, err = sqlx.Open("sqlite3", "file:test.db?cache=shared&mode=memory")
	require.NoError(s.T(), err, "failed to open database")
	ms, err := migrations.GetMigrations()
	require.NoError(s.T(), err, "failed to get migrations")
	err = migrator.Migrate(s.logger.Sugar(), s.db.DB, ms)
	require.NoError(s.T(), err, "failed to up migrations")
}

func (s *KVPairRepoTestSuit) TearDownSuite() {
	err := s.db.Close()
	require.NoError(s.T(), err, "failed to close database")
}

func Test(t *testing.T) {
	suite.Run(t, new(KVPairRepoTestSuit))
}

func (s *KVPairRepoTestSuit) TestMethods() {
	ctx := context.Background()

	trm, err := manager.New(trmsqlx.NewDefaultFactory(s.db))
	require.NoError(s.T(), err, "failed to create transaction manager")
	sut := NewKVPairRepo(s.db, trmsqlx.DefaultCtxGetter, trm)
	pairs, err := sut.Load(ctx)
	require.NoError(s.T(), err, "failed to load pairs")
	require.Empty(s.T(), pairs, "pairs should be empty")

	m := map[string]string{
		"key1": "value1",
		"key2": "value2",
		"key3": "value3",
	}
	uploadPairs := make([]entities.KVPair, 0, len(m))
	for k, v := range m {
		uploadPairs = append(uploadPairs, entities.KVPair{Key: k, Value: v})
	}
	err = sut.Upload(ctx, uploadPairs)
	require.NoError(s.T(), err, "failed to upload pairs")

	pairs, err = sut.Load(ctx)
	require.NoError(s.T(), err, "failed to load pairs")
	require.Len(s.T(), pairs, len(m), "pairs should be equal")
	for _, pair := range pairs {
		require.Equal(s.T(), pair.Value, m[pair.Key], "pairs should be equal")
	}

	v1, err := sut.Get(ctx, "key1")
	require.NoError(s.T(), err, "failed to get value")
	require.Equal(s.T(), v1, m["key1"], "value should be equal")

	err = sut.Set(ctx, "key1", "value1_new")
	require.NoError(s.T(), err, "failed to set value")
	v1, err = sut.Get(ctx, "key1")
	require.NoError(s.T(), err, "failed to get value")
	require.Equal(s.T(), v1, "value1_new", "value should be equal")

	err = sut.Set(ctx, "key4", "value4")
	require.NoError(s.T(), err, "failed to set value")
	v4, err := sut.Get(ctx, "key4")
	require.NoError(s.T(), err, "failed to get value")
	require.Equal(s.T(), v4, "value4", "value should be equal")

	_, err = sut.Get(ctx, "key5")
	require.ErrorIs(s.T(), err, entities.ErrKVPairNotFound, "expected not found error")
}
