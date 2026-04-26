package user

import (
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

func NewRepository(db *sql.DB) *Repository {
	return &Repository{
		db: db,
	}
}

func (r *Repository) Create(user domain.User) (domain.User, error) {
	const op = "Create"
	const query = `
		INSERT INTO users (username, password)
		VALUES ($1, $2)
		RETURNING id
	`

	err := r.db.QueryRow(query, user.Username, user.Password).Scan(&user.ID)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return domain.User{}, storage.ErrDuplicate
		}
		return domain.User{}, errs.Wrap(pkg, op, err)
	}

	return user, nil
}

func (r *Repository) IsExistsByID(id int64) (bool, error) {
	const op = "IsExistsByID"
	const query = `
		SELECT EXISTS(
			SELECT 1
			FROM users
			WHERE id = $1
		)
	`

	var exists bool

	err := r.db.QueryRow(query, id).Scan(&exists)
	if err != nil {
		return false, errs.Wrap(pkg, op, err)
	}

	return exists, nil
}

func (r *Repository) IsExistsByName(name string) (bool, error) {
	const op = "IsExistsByName"
	const query = `
		SELECT EXISTS(
			SELECT 1
			FROM users
			WHERE username = $1
		)
	`

	var exists bool

	err := r.db.QueryRow(query, name).Scan(&exists)
	if err != nil {
		return false, errs.Wrap(pkg, op, err)
	}

	return exists, nil
}

func (r *Repository) ByName(name string) (domain.User, error) {
	const op = "ByEmail"
	const query = `
		SELECT id, username, password
		FROM users
		WHERE username = $1
	`

	var user domain.User

	err := r.db.QueryRow(query, name).Scan(
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

func (r *Repository) ById(userId int64) (domain.User, error) {
	const op = "ById"
	const query = `
		SELECT id, username, password
		FROM users
		WHERE id = $1
	`

	var user domain.User

	err := r.db.QueryRow(query, userId).Scan(
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
