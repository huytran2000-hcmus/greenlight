package data

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type TokenModel struct {
	DB *sql.DB
}

func (m TokenModel) New(scope string, userID int64, ttl time.Duration) (*Token, error) {
	token, err := generateToken(scope, userID, ttl)
	if err != nil {
		return nil, fmt.Errorf("data: generate a token: %s", err)
	}

	err = m.Insert(token)
	return token, err
}

func (m TokenModel) Insert(token *Token) error {
	query := `
    INSERT INTO tokens (hash, user_id, expiry, scope)
    VALUES ($1, $2, $3, $4)`

	args := []any{token.Hash, token.UserID, token.Expiry, token.Scope}

	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	_, err := m.DB.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("data: insert a token: %s", err)
	}

	return nil
}

func (m TokenModel) DeleteAllForUser(scope string, userID int64) error {
	query := `
    DELETE FROM tokens
    WHERE scope = $1 AND user_id = $2`

	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	_, err := m.DB.ExecContext(ctx, query, scope, userID)
	if err != nil {
		return fmt.Errorf("data: delete all token of scope %q for user with id=%d: %s", scope, userID, err)
	}

	return nil
}
