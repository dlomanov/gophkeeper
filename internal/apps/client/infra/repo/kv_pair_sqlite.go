package repo

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	trmsqlx "github.com/avito-tech/go-transaction-manager/drivers/sqlx/v2"
	"github.com/avito-tech/go-transaction-manager/trm/v2"
	"github.com/dlomanov/gophkeeper/internal/apps/client/entities"
	"github.com/jmoiron/sqlx"
)

type KVPairRepo struct {
	db     *sqlx.DB
	getter *trmsqlx.CtxGetter
	trm    trm.Manager
}

func NewKVPairRepo(
	db *sqlx.DB,
	getter *trmsqlx.CtxGetter,
	trm trm.Manager,
) *KVPairRepo {
	return &KVPairRepo{
		db:     db,
		getter: getter,
		trm:    trm,
	}
}

func (r *KVPairRepo) Get(ctx context.Context, key string) (string, error) {
	var value string
	err := r.getDB(ctx).GetContext(ctx, &value, `select value from user_kv where key = $1`, key)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return "", entities.ErrKVPairNotFound
	case err != nil:
		return "", fmt.Errorf("user_kv_repo: failed to get value: %w", err)
	}
	return value, nil
}

func (r *KVPairRepo) Set(ctx context.Context, key string, value string) error {
	res, err := r.getDB(ctx).ExecContext(ctx, `
		insert into user_kv (key, value) values ($1, $2)
		on conflict (key)
		    do update set value = excluded.value`, key, value)
	if err != nil {
		return fmt.Errorf("user_kv_repo: failed to set value: %w", err)
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("user_kv_repo: failed to get rows affected: %w", err)
	}
	if affected == 0 {
		return entities.ErrKVPairNotFound
	}
	return err
}

func (r *KVPairRepo) Upload(ctx context.Context, pairs []entities.KVPair) error {
	if len(pairs) == 0 {
		return nil
	}
	upload := func(ctx context.Context, db trmsqlx.Tr) error {
		stmt, err := db.PreparexContext(ctx, `
			insert into user_kv (key, value) values ($1, $2)
			on conflict (key) do update set value = excluded.value;`)
		if err != nil {
			return fmt.Errorf("user_kv_repo: failed to prepare statement: %w", err)
		}
		defer func(stmt *sqlx.Stmt) { _ = stmt.Close() }(stmt)
		for _, pair := range pairs {
			_, err = stmt.ExecContext(ctx, pair.Key, pair.Value)
			if err != nil {
				return fmt.Errorf("user_kv_repo: failed to execute statement: %w", err)
			}
		}
		return nil
	}
	if err := r.trm.Do(ctx, func(ctx context.Context) error {
		return upload(ctx, r.getDB(ctx))
	}); err != nil {
		return fmt.Errorf("user_kv_repo: failed to commit tx: %w", err)
	}
	return nil
}

func (r *KVPairRepo) Load(ctx context.Context) ([]entities.KVPair, error) {
	rows, err := r.getDB(ctx).QueryContext(ctx, `select key, value from user_kv;`)
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

func (r *KVPairRepo) getDB(ctx context.Context) trmsqlx.Tr {
	return r.getter.DefaultTrOrDB(ctx, r.db)
}
