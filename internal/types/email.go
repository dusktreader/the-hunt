package types

import (
	"net/mail"

	"github.com/dusktreader/the-hunt/internal/validator"
)

type Email string

func (e Email) Validate(v *validator.Validator) {
	v.Check(e != "", "email", "must be provided")
	_, err := mail.ParseAddress(string(e))
	v.Check(err == nil, "email", "must be a valid email address")
}
