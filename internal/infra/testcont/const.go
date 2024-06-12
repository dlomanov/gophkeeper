package testcont

import "time"

const (
	PostgresStartup = 5 * time.Second
	PostgresDSN     = "host=localhost port=5432 user=postgres password=1 dbname=gophkeeper sslmode=disable"
	TeardownTimeout = 10 * time.Second
)
