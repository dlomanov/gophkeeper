package migrations

import (
	"embed"
	"github.com/dlomanov/gophkeeper/internal/infra/pg/migrator"
)

//go:embed *
var fs embed.FS
var files = []file{
	{Name: "m0001.sql", Title: "M0001: Users table", NoTx: false},
	{Name: "m0002.sql", Title: "M0002: Entries table", NoTx: false},
}

type file struct {
	Name  string
	Title string
	NoTx  bool
}

func GetMigrations() ([]migrator.Migration, error) {
	result := make([]migrator.Migration, len(files))

	for i, f := range files {
		query, err := fs.ReadFile(f.Name)
		if err != nil {
			return nil, err
		}

		result[i] = migrator.Migration{
			Name:  f.Title,
			Query: string(query),
			NoTx:  f.NoTx,
		}
	}

	return result, nil
}
