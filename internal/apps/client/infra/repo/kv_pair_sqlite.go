package repo

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/dlomanov/gophkeeper/internal/apps/client/entities"
	"github.com/jmoiron/sqlx"
)

type KVPairRepo struct {
	db *sqlx.DB
}

func NewKVPairRepo(
	db *sqlx.DB,
) *KVPairRepo {
	return &KVPairRepo{
		db: db,
	}
}

func (r *KVPairRepo) Upload(ctx context.Context, pairs []entities.KVPair) error {
	if len(pairs) == 0 {
		return nil
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("user_kv_repo: failed to begin tx: %w", err)
	}
	stmt, err := tx.PrepareContext(ctx, `
		insert into user_kv (key, value) values ($1, $2)
		on conflict (key) do update set value = excluded.value;`)
	if err != nil {
		return fmt.Errorf("user_kv_repo: failed to prepare statement: %w", err)
	}
	defer func(stmt *sql.Stmt) { _ = stmt.Close() }(stmt)
	for _, pair := range pairs {
		_, err = stmt.ExecContext(ctx, pair.Key, pair.Value)
		if err != nil {
			return fmt.Errorf("user_kv_repo: failed to execute statement: %w", err)
		}
	}
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("user_kv_repo: failed to commit tx: %w", err)
	}

	return nil
}

func (r *KVPairRepo) Load(ctx context.Context) ([]entities.KVPair, error) {
	rows, err := r.db.QueryContext(ctx, `select key, value from user_kv;`)
	if err != nil {
		return nil, fmt.Errorf("user_kv_repo: failed to query: %w", err)
	}
	defer func(rows *sql.Rows) { _ = rows.Close() }(rows)
	var pairs []entities.KVPair
	for rows.Next() {
		var pair entities.KVPair
		if err = rows.Scan(&pair.Key, &pair.Value); err != nil {
			return nil, fmt.Errorf("user_kv_repo: failed to scan: %w", err)
		}
		pairs = append(pairs, pair)
	}
	if err = rows.Close(); err != nil {
		return nil, fmt.Errorf("user_kv_repo: failed to close rows: %w", err)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("user_kv_repo: failed to iterate rows: %w", err)
	}
	return pairs, nil
}
