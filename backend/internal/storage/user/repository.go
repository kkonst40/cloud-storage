package user

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/kkonst40/cloud-storage/backend/internal/domain"
	errs "github.com/kkonst40/cloud-storage/backend/internal/errors"
	"github.com/kkonst40/cloud-storage/backend/internal/storage"
)

const pkg = "UserRepository"

type Repository struct {
	db *sql.DB
}

func New(db *sql.DB) *Repository {
	return &Repository{
		db: db,
	}
}

func (r *Repository) GetById(ctx context.Context, userId int64) (domain.User, error) {
	const op = "GetById"
	const query = `
		SELECT id, username, password
		FROM users
		WHERE id = $1
	`

	var user domain.User

	err := r.db.QueryRowContext(ctx, query, userId).Scan(
		&user.ID,
		&user.Username,
		&user.Password,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.User{}, storage.ErrNotFound
		}
		return domain.User{}, errs.Wrap(pkg, op, err)
	}
	return user, nil
}

func (r *Repository) GetByName(ctx context.Context, name string) (domain.User, error) {
	const op = "GetByName"
	const query = `
		SELECT id, username, password
		FROM users
		WHERE username = $1
	`

	var user domain.User

	err := r.db.QueryRowContext(ctx, query, name).Scan(
		&user.ID,
		&user.Username,
		&user.Password,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.User{}, storage.ErrNotFound
		}
		return domain.User{}, errs.Wrap(pkg, op, err)
	}
	return user, nil
}

func (r *Repository) Create(ctx context.Context, user domain.User) (domain.User, error) {
	const op = "Create"
	const query = `
		INSERT INTO users (username, password)
		VALUES ($1, $2)
		RETURNING id
	`

	err := r.db.QueryRowContext(ctx, query, user.Username, user.Password).Scan(&user.ID)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return domain.User{}, storage.ErrDuplicate
		}
		return domain.User{}, errs.Wrap(pkg, op, err)
	}

	return user, nil
}
