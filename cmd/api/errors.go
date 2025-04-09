package main

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/dusktreader/the-hunt/internal/data"
	"github.com/dusktreader/the-hunt/internal/types"
)

func (app *application) errorResponse(
	w http.ResponseWriter,
	r *http.Request,
	ep *data.ErrorPackage,
) {
	app.logError(r, ep)
	err := app.writeJSON(w, &data.JSONResponse{
		Envelope:   data.Envelope{"error": ep},
		StatusCode: ep.StatusCode,
	})
	if err != nil {
		slog.Error("Couldn't serialize error response", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func (app *application) serverErrorResponse(w http.ResponseWriter, r *http.Request, err error, message ...string) {
	msg := "There was an error processing your request"
	if len(message) > 0 {
		msg = message[0]
	}
	app.errorResponse(w, r, &data.ErrorPackage{
		StatusCode: http.StatusInternalServerError,
		Message:    msg,
		Error:      err,
	})
}

func (app *application) badIdResponse(w http.ResponseWriter, r *http.Request, err error) {
	app.errorResponse(w, r, &data.ErrorPackage{
		StatusCode: http.StatusBadRequest,
		Message:    "Invalid ID provided",
		Details:    err.Error(),
		Error:      err,
	})
}

func (app *application) duplicateKeyResponse(w http.ResponseWriter, r *http.Request) {
	app.errorResponse(w, r, &data.ErrorPackage{
		StatusCode: http.StatusBadRequest,
		Message:    "Duplicate key provided",
	})
}

func (app *application) notFoundResponse(w http.ResponseWriter, r *http.Request, lookupKey ...any) {
	var details string
	if len(lookupKey) > 0 {
		details = fmt.Sprintf("No match for key %v", lookupKey[0])
	}

	app.errorResponse(w, r, &data.ErrorPackage{
		StatusCode: http.StatusNotFound,
		Message:    "Could not find the record you requested",
		Details:    details,
	})
}

func (app *application) routeNotFoundResponse(w http.ResponseWriter, r *http.Request) {
	app.errorResponse(w, r, &data.ErrorPackage{
		StatusCode: http.StatusNotFound,
		Message:    "Could not find the route you requested",
	})
}

func (app *application) notAllowedResponse(w http.ResponseWriter, r *http.Request) {
	app.errorResponse(w, r, &data.ErrorPackage{
		StatusCode: http.StatusMethodNotAllowed,
		Message:    "The requested method is not allowed for this resource",
	})
}

func (app *application) badRequestResponse(w http.ResponseWriter, r *http.Request, err error) {
	app.errorResponse(w, r, &data.ErrorPackage{
		StatusCode: http.StatusBadRequest,
		Message:    "Invalid request payload",
		Details:    err.Error(),
		Error:      err,
	})
}

func (app *application) failedValidationResponse(w http.ResponseWriter, r *http.Request, errors map[string]any) {
	errBytes, err := json.Marshal(errors)
	if err != nil {
		panic(err)
	}

	app.errorResponse(w, r, &data.ErrorPackage{
		StatusCode: http.StatusUnprocessableEntity,
		Message:    "Invalid request payload",
		Details:    errors,
		Error:      fmt.Errorf("%s", errBytes),
	})
}

func (app *application) editConflictResponse(w http.ResponseWriter, r *http.Request) {
	app.errorResponse(w, r, &data.ErrorPackage{
		StatusCode: http.StatusConflict,
		Message:    "Unable to update the record due to an edit conflict. Please try again",
	})
}

func (app *application) rateLimitExceededResponse(w http.ResponseWriter, r *http.Request) {
	app.errorResponse(w, r, &data.ErrorPackage{
		StatusCode: http.StatusTooManyRequests,
		Message:    "Rate limit exceeded",
	})
}

func (app *application) userNotActivatedResponse(w http.ResponseWriter, r *http.Request) {
	app.errorResponse(w, r, &data.ErrorPackage{
		StatusCode: http.StatusUnauthorized,
		Message:    "User is not activated yet. Please check your email for the activation link",
	})
}

func (app *application) unauthorizedResponse(w http.ResponseWriter, r *http.Request) {
	app.errorResponse(w, r, &data.ErrorPackage{
		StatusCode: http.StatusUnauthorized,
		Message:    "Unauthorized. Please try logging in again",
	})
}

func (app *application) forbiddenResponse(w http.ResponseWriter, r *http.Request) {
	app.errorResponse(w, r, &data.ErrorPackage{
		StatusCode: http.StatusForbidden,
		Message:    "Forbidden. Please check your permissions",
	})
}

func (app *application) invalidTokenResponse(w http.ResponseWriter, r *http.Request, scope types.TokenScope) {
	if scope == types.ScopeAuthentication {
		w.Header().Set("WWW-Authenticate", "Bearer")
	}
	app.errorResponse(w, r, &data.ErrorPackage{
		StatusCode: http.StatusUnauthorized,
		Message:    fmt.Sprintf("Invalid token; please request new %s token", scope),
	})
}
