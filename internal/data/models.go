package data

import (
	"database/sql"
	"errors"
)

var (
	ErrRecordNotFound = errors.New("record not found")
)

type Models struct {
	Company CompanyModel
}

func NewModels(db *sql.DB) Models {
	return Models{
		Company: CompanyModel{DB: db},
	}
}
