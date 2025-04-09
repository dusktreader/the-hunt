package data

import (
	"context"
	"database/sql"
	"time"

	"github.com/dusktreader/the-hunt/internal/types"
)

type TokenModel struct {
	DB  *sql.DB
	CFG ModelConfig
}

func (m TokenModel) New(userID int64, ttl time.Duration, scope types.TokenScope, isAdmin bool) (*types.Token, error) {
	token := types.GenerateToken(userID, ttl, scope, isAdmin)
	err := m.Insert(token)
	return token, err
}

func (m TokenModel) Insert(t *types.Token) error {
	var id any
	if !t.IsAdmin {
		id = any(t.UserID)
	}

	query := `
		insert into tokens (hash, user_id, expires_at, scope, is_admin)
		values ($1, $2, $3, $4, $5)
	`
	args := []any{
		t.Hash,
		id,
		t.ExpiresAt,
		t.Scope,
		t.IsAdmin,
	}

	ctx, cancel := context.WithTimeout(context.Background(), m.CFG.QueryTimeout)
	defer cancel()

	_, err := m.DB.ExecContext(
		ctx,
		query,
		args...,
	)
	return err
}

func (m TokenModel) GetOne(pt types.PlainToken, scope types.TokenScope) (*types.Token, error) {
	query := `
		select coalesce(user_id, 0), expires_at, scope, is_admin
		from tokens
		where hash = $1
		and scope = $2
		and expires_at > $3
	`
	args := []any{
		types.Hash(string(pt)),
		scope,
		time.Now(),
	}

	var t types.Token

	ctx, cancel := context.WithTimeout(context.Background(), m.CFG.QueryTimeout)
	defer cancel()

	return &t, types.MapError(
		m.DB.QueryRowContext(ctx, query, args...).Scan(
			&t.UserID,
			&t.ExpiresAt,
			&t.Scope,
			&t.IsAdmin,
		),
		types.ErrorMap{sql.ErrNoRows: types.ErrRecordNotFound},
	)
}

func (m TokenModel) DeleteForUser(scope string, userID int64) error {
	query := `
		delete from tokens
		where user_id = $1 and scope = $2
	`

	ctx, cancel := context.WithTimeout(context.Background(), m.CFG.QueryTimeout)
	defer cancel()
	_, err := m.DB.ExecContext(ctx, query, userID, scope)
	return err
}
