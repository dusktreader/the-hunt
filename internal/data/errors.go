package data

import (
	"errors"
	"fmt"
	"log/slog"
	"regexp"
)

var (
	ErrRecordNotFound = errors.New("record not found")
	ErrEditConflict = errors.New("edit conflict")
	ErrInvalidParam = errors.New("invalid query parameter")
	ErrDuplicateKey = errors.New("duplicate key")
	ErrUnknown = errors.New("unknown error")
)

type ErrorMapUnion interface {}

type ErrorMap map[ErrorMapUnion]error

func MapError(err error, errMap ErrorMap) error {
	if err == nil {
		return nil
	}

	slog.Debug("Attempting to map original error", "err", err)
	for inErr, outErr := range errMap {
		switch v := inErr.(type) {
			case error:
				if errors.Is(err, v) {
					return outErr
				}
			case string:
				rex := regexp.MustCompile(v)
				if rex.Match([]byte(err.Error())) {
					return outErr
				}
			default:
				panic(fmt.Sprintf("Invalid type in ErrorMap: %v. Expected error or string", inErr))
		}
	}

	slog.Warn("Failed to map error", "err", err)
	return ErrUnknown
}
