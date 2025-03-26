package types

import (
	"github.com/dusktreader/the-hunt/internal/validator"
)

type Login struct {
	Email		Email
	Password	PlainPW
}

func NewLogin(email Email, password PlainPW) *Login {
	return &Login{
		Email:		email,
		Password:	password,
	}
}

func (l *Login) Validate(v *validator.Validator) {
	l.Email.Validate(v)
}
