package repo

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	trmsqlx "github.com/avito-tech/go-transaction-manager/drivers/sqlx/v2"
	"github.com/dlomanov/gophkeeper/internal/apps/client/entities"
	"github.com/dlomanov/gophkeeper/internal/core"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"time"
)

type (
	EntryRepo struct {
		db     *sqlx.DB
		getter *trmsqlx.CtxGetter
	}
	entryRow struct {
		ID            string         `db:"id"`
		Key           string         `db:"key"`
		Type          string         `db:"type"`
		Meta          sql.NullString `db:"meta"`
		Data          []byte         `db:"data"`
		GlobalVersion int64          `db:"global_version"`
		Version       int64          `db:"version"`
		CreatedAt     string         `db:"created_at"`
		UpdatedAt     string         `db:"updated_at"`
	}
	entryVersionRow struct {
		ID            string `db:"id"`
		GlobalVersion int64  `db:"global_version"`
	}
)

func NewEntryRepo(
	db *sqlx.DB,
	getter *trmsqlx.CtxGetter,
) *EntryRepo {
	return &EntryRepo{
		db:     db,
		getter: getter,
	}
}

func (r *EntryRepo) GetAll(ctx context.Context) ([]entities.Entry, error) {
	var rows []entryRow
	err := r.getDB(ctx).SelectContext(ctx, &rows, `
		SELECT id, key, type, meta, data, global_version, version, created_at, updated_at
		FROM entries
		ORDER BY created_at;`)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return nil, nil
	case err != nil:
		return nil, fmt.Errorf("entry_repo: failed to get entries: %w", err)
	}
	return r.toEntities(rows)
}

func (r *EntryRepo) Get(ctx context.Context, id uuid.UUID) (entities.Entry, error) {
	row := entryRow{}
	err := r.getDB(ctx).GetContext(ctx, &row, `
		SELECT id, key, type, meta, data, global_version, version, created_at, updated_at
		FROM entries
		WHERE id = $1;`, id)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return entities.Entry{}, fmt.Errorf("entry_repo: %w", entities.ErrEntryNotFound)
	case err != nil:
		return entities.Entry{}, fmt.Errorf("entry_repo: failed to get entry: %w", err)
	}
	return r.toEntry(row)
}

func (r *EntryRepo) GetVersions(ctx context.Context) ([]core.EntryVersion, error) {
	var rows []entryVersionRow
	err := r.getDB(ctx).SelectContext(ctx, &rows, `SELECT id, global_version FROM entries`)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return nil, nil
	case err != nil:
		return nil, fmt.Errorf("entry_repo: failed to get entries: %w", err)
	}
	result := make([]core.EntryVersion, len(rows))
	for i, row := range rows {
		result[i] = core.EntryVersion{
			ID:      uuid.MustParse(row.ID),
			Version: row.GlobalVersion,
		}
	}
	return result, nil
}

func (r *EntryRepo) Create(ctx context.Context, entry entities.Entry) error {
	row, err := r.toRow(entry)
	if err != nil {
		return fmt.Errorf("entry_repo: failed to map entry to row: %w", err)
	}
	res, err := r.getDB(ctx).NamedExecContext(ctx, `
		insert into entries (id, key, type, meta, data, global_version, version, created_at, updated_at)
		values (:id, :key, :type, :meta, :data, :global_version, :version, :created_at, :updated_at)
		on conflict do nothing;`,
		row)
	if err != nil {
		return fmt.Errorf("entry_repo: failed to insert entry to db: %w", err)
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("entry_repo: failed to get rows affected: %w", err)
	}
	if affected == 0 {
		return fmt.Errorf("entry_repo: %w", entities.ErrEntryExists)
	}
	return nil
}

func (r *EntryRepo) Update(ctx context.Context, entry entities.Entry) error {
	row, err := r.toRow(entry)
	if err != nil {
		return fmt.Errorf("entry_repo: failed to map entry to row: %w", err)
	}
	res, err := r.getDB(ctx).NamedExecContext(ctx, `
		update entries
		set meta = :meta,
		    data = :data,
		    global_version = :global_version,
		    version = :version,
		    updated_at = :updated_at
		where id = :id;`,
		row)
	if err != nil {
		return fmt.Errorf("entry_repo: failed to update entry: %w", err)
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("entry_repo: failed to get rows affected: %w", err)
	}
	if affected == 0 {
		return fmt.Errorf("entry_repo: %w", entities.ErrEntryNotFound)
	}
	return nil
}

func (r *EntryRepo) Delete(ctx context.Context, id uuid.UUID) error {
	res, err := r.getDB(ctx).ExecContext(ctx, `delete from entries where id = $1;`, id.String())
	if err != nil {
		return fmt.Errorf("entry_repo: failed to delete entry: %w", err)
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("entry_repo: failed to get rows affected: %w", err)
	}
	if affected == 0 {
		return fmt.Errorf("entry_repo: %w", entities.ErrEntryNotFound)
	}
	return err
}

func (r *EntryRepo) getDB(ctx context.Context) trmsqlx.Tr {
	return r.getter.DefaultTrOrDB(ctx, r.db)
}

func (r *EntryRepo) toEntities(rows []entryRow) ([]entities.Entry, error) {
	entries := make([]entities.Entry, len(rows))
	for i, row := range rows {
		entry, err := r.toEntry(row)
		if err != nil {
			return nil, err
		}
		entries[i] = entry
	}
	return entries, nil
}

func (r *EntryRepo) toEntry(row entryRow) (entry entities.Entry, err error) {
	entry.Key = row.Key
	entry.Data = row.Data
	entry.GlobalVersion = row.GlobalVersion
	entry.Version = row.Version
	entry.Type = core.EntryType(row.Type)
	entry.ID, err = uuid.Parse(row.ID)
	if err != nil {
		return entry, fmt.Errorf("entry_repo: invalid entry id: %s", row.ID)
	}
	if !entry.Type.Valid() {
		return entry, fmt.Errorf("entry_repo: invalid entry type: %s", row.Type)
	}
	entry.CreatedAt, err = time.Parse(time.RFC3339Nano, row.CreatedAt)
	if err != nil {
		return entry, fmt.Errorf("entry_repo: failed to parse created_at: %w", err)
	}
	entry.UpdatedAt, err = time.Parse(time.RFC3339Nano, row.UpdatedAt)
	if err != nil {
		return entry, fmt.Errorf("entry_repo: failed to parse updated_at: %w", err)
	}
	if row.Meta.Valid {
		err = json.Unmarshal([]byte(row.Meta.String), &entry.Meta)
		if err != nil {
			return entry, fmt.Errorf("entry_repo: failed to unmarshal meta: %w", err)
		}
	}
	return entry, nil
}

func (r *EntryRepo) toRow(entry entities.Entry) (row entryRow, err error) {
	row.ID = entry.ID.String()
	row.Key = entry.Key
	row.Type = string(entry.Type)
	row.Data = entry.Data
	row.GlobalVersion = entry.GlobalVersion
	row.Version = entry.Version
	row.CreatedAt = entry.CreatedAt.Format(time.RFC3339Nano)
	row.UpdatedAt = entry.UpdatedAt.Format(time.RFC3339Nano)
	if entry.Meta != nil {
		meta, err := json.Marshal(entry.Meta)
		if err != nil {
			return row, fmt.Errorf("entry_repo: failed to marshal meta: %w", err)
		}
		row.Meta = sql.NullString{Valid: true, String: string(meta)}
	}
	return row, nil
}
