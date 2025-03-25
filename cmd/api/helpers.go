package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strconv"

	"github.com/dusktreader/the-hunt/internal/data"
	"github.com/julienschmidt/httprouter"
)

func MaybeDie(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, "There was an error:", err)
		os.Exit(1)
	}
}

func Die(msg string, flags ...interface{}) {
	msg = fmt.Sprintf(msg, flags...)
	fmt.Fprintln(os.Stderr, "Aborting:", msg)
	os.Exit(1)
}

func (app *application) logError(r *http.Request, er *data.ErrorPackage) {
	var logMessage string
	if er.LogMessage == "" {
		logMessage = fmt.Sprintf("Error encountered: %s", er.Message)
	} else {
		logMessage = er.LogMessage
	}
	slog.Error(
		logMessage,
		"error", er.Error,
		"method", r.Method,
		"uri", r.URL.RequestURI(),
	)

}

func (app *application) parseIdParam(r *http.Request) (int64, error) {
	params := httprouter.ParamsFromContext(r.Context())
	id, err := strconv.ParseInt(params.ByName("id"), 10, 64)
	if err != nil {
		return 0, fmt.Errorf("Int is required")
	} else if id < 0 {
		return 0, fmt.Errorf("Negative ids are not allowed")
	} else if id == 0 {
		return 0, fmt.Errorf("0 is not allowed")
	}
	return id, nil
}

func (app *application) writeJSON(w http.ResponseWriter, jr *data.JSONResponse) error {
	var serialized []byte
	var err error
	if app.config.APIEnv == "development" {
		serialized, err = json.MarshalIndent(jr.Envelope, "", "  ")
	} else {
		serialized, err = json.Marshal(jr.Envelope)
	}
	if err != nil {
		return fmt.Errorf("Failed to serialize response data: %w", err)
	}
	serialized = append(serialized, '\n')

	w.Header().Set("Content-Type", "application/json")
	for key, val := range jr.Headers {
		w.Header()[key] = val
	}

	w.WriteHeader(jr.StatusCode)
	w.Write(serialized)
	return nil
}

func (app *application) readJSON(w http.ResponseWriter, r *http.Request, dst any) error {
	r.Body = http.MaxBytesReader(w, r.Body, 1_048_576)

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	err := dec.Decode(dst)
	if err != nil {

		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError
		var invalidUnmarshalError *json.InvalidUnmarshalError

		switch {
			case errors.As(err, &syntaxError):
				return fmt.Errorf("Body contains badly-formed JSON (at character %d)", syntaxError.Offset)

			case errors.Is(err, io.ErrUnexpectedEOF):
				return fmt.Errorf("Body contains badly-formed JSON")

			case errors.As(err, &unmarshalTypeError):
				if unmarshalTypeError.Field != "" {
					return fmt.Errorf("Body contains an incorrect JSON type for the %q field", unmarshalTypeError.Field)
				}
				return fmt.Errorf("Body contains an incorrect JSON type (at character %d)", unmarshalTypeError.Offset)

			case errors.Is(err, io.EOF):
				return fmt.Errorf("Body must not be empty")

			case errors.As(err, &invalidUnmarshalError):
				panic(err)

			default:
				return err
		}
	}

	err = dec.Decode(&struct{}{})
	if !errors.Is(err, io.EOF) {
		return fmt.Errorf("Body must only contain a single JSON value")
	}
	return nil
}
