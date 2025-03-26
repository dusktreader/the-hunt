package types_test

import (
	"errors"
	"testing"

	"github.com/dusktreader/the-hunt/internal/types"
)

var (
	ErrTestJawa = errors.New("err::jawa")
	ErrTestEwok = errors.New("err::ewok")
	ErrTestHutt = errors.New("err::hutt")
	ErrTestPyke = errors.New("err::pyke")
)

func TestMapError(t *testing.T) {
	errMap := types.ErrorMap{
		ErrTestJawa: ErrTestEwok,
		ErrTestPyke: ErrTestHutt,
	}

	cases := []struct {
		name	string
		err		error
		wantErr	error
	}{
		{
			name:		"jawa mapped to ewok",
			err:		ErrTestJawa,
			wantErr:	ErrTestEwok,
		},
		{
			name:		"hutt not mapped",
			err:		ErrTestHutt,
			wantErr:	ErrTestHutt,
		},
		{
			name:		"pyke mapped to hutt",
			err:		ErrTestPyke,
			wantErr:	ErrTestHutt,
		},
		{
			name:		"nil argument",
			err:		nil,
			wantErr:	nil,
		},
	}
	for _, c := range cases {
		gotErr := types.MapError(c.err, errMap)
		if (gotErr == nil && c.wantErr != nil) || (!errors.Is(gotErr, c.wantErr)) {
			t.Errorf("%s: expected %v for %v, got %v", c.name, c.wantErr, c.err, gotErr)
		}
	}
}
