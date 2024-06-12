package repo

import (
	"context"
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

func TestKVPairRepo(t *testing.T) {
	suite.Run(t, new(KVPairRepoTestSuit))
}

func (s *KVPairRepoTestSuit) TestMethods() {
	ctx := context.Background()

	sut := NewKVPairRepo(s.db)
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
}
