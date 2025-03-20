package data

import (
	"errors"
)

var (
	ErrRecordNotFound = errors.New("record not found")
	ErrEditConflict = errors.New("edit conflict")

	ErrInvalidParam = errors.New("invalid query parameter")
)

type ErrorMap map[error]error

func MapError(err error, errMap ErrorMap) error {
	if err == nil {
		return nil
	}

	for inErr, outErr := range errMap {
		if errors.Is(err, inErr) {
			return outErr
		}
	}
	return err
}
