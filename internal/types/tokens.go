package types

import (
	"crypto/rand"
	"crypto/sha256"
	"time"

	"github.com/dusktreader/the-hunt/internal/validator"
)

type TokenScope string

const ScopeActivation TokenScope = "activation"
const ScopeAuthentication TokenScope = "authentication"

type PlainToken string

type Token struct {
	Plaintext 	PlainToken	`json:"token"`
	Hash		[]byte		`json:"-"`
	UserID		int64		`json:"-"`
	ExpiresAt	time.Time	`json:"expires_at"`
	Scope		TokenScope	`json:"-"`
	IsAdmin		bool		`json:"-"`
}

func Hash(plaintext string) []byte {
	hash := sha256.Sum256([]byte(plaintext))
	return hash[:]
}

func GenerateToken(userID int64, ttl time.Duration, scope TokenScope, isAdmin  bool) *Token {
	plaintext := rand.Text()
	return &Token{
		Plaintext:	PlainToken(plaintext),
		Hash:		Hash(plaintext),
		UserID:		userID,
		ExpiresAt:	time.Now().Add(ttl),
		Scope:		scope,
		IsAdmin:	isAdmin,
	}
}

func (pt PlainToken) Validate(v *validator.Validator) {
	v.Check(pt != "", "token", "must be provided")
	v.Check(len(pt) == 26, "token", "must be exactly 26 bytes")
}

