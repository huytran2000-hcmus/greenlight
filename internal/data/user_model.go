package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/lib/pq"
)

var ErrDuplicateEmail = errors.New("duplicate email")

type UserModel struct {
	DB *sql.DB
}

func (m UserModel) Insert(user *User) error {
	err := user.Password.Hash()
	if err != nil {
		return fmt.Errorf("data: hash a user password: %s", err)
	}
	query := `
    INSERT INTO users (name, email, password_hash, activated)
    VALUES ($1, $2, $3, $4)
    RETURNING id, created_at, version`

	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	args := []any{user.Name, user.Email, user.Password.hash, user.Activated}
	err = m.DB.QueryRowContext(ctx, query, args...).Scan(&user.ID, &user.CreatedAt, &user.Version)
	if err != nil {
		switch {
		case isUniqueEmailConstrainstError(err):
			return ErrDuplicateEmail
		default:
			return fmt.Errorf("data: insert a user: %s", err)
		}
	}

	return nil
}

func (m UserModel) GetByEmail(email string) (*User, error) {
	query := `
    SELECT id, email, password_hash, name, created_at, activated, version
    FROM users
    WHERE email = $1`

	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	var user User
	err := m.DB.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.Email,
		&user.Password.hash,
		&user.Name,
		&user.CreatedAt,
		&user.Activated,
		&user.Version,
	)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, fmt.Errorf("data: select a user: %s", err)
		}
	}

	return &user, nil
}

func (m UserModel) Update(user *User) error {
	err := user.Password.Hash()
	if err != nil {
		return fmt.Errorf("data: hash a user password: %s", err)
	}
	query := `
    UPDATE users
    SET email = $1, password_hash = $2, name = $3, activated = $4, version = version + 1
    WHEREE id = $5 AND version = $6
    RETURNING version
    `

	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	args := []any{user.Email, user.Password.hash, user.Name, user.Activated, user.ID, user.Version}
	err = m.DB.QueryRowContext(ctx, query, args).Scan(&user.Version)
	if err != nil {
		switch {
		case isUniqueEmailConstrainstError(err):
			return ErrDuplicateEmail
		case errors.Is(err, sql.ErrNoRows):
			return ErrEditConflict
		default:
			return fmt.Errorf("data: update a user: %s", err)
		}
	}

	return nil
}

func isUniqueEmailConstrainstError(err error) bool {
	var pgErr *pq.Error
	ok := errors.As(err, &pgErr)
	if !ok {
		return false
	}
	return pgErr.Code == "23505" && pgErr.Constraint == "users_email_key"
}
