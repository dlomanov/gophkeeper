package container

import (
	"context"
	"errors"
	"fmt"
	"github.com/docker/go-connections/nat"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"strconv"
	"strings"
	"time"
)

const (
	postgresStartup = 5 * time.Second
)

// RunPostgres - starts postgres test container and returns updated DSN with new host and port
func RunPostgres(ctx context.Context, dsn string) (pgc *postgres.PostgresContainer, updatedDSN string, err error) {
	values := strings.Split(dsn, " ")
	if len(values) == 0 {
		return nil, "", errors.New("failed to parse database uri")
	}
	kmap := make(map[int]string, len(values))
	vmap := make(map[string]string, len(values))
	for i, v := range values {
		kv := strings.Split(v, "=")
		if len(kv) != 2 {
			return nil, "", errors.New("failed to parse database uri value")
		}
		kmap[i] = kv[0]
		vmap[kv[0]] = kv[1]
	}
	port, ok := vmap["port"]
	if !ok {
		return nil, "", errors.New("failed to get database port")
	}
	username, ok := vmap["user"]
	if !ok {
		return nil, "", errors.New("failed to get database user")
	}
	password, ok := vmap["password"]
	if !ok {
		return nil, "", errors.New("failed to get database password")
	}
	dbname, ok := vmap["dbname"]
	if !ok {
		return nil, "", errors.New("failed to get database name")
	}

	pgc, err = postgres.RunContainer(ctx,
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
	if err != nil {
		return nil, "", fmt.Errorf("failed to start postgres container: %w", err)
	}

	newHost, err := pgc.Host(ctx)
	if err != nil {
		return nil, "", fmt.Errorf("failed to get postgres container host: %w", err)
	}
	newPort, err := pgc.MappedPort(ctx, nat.Port(port))
	if err != nil {
		return nil, "", fmt.Errorf("failed to get postgres container port: %w", err)
	}
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

	return pgc, dsn, nil
}
