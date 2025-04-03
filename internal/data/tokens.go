package data

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"time"

	"github.com/dusktreader/the-hunt/internal/validator"
)

const ScopeActivation = "actiation"

type Token struct {
	Plaintext 	string
	Hash		[]byte
	UserID		int64
	ExpiresAt	time.Time
	Scope		string
}

func Hash(plaintext string) []byte {
	hash := sha256.Sum256([]byte(plaintext))
	return hash[:]
}

func generateToken(userID int64, ttl time.Duration, scope string) *Token {
	plaintext := rand.Text()
	return &Token{
		Plaintext:	plaintext,
		Hash:		Hash(plaintext),
		UserID:		userID,
		ExpiresAt:	time.Now().Add(ttl),
		Scope:		scope,
	}
}

type TokenModel struct {
	DB *sql.DB
	CFG ModelConfig
}

func (t *Token) Validate(v *validator.Validator) {
	v.Check(t.Plaintext != "", "plaintext", "must be provided")
	v.Check(len(t.Plaintext) != 26, "plaintext", "must not be exactly 26 bytes")
}

func (m TokenModel) New(userID int64, ttl time.Duration, scope string) (*Token, error) {
	token := generateToken(userID, ttl, scope)
	err := m.Insert(token)
	return token, err
}

func (m TokenModel) Insert(t *Token) error {
	query := `
		insert into tokens (hash, user_id, expires_at, scope)
		values ($1, $2, $3, $4)
	`
	args := []any{
		t.Hash,
		t.UserID,
		t.ExpiresAt,
		t.Scope,
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
