package data

import (
	"errors"

	"golang.org/x/crypto/bcrypt"
)


type Password struct {
	Plaintext	*string
	Hashed		[]byte
}

func NewPassword(plaintext string) (*Password, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(plaintext), 12)
	if err != nil {
		return nil, err
	}

	return &Password{
		Plaintext:	&plaintext,
		Hashed:		hash,
	}, nil
}

func (p Password) Matches(plaintext string) (bool, error) {
	err := bcrypt.CompareHashAndPassword(p.Hashed, []byte(plaintext))
	if err != nil {
		switch {
			case errors.Is(err, bcrypt.ErrMismatchedHashAndPassword):
				return false, nil
			default:
				return false, err
		}
	}
	return true, nil
}
