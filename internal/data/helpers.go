package data

import (
	"errors"
	"net/http"

	"golang.org/x/crypto/bcrypt"
)

type JSONResponse struct {
	Envelope	Envelope
	StatusCode	int
	Headers		http.Header
}

type Envelope map[string]any

type ErrorPackage struct {
	Error		error	`json:"-"`
	Message		string	`json:"message"`
	LogMessage	string	`json:"-"`
	Details		any		`json:"details,omitempty"`
	StatusCode	int		`json:"-"`
}

type ListMetadata struct {
	CurrentPage		int		`json:"current_page"`
	PageSize		int		`json:"page_size"`
	FirstPage		int		`json:"first_page"`
	LastPage		int		`json:"last_page"`
	RecordCount		int		`json:"record_count"`
}

func NewListMetadata(f Filters, recordCount int) ListMetadata {
	if recordCount == 0 {
		return ListMetadata{}
	}

	return ListMetadata{
		CurrentPage:	*f.Page,
		PageSize:		*f.PageSize,
		FirstPage:		1,
		LastPage:		(recordCount + *f.PageSize - 1) / *f.PageSize,
		RecordCount:	recordCount,
	}
}

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
