package types

import (
	"time"

	"github.com/dusktreader/the-hunt/internal/validator"
)

type User struct {
	ID				int64		`json:"id"`
	CreatedAt		time.Time	`json:"created_at"`
	UpdatedAt		time.Time	`json:"updated_at"`
	Name			string		`json:"name"`
	Email			Email		`json:"email"`
	PlainPassword	PlainPW		`json:"-"`
	HashedPassword	HashPW		`json:"-"`
	Activated		bool		`json:"activated"`
	Version			int64		`json:"version"`
}

type PartialUser struct {
	Name			*string		`json:"name"`
	Email			*Email		`json:"email"`
	PlainPassword	*PlainPW	`json:"-"`
	HashedPassword	*HashPW		`json:"-"`
}

var AnonymousUser = &User{}
const AdminUserID = 0

func (u *User) Validate(v *validator.Validator) {
	v.Check(u.Name != "", "name", "must be provided")
	v.Check(len(u.Name) <= 128, "name", "must not be more than 128 bytes")

	u.Email.Validate(v)

	if len(u.PlainPassword) > 0 {
		u.PlainPassword.Validate(v)
	}
}

func (u *User) IsAnonymous() bool {
	return u == AnonymousUser
}

func (uc *PartialUser) Validate(v *validator.Validator) {
	if uc.Name != nil {
		v.Check(*uc.Name != "", "name", "must be provided")
		v.Check(len(*uc.Name) <= 128, "name", "must not be more than 128 bytes")
	}

	if uc.Email != nil {
		uc.Email.Validate(v)
	}

	if uc.PlainPassword != nil {
		uc.PlainPassword.Validate(v)
	}
}
