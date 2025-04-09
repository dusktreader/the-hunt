package types

import (
	"errors"
	"fmt"
	"log/slog"
	"regexp"
)

var (
	ErrRecordNotFound   = errors.New("record not found")
	ErrNoTokenMatch     = errors.New("no valid token")
	ErrEditConflict     = errors.New("edit conflict")
	ErrInvalidParam     = errors.New("invalid query parameter")
	ErrDuplicateKey     = errors.New("duplicate key")
	ErrUnknown          = errors.New("unknown error")
	ErrUnauthorized     = errors.New("unauthorized")
	ErrForbidden        = errors.New("forbidden")
	ErrPasswordMismatch = errors.New("password mismatch")
	ErrUserNotActivated = errors.New("user not activated")
)

type ErrorMapUnion any

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
	return err
}
