package main

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/dusktreader/the-hunt/internal/data"
	"github.com/dusktreader/the-hunt/internal/types"
	"github.com/dusktreader/the-hunt/internal/validator"
)

func (app *application) loginHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Email         types.Email   `json:"email"`
		PlainPassword types.PlainPW `json:"password"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	slog.Debug("Creating authentication token", "email", input.Email)

	v := validator.New()

	l := types.NewLogin(input.Email, input.PlainPassword)

	l.Validate(v)
	if !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors())
		return
	}

	var userID int64
	var isAdmin bool
	if app.config.APIEnv.IsDev() && l.Email == app.config.AdminEmail {
		if l.Password == app.config.AdminPassword {
			userID = types.AdminUserID
			isAdmin = true
		} else {
			app.serverErrorResponse(w, r, err, "Couldn't authenticate user")
			return
		}
	} else {
		u, err := app.models.User.GetForLogin(l)
		if err != nil {
			slog.Debug("Couldn't retrieve user", "email", l.Email, "error", err)
			switch {
			case errors.Is(err, types.ErrUserNotActivated):
				app.userNotActivatedResponse(w, r)
			case errors.Is(err, types.ErrUnauthorized):
				app.unauthorizedResponse(w, r)
			default:
				app.serverErrorResponse(w, r, err, "Couldn't authenticate user")
			}
			return
		}
		userID = u.ID
	}

	token, err := app.models.Token.New(
		userID,
		app.config.AuthTTL,
		types.ScopeAuthentication,
		isAdmin,
	)
	if err != nil {
		app.serverErrorResponse(w, r, err, "Couldn't authenticate user")
		return
	}

	err = app.writeJSON(w, &data.JSONResponse{
		Envelope:   data.Envelope{"auth": token},
		StatusCode: http.StatusCreated,
	})
	if err != nil {
		app.serverErrorResponse(w, r, err, "Failed to serialize data")
	}
}
