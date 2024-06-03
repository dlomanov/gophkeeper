package repo

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	trmsqlx "github.com/avito-tech/go-transaction-manager/drivers/sqlx/v2"
	"github.com/dlomanov/gophkeeper/internal/entities"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"time"
)

type (
	EntryRepo struct {
		db     *sqlx.DB
		getter *trmsqlx.CtxGetter
	}
	entryRow struct {
		ID        uuid.UUID      `db:"id"`
		UserID    uuid.UUID      `db:"user_id"`
		Key       string         `db:"key"`
		Type      string         `db:"type"`
		Meta      sql.NullString `db:"meta"`
		Data      []byte         `db:"data"`
		Version   int64          `db:"version"`
		CreatedAt time.Time      `db:"created_at"`
		UpdatedAt time.Time      `db:"updated_at"`
	}
	entryVersionRow struct {
		ID      uuid.UUID `db:"id"`
		Version int64     `db:"version"`
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

func (r *EntryRepo) Get(ctx context.Context, userID uuid.UUID, id uuid.UUID) (*entities.Entry, error) {
	row := entryRow{}
	err := r.getDB(ctx).GetContext(ctx, &row, `
		SELECT id, user_id, key, type, meta, data, version, created_at, updated_at
		FROM entries
		WHERE id = $1 AND user_id = $2;`, id, userID)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return nil, fmt.Errorf("entry_repo: %w", entities.ErrEntryNotFound)
	case err != nil:
		return nil, fmt.Errorf("entry_repo: failed to get entry: %w", err)
	}
	return r.toEntity(row)
}

func (r *EntryRepo) GetAll(ctx context.Context, userID uuid.UUID) ([]entities.Entry, error) {
	var rows []entryRow
	err := r.getDB(ctx).SelectContext(ctx, &rows, `
		SELECT id, user_id, key, type, meta, data, version, created_at, updated_at
		FROM entries
		WHERE user_id = $1
		ORDER BY created_at;`, userID)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return nil, nil
	case err != nil:
		return nil, fmt.Errorf("entry_repo: failed to get entries: %w", err)
	}
	return r.toEntities(rows)
}

func (r *EntryRepo) GetVersions(ctx context.Context, userID uuid.UUID) ([]entities.EntryVersion, error) {
	var rows []entryVersionRow
	err := r.getDB(ctx).SelectContext(ctx, &rows, `
		SELECT id, version FROM entries WHERE user_id = $1;`, userID)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return nil, nil
	case err != nil:
		return nil, fmt.Errorf("entry_repo: failed to get entry versions: %w", err)
	}
	return r.toEntryVersions(rows)
}

func (r *EntryRepo) GetByIds(
	ctx context.Context,
	userID uuid.UUID,
	entryIds []uuid.UUID,
) ([]entities.Entry, error) {
	var rows []entryRow
	err := r.getDB(ctx).SelectContext(ctx, &rows, `
		SELECT id, user_id, key, type, meta, data, version, created_at, updated_at
		FROM entries
		WHERE user_id = $1 AND id = ANY($2)
		ORDER BY created_at;`, userID, pq.Array(entryIds))
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return nil, nil
	case err != nil:
		return nil, fmt.Errorf("entry_repo: failed to get entries: %w", err)
	}
	return r.toEntities(rows)
}

func (r *EntryRepo) Create(ctx context.Context, e *entities.Entry) error {
	row, err := r.toRow(e)
	if err != nil {
		return err
	}
	result, err := r.getDB(ctx).NamedExecContext(ctx, `
		INSERT INTO entries (id, user_id, key, type, meta, data, version, created_at, updated_at)
		VALUES (:id, :user_id, :key, :type, :meta, :data, :version, :created_at, :updated_at)
		ON CONFLICT DO NOTHING
	`, row)
	if err != nil {
		return fmt.Errorf("entry_repo: failed to create entry: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("entry_repo: failed to create entry: %w", err)
	}
	if affected == 0 {
		return fmt.Errorf("entry_repo: %w", entities.ErrEntryExists)
	}
	return nil
}

func (r *EntryRepo) Update(ctx context.Context, e *entities.Entry) error {
	row, err := r.toRow(e)
	if err != nil {
		return err
	}
	result, err := r.getDB(ctx).NamedExecContext(ctx, `
		UPDATE entries
		SET meta = :meta,
		    data = :data,
		    version = :version,
		    updated_at = :updated_at
		WHERE id = :id AND user_id = :user_id
	`, row)
	if err != nil {
		return fmt.Errorf("entry_repo: failed to update entry: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("entry_repo: failed to update entry: %w", err)
	}
	if affected == 0 {
		return fmt.Errorf("entry_repo: %w", entities.ErrEntryNotFound)
	}
	return nil
}

func (r *EntryRepo) Delete(ctx context.Context, userID uuid.UUID, id uuid.UUID) error {
	result, err := r.getDB(ctx).ExecContext(ctx, `
		DELETE FROM entries
		WHERE id = $1 AND user_id = $2;`, id, userID)
	if err != nil {
		return fmt.Errorf("entry_repo: failed to delete entry: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("entry_repo: failed to delete entry: %w", err)
	}
	if affected == 0 {
		return fmt.Errorf("entry_repo: %w", entities.ErrEntryNotFound)
	}
	return nil
}

func (r *EntryRepo) getDB(ctx context.Context) trmsqlx.Tr {
	return r.getter.DefaultTrOrDB(ctx, r.db)
}

func (*EntryRepo) toRow(e *entities.Entry) (entryRow, error) {
	if e == nil {
		return entryRow{}, fmt.Errorf("entry_repo: %w", entities.ErrEntryIsNil)
	}

	row := entryRow{
		ID:        e.ID,
		UserID:    e.UserID,
		Key:       e.Key,
		Type:      string(e.Type),
		Meta:      sql.NullString{},
		Data:      e.Data,
		Version:   e.Version,
		CreatedAt: e.CreatedAt,
		UpdatedAt: e.UpdatedAt,
	}

	if e.Meta != nil {
		meta, err := json.Marshal(e.Meta)
		if err != nil {
			return entryRow{}, fmt.Errorf("entry_repo: failed to marshal entry meta: %w", err)
		}
		row.Meta = sql.NullString{Valid: true, String: string(meta)}
	}

	return row, nil
}

func (r *EntryRepo) toEntities(rows []entryRow) ([]entities.Entry, error) {
	entries := make([]entities.Entry, 0, len(rows))
	for _, row := range rows {
		e, err := r.toEntity(row)
		if err != nil {
			return nil, err
		}
		entries = append(entries, *e)
	}
	return entries, nil
}

func (*EntryRepo) toEntity(row entryRow) (*entities.Entry, error) {
	entry := &entities.Entry{
		ID:        row.ID,
		UserID:    row.UserID,
		Key:       row.Key,
		Type:      "",
		Meta:      nil,
		Data:      row.Data,
		Version:   row.Version,
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
	}

	typ := entities.EntryType(row.Type)
	if !typ.Valid() {
		return nil, fmt.Errorf("entry_repo: invalid entry type: %s", row.Type)
	}
	entry.Type = typ

	if row.Meta.Valid {
		meta := make(map[string]string)
		if err := json.Unmarshal([]byte(row.Meta.String), &meta); err != nil {
			return nil, fmt.Errorf("entry_repo: failed to unmarshal entry meta: %w", err)
		}
		entry.Meta = meta
	}

	return entry, nil
}

func (r *EntryRepo) toEntryVersions(rows []entryVersionRow) ([]entities.EntryVersion, error) {
	versions := make([]entities.EntryVersion, 0, len(rows))
	for _, row := range rows {
		versions = append(versions, entities.EntryVersion{
			ID:      row.ID,
			Version: row.Version,
		})
	}
	return versions, nil
}
