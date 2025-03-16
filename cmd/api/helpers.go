package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/dusktreader/the-hunt/internal/data"
	"github.com/julienschmidt/httprouter"
)

func (app *application) logError(r *http.Request, er *data.ErrorPackage) {
	var logMessage string
	if er.LogMessage == "" {
		logMessage = fmt.Sprintf("Error encountered: %s", er.Message)
	} else {
		logMessage = er.LogMessage
	}
	app.logger.Error(
		logMessage,
		"error", er.Error,
		"method", r.Method,
		"uri", r.URL.RequestURI(),
	)

}

func (app *application) parseIdParam(r *http.Request) (uint64, error) {
	params := httprouter.ParamsFromContext(r.Context())
	id, err := strconv.ParseUint(params.ByName("id"), 10, 64)
	if err != nil {
		return 0, fmt.Errorf("UInt is required")
	} else if id == 0 {
		return 0, fmt.Errorf("0 is not allowed")
	}
	return id, nil
}

func (app *application) errorResponse(
	w http.ResponseWriter,
	r *http.Request,
	ep *data.ErrorPackage,
) {
	app.logError(r, ep)
	err := app.writeJSON(w, &data.JSONResponse{
		Data:			ep,
		StatusCode:		ep.StatusCode,
		EnvelopeKey:	"error",
	})
	if err != nil {
		app.logger.Error("Couldn't serialize error response", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func (app *application) notFoundResponse(w http.ResponseWriter, r *http.Request) {
	app.errorResponse(w, r, &data.ErrorPackage{
		StatusCode:	http.StatusNotFound,
		Message:	"Could not find the route you requested",
	})
}

func (app *application) notAllowedResponse(w http.ResponseWriter, r *http.Request) {
	app.errorResponse(w, r, &data.ErrorPackage{
		StatusCode:	http.StatusMethodNotAllowed,
		Message:	"The requested method is not allowed for this resource",
	})
}

func (app *application) badRequestResponse(w http.ResponseWriter, r *http.Request, err error) {
	app.errorResponse(w, r, &data.ErrorPackage{
		StatusCode:	http.StatusBadRequest,
		Message:	"Invalid request payload",
		Details:	err.Error(),
		Error:		err,
	})
}

func (app *application) failedValidationResponse(w http.ResponseWriter, r *http.Request, errors map[string]any) {
	errBytes, err := json.Marshal(errors)
	if err != nil {
		panic(err)
	}

	app.errorResponse(w, r, &data.ErrorPackage{
		StatusCode:	http.StatusUnprocessableEntity,
		Message:	"Invalid request payload",
		Details:	errors,
		Error:		fmt.Errorf("%s", errBytes),
	})
}

func (app *application) writeJSON(w http.ResponseWriter, jr *data.JSONResponse) error {
	key := "data"
	if jr.EnvelopeKey != "" {
		key = jr.EnvelopeKey
	}
	env := make(data.Envelope)
	env[key] = jr.Data

	var serialized []byte
	var err error
	if app.config.env == "development" {
		serialized, err = json.MarshalIndent(env, "", "  ")
	} else {
		serialized, err = json.Marshal(env)
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
