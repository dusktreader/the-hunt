package main

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/dusktreader/the-hunt/internal/data"
)


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
		StatusCode:	http.StatusInternalServerError,
		Message:	msg,
		Error:		err,
	})
}

func (app *application) badIdResponse(w http.ResponseWriter, r *http.Request, err error) {
	app.errorResponse(w, r, &data.ErrorPackage{
		StatusCode:	http.StatusBadRequest,
		Message:	"Invalid ID provided",
		Details:	err.Error(),
		Error:		err,
	})
}

func (app *application) notFoundResponse(w http.ResponseWriter, r *http.Request, lookupKey ...any) {
	var details string
	if len(lookupKey) > 0 {
		details = fmt.Sprintf("No match for key %v", lookupKey[0])
	}

	app.errorResponse(w, r, &data.ErrorPackage{
		StatusCode:	http.StatusNotFound,
		Message:	"Could not find the record you requested",
		Details:	details,
	})
}

func (app *application) routeNotFoundResponse(w http.ResponseWriter, r *http.Request) {
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

func (app *application) editConflictResponse(w http.ResponseWriter, r *http.Request) {
	app.errorResponse(w, r, &data.ErrorPackage{
		StatusCode:	http.StatusConflict,
		Message:	"Unable to update the record due to an edit conflict; please try again",
	})
}
