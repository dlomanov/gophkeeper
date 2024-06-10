package repo

import (
	"context"
	"database/sql"
	"errors"
	trmsqlx "github.com/avito-tech/go-transaction-manager/drivers/sqlx/v2"
	"github.com/dlomanov/gophkeeper/internal/apps/server/entities"
	"github.com/dlomanov/gophkeeper/internal/apps/server/usecases"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"time"
)

var _ usecases.UserRepo = (*UserRepo)(nil)

type (
	UserRepo struct {
		db     *sqlx.DB
		getter *trmsqlx.CtxGetter
	}
	userRow struct {
		ID        uuid.UUID `db:"id"`
		Login     string    `db:"login"`
		PassHash  string    `db:"pass_hash"`
		CreatedAt time.Time `db:"created_at"`
		UpdatedAt time.Time `db:"updated_at"`
	}
)

func NewUserRepo(
	db *sqlx.DB,
	getter *trmsqlx.CtxGetter,
) *UserRepo {
	return &UserRepo{
		db:     db,
		getter: getter,
	}
}

func (r *UserRepo) Exists(ctx context.Context, login entities.Login) (result bool, err error) {
	row := r.getDB(ctx).QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM users WHERE login = $1);`, login)
	if err = row.Err(); err != nil {
		return false, err
	}

	if err = row.Scan(&result); err != nil {
		return false, err
	}

	return result, nil
}

func (r *UserRepo) Get(ctx context.Context, login entities.Login) (user entities.User, err error) {
	db := r.getDB(ctx)
	row := userRow{}

	err = db.GetContext(ctx, &row, `SELECT id, login, pass_hash, created_at, updated_at FROM users WHERE login = $1;`, login)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return user, entities.ErrUserNotFound
		default:
			return user, err
		}
	}

	return r.toEntity(row), nil
}

func (r *UserRepo) Create(ctx context.Context, user entities.User) error {
	db := r.getDB(ctx)
	row := r.toRow(user)

	result, err := db.NamedExecContext(ctx, `
		INSERT INTO users (id, login, pass_hash, created_at, updated_at)
		VALUES (:id, :login, :pass_hash, :created_at, :updated_at)
		ON CONFLICT (id) DO NOTHING;`,
		row)
	if err != nil {
		return err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return entities.ErrUserExists
	}

	return nil
}

func (r *UserRepo) getDB(ctx context.Context) trmsqlx.Tr {
	return r.getter.DefaultTrOrDB(ctx, r.db)
}

func (*UserRepo) toRow(user entities.User) userRow {
	return userRow{
		ID:        user.ID,
		Login:     string(user.Login),
		PassHash:  string(user.PassHash),
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}
}

func (*UserRepo) toEntity(row userRow) entities.User {
	return entities.User{
		ID: row.ID,
		HashCreds: entities.HashCreds{
			Login:    entities.Login(row.Login),
			PassHash: entities.PassHash(row.PassHash),
		},
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
	}
}
