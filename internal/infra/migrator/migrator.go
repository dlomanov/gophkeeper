package migrator

import (
	"database/sql"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/lopezator/migrator"
	"go.uber.org/zap"
)

const migrationTable = "__migrations"

type Migration struct {
	Name  string
	Query string
	NoTx  bool
}

func Migrate(sugar *zap.SugaredLogger, db *sql.DB, ms []Migration) error {
	logger := migrator.LoggerFunc(func(msg string, args ...any) { sugar.Info(msg, args) })
	m, err := migrator.New(
		toOptions(ms),
		migrator.WithLogger(logger),
		migrator.TableName(migrationTable),
	)
	if err != nil {
		return err
	}
	return m.Migrate(db)
}

func toOptions(ms []Migration) migrator.Option {
	result := make([]any, len(ms))
	for i, m := range ms {
		if m.NoTx {
			result[i] = newMigrationNoTx(m.Name, m.Query)
			continue
		}
		result[i] = newMigration(m.Name, m.Query)
	}

	return migrator.Migrations(result...)
}

func newMigration(name, query string) *migrator.Migration {
	return &migrator.Migration{
		Name: name,
		Func: func(tx *sql.Tx) error {
			if _, err := tx.Exec(query); err != nil {
				return err
			}
			return nil
		},
	}
}

func newMigrationNoTx(name, query string) *migrator.MigrationNoTx {
	return &migrator.MigrationNoTx{
		Name: name,
		Func: func(tx *sql.DB) error {
			if _, err := tx.Exec(query); err != nil {
				return err
			}
			return nil
		},
	}
}
