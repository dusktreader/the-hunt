package types

import (
	"errors"
	"log/slog"

	"golang.org/x/crypto/bcrypt"

	"github.com/dusktreader/the-hunt/internal/validator"
)

type PlainPW string
type HashPW []byte

func (pp PlainPW) Validate(v *validator.Validator) {
	v.Check(pp != "", "password", "must be provided")
	v.Check(len(pp) >= 8, "password", "must be at least 8 bytes long")
	v.Check(len(pp) <= 128, "password", "must not be more than 128 bytes")
}

func NewHashPW(pp PlainPW) (HashPW, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(pp), 12)
	if err != nil {
		return nil, err
	}

	return hash, nil
}

func (hp HashPW) Compare(pp PlainPW) error {
	err := bcrypt.CompareHashAndPassword(hp, []byte(pp))
	if err != nil {
		switch {
		case errors.Is(err, bcrypt.ErrMismatchedHashAndPassword):
			return ErrPasswordMismatch
		default:
			slog.Debug("Error comparing plaintext password to hash", "err", err)
			return err
		}
	}
	return nil
}
