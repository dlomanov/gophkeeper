package repo

import (
	"context"
	trmsqlx "github.com/avito-tech/go-transaction-manager/drivers/sqlx/v2"
	"github.com/dlomanov/gophkeeper/internal/apps/server/migrations"
	"github.com/dlomanov/gophkeeper/internal/entities"
	"github.com/dlomanov/gophkeeper/internal/infra/migrator"
	"github.com/docker/go-connections/nat"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
	"strconv"
	"strings"
	"testing"
	"time"
)

const (
	teardownTimeout = 10 * time.Second
	postgresStartup = 5 * time.Second
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
	s.pgc, dsn = runPG(s.T(), dsn)
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

// Run postgres container and returns updated DSN with new host and port
func runPG(t *testing.T, dsn string) (*postgres.PostgresContainer, string) {
	values := strings.Split(dsn, " ")
	require.NotEmpty(t, values, "failed to parse database uri")
	kmap := make(map[int]string, len(values))
	vmap := make(map[string]string, len(values))
	for i, v := range values {
		kv := strings.Split(v, "=")
		require.Len(t, kv, 2, "failed to parse database uri value")
		kmap[i] = kv[0]
		vmap[kv[0]] = kv[1]
	}
	port, ok := vmap["port"]
	require.True(t, ok, "failed to get database port")
	username, ok := vmap["user"]
	require.True(t, ok, "failed to get database user")
	password, ok := vmap["password"]
	require.True(t, ok, "failed to get database password")
	dbname, ok := vmap["dbname"]
	require.True(t, ok, "failed to get database name")

	ctx := context.Background()
	pgc, err := postgres.RunContainer(ctx,
		testcontainers.WithImage("docker.io/postgres:latest"),
		postgres.WithDatabase(dbname),
		postgres.WithUsername(username),
		postgres.WithPassword(password),
		postgres.WithInitScripts(),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(postgresStartup),
		),
	)
	require.NoError(t, err, "failed to start postgres container")

	newHost, err := pgc.Host(ctx)
	require.NoError(t, err, "failed to get postgres container host")
	newPort, err := pgc.MappedPort(ctx, nat.Port(port))
	require.NoError(t, err, "failed to get postgres container port")
	var sb strings.Builder
	for i := 0; i < len(values); i++ {
		k := kmap[i]

		_, _ = sb.WriteString(k)
		_ = sb.WriteByte('=')
		switch {
		case k == "host":
			_, _ = sb.WriteString(newHost)
		case k == "port":
			_, _ = sb.WriteString(strconv.Itoa(newPort.Int()))
		default:
			_, _ = sb.WriteString(vmap[k])
		}
		_ = sb.WriteByte(' ')
	}
	dsn = strings.TrimRight(sb.String(), " ")

	return pgc, dsn
}

func must[T any](t *testing.T, fn func() (T, error)) T {
	v, err := fn()
	require.NoError(t, err, "must not return error")
	return v
}
