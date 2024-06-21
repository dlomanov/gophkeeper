package repo

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	trmsqlx "github.com/avito-tech/go-transaction-manager/drivers/sqlx/v2"
	"github.com/dlomanov/gophkeeper/internal/apps/client/entities"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"time"
)

type (
	EntrySyncRepo struct {
		db     *sqlx.DB
		getter *trmsqlx.CtxGetter
	}
	entrySyncRow struct {
		ID        string `db:"id"`
		CreatedAt string `db:"created_at"`
	}
)

func NewEntrySyncRepo(
	db *sqlx.DB,
	getter *trmsqlx.CtxGetter,
) *EntrySyncRepo {
	return &EntrySyncRepo{
		db:     db,
		getter: getter,
	}
}

func (r *EntrySyncRepo) GetAll(ctx context.Context) ([]entities.EntrySync, error) {
	var rows []entrySyncRow
	err := r.getDB(ctx).SelectContext(ctx, &rows, `select id, created_at from entries_sync order by created_at;`)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return nil, nil
	case err != nil:
		return nil, fmt.Errorf("entry_sync_repo: failed to get entry syncs: %w", err)
	}
	res := make([]entities.EntrySync, len(rows))
	for i, row := range rows {
		createdAt, err := time.Parse(time.RFC3339, row.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("entry_sync_repo: failed to parse created_at: %w", err)
		}
		res[i] = entities.EntrySync{
			ID:        uuid.MustParse(row.ID),
			CreatedAt: createdAt,
		}
	}
	return res, nil
}

func (r *EntrySyncRepo) Create(ctx context.Context, entrySync entities.EntrySync) error {
	_, err := r.getDB(ctx).ExecContext(ctx, `
		insert into entries_sync (id, created_at)
		values (:id, :created_at)
		on conflict (id) do nothing ;`,
		entrySync.ID.String(), entrySync.CreatedAt.Format(time.RFC3339))
	if err != nil {
		return fmt.Errorf("entry_sync_repo: failed to create entry sync: %w", err)
	}
	return nil
}

func (r *EntrySyncRepo) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.getDB(ctx).ExecContext(ctx, `delete from entries_sync where id = $1 ;`, id.String())
	if err != nil {
		return fmt.Errorf("entry_sync_repo: failed to delete entry sync: %w", err)
	}
	return nil
}

func (r *EntrySyncRepo) getDB(ctx context.Context) trmsqlx.Tr {
	return r.getter.DefaultTrOrDB(ctx, r.db)
}
