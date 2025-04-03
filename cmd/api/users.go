package main

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/dusktreader/the-hunt/internal/data"
	"github.com/dusktreader/the-hunt/internal/validator"
)

func (app *application) createUserHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: Can we use the partial struct here?
	var input struct {
		Name		string		`json:"name"`
		Email		string		`json:"email"`
		Password	string		`json:"password"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	slog.Debug("Creating a new user", "input", input)

	v := validator.New()

	u := &data.User{
		Name:		input.Name,
		Email:		input.Email,
		Activated:	false,
	}

	pw, err := data.NewPassword(input.Password)
	if err != nil {
		app.serverErrorResponse(w, r, err, "Failed to hash password")
		return
	}
	u.Password = *pw

	slog.Debug("Validating new user")

	u.Validate(v)
	if !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors())
		return
	}

	slog.Debug("Inserting new user into database")

	err = app.models.User.Insert(u)
	if err != nil {
		slog.Debug("Got an error on user insert", "err", err)
		switch {
			case errors.Is(err, data.ErrDuplicateKey):
				// TODO: We probably don't want to use this to avoid user enumeration
				app.duplicateKeyResponse(w, r)
			default:
				app.serverErrorResponse(w, r, err, "Couldn't add user")
		}
		return
	}

	t, err := app.models.Token.New(u.ID, app.config.ActivationTTL, "activation")
	if err != nil {
		slog.Debug("Got an error on token insert", "err", err)
		app.serverErrorResponse(w, r, err, "Couldn't create activation token")
		return
	}

	templateData := map[string]any{
		"user":		u,
		"token":	t,
	}

	slog.Debug("Starting mail sender go routine")
	app.background(app.mailer.Send, u.Email, "user_welcome.tmpl", templateData)

	slog.Debug("Serializing response")

	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/v1/users/%d", u.ID))

	err = app.writeJSON(w, &data.JSONResponse{
		Envelope: 		data.Envelope{"user": u},
		StatusCode:		http.StatusAccepted,
		Headers:		headers,
	})
	if err != nil {
		app.serverErrorResponse(w, r, err, "Failed to serialize user data")
		return
	}
}

func (app *application) activateUserHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Plaintext	string		`json:"token"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		slog.Debug("Error reading JSON", "error", err)
		app.badRequestResponse(w, r, err)
		return
	}

	slog.Debug("Activating a user", "token", input.Plaintext)

	id, err := app.models.User.Activate(input.Plaintext)
	if err != nil {
		switch {
			case errors.Is(err, data.ErrRecordNotFound):
				app.invalidTokenResponse(w, r)
			default:
				app.serverErrorResponse(w, r, err, "Couldn't activate user")
		}
	}
	slog.Debug("Activated user", "id", id)

	slog.Debug("Starting token cleanup go routine")
	app.background(app.models.User.DeleteTokensForUser, id, "activation")
}

func safeUnpack(from []any, to ...*any) error {
	if len(from) != len(to) {
		return fmt.Errorf("Args count didn't match. From count %d, to count %d", len(from), len(to))
	}
	for i, v := range from {
		*to[i] = v
	}
	return nil
}

func (app *application) cleanupActivations(args ...any) {
	(userID, scope) := safe



func (app *application) readUserHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.parseIdParam(r)
	if err != nil {
		app.badIdResponse(w, r, err)
		return
	}
	slog.Debug("Fetching user details", "id", id)

	u, err := app.models.User.GetOne(id)
	if err != nil {
		switch {
			case errors.Is(err, data.ErrRecordNotFound):
				app.notFoundResponse(w, r, id)
			default:
				app.serverErrorResponse(w, r, err, "Couldn't retrieve user")
		}
		return
	}
	slog.Debug("Retrieved user", "User", *u)

	err = app.writeJSON(w, &data.JSONResponse{
		Envelope: 		data.Envelope{"user": u},
		StatusCode:		http.StatusOK,
	})
	if err != nil {
		app.serverErrorResponse(w, r, err, "Failed to serialize user data")
	}
}

func (app *application) readManyUsersHandler(w http.ResponseWriter, r *http.Request) {
	slog.Debug("Fetching user list")

	qs := r.URL.Query()
	v := validator.New()
	filters := data.ParseFilters(
		qs,
		v,
		data.FilterConstraints{
			Search:		data.UserSearchFields.Check,
			Sort:		data.UserSortFields.Check,
		},
	)
	if !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors())
		return
	}

	slog.Debug("Retrieved filters", "filters", filters)

	users, metadata, err := app.models.User.GetMany(filters)
	if err != nil {
		app.serverErrorResponse(w, r, err, "Couldn't retrieve users")
	}
	slog.Debug("Fetched users", "metadata", metadata)

	err = app.writeJSON(w, &data.JSONResponse{
		StatusCode:		http.StatusOK,
		Envelope: 		data.Envelope{
			"users":	users,
			"metadata":	metadata,
		},
	})
	if err != nil {
		app.serverErrorResponse(w, r, err, "Failed to serialize user data")
	}
}

func (app *application) updateUserHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.parseIdParam(r)
	if err != nil {
		app.badIdResponse(w, r, err)
		return
	}

	slog.Debug("Updating user", "id", id)

	u, err := app.models.User.GetOne(id)
	if err != nil {
		switch {
			case errors.Is(err, data.ErrRecordNotFound):
				app.notFoundResponse(w, r, id)
			default:
				app.serverErrorResponse(w, r, err)
		}
		return
	}
	slog.Debug("Retrieved user", "User", *u)

	var input struct {
		Name		string		`json:"name"`
		Email		string		`json:"email"`
		Password	string		`json:"password"`
	}

	err = app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	slog.Debug("Updating user with request payload", "id", id, "input", input)

	u.Name = input.Name
	u.Email = input.Email

	const genericMessage = "Couldn't update user"


	pw, err := data.NewPassword(input.Password)
	if err != nil {
		app.serverErrorResponse(w, r, err, genericMessage)
		return
	}
	ok, err := pw.Matches(input.Password)
	if err != nil {
		app.serverErrorResponse(w, r, err, genericMessage)
		return
	}
	if ok {
		app.unchangedPasswordResponse(w, r)
		return
	}

	slog.Debug("Validating updated company", "id", id, "company", u)

	v := validator.New()
	u.Validate(v)
	if !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors())
		return
	}

	slog.Debug("Updating company in database")

	err = app.models.User.Update(u)
	if err != nil {
		switch {
			case errors.Is(err, data.ErrEditConflict):
				app.editConflictResponse(w, r)
			case errors.Is(err, data.ErrDuplicateKey):
				app.duplicateKeyResponse(w, r)
			default:
				app.serverErrorResponse(w, r, err, genericMessage)
		}
		return
	}

	slog.Debug("Serializing response")

	err = app.writeJSON(w, &data.JSONResponse{
		Envelope: 		data.Envelope{"company": u},
		StatusCode:		http.StatusOK,
	})
	if err != nil {
		app.serverErrorResponse(w, r, err, "Failed to serialize user data")
		return
	}
}

func (app *application) updatePartialUserHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.parseIdParam(r)
	if err != nil {
		app.badIdResponse(w, r, err)
		return
	}

	slog.Debug("Partially updating user", "id", id)

	version, err := app.models.User.GetVersion(id)
	if err != nil {
		switch {
			case errors.Is(err, data.ErrRecordNotFound):
				app.notFoundResponse(w, r, id)
			default:
				app.serverErrorResponse(w, r, err)
		}
		return
	}
	slog.Debug("Retrieved version", "Version", version)

	pc := data.PartialUser{}

	err = app.readJSON(w, r, &pc)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	slog.Debug("Updating user with request payload", "id", id, "input", pc)

	slog.Debug("Validating partial user", "id", id, "partial_user", pc)
	v := validator.New()
	pc.Validate(v)
	if !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors())
		return
	}

	slog.Debug("Updating user in database")
	c, err := app.models.User.PartialUpdate(id, version, &pc)
	if err != nil {
		switch {
			case errors.Is(err, data.ErrEditConflict):
				app.editConflictResponse(w, r)
			default:
				app.serverErrorResponse(w, r, err, "Couldn't update user")
		}
		return
	}

	slog.Debug("Serializing response")

	err = app.writeJSON(w, &data.JSONResponse{
		Envelope: 		data.Envelope{"user": c},
		StatusCode:		http.StatusOK,
	})
	if err != nil {
		app.serverErrorResponse(w, r, err, "Failed to serialize user data")
		return
	}
}

func (app *application) deleteUserHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.parseIdParam(r)
	if err != nil {
		app.badIdResponse(w, r, err)
		return
	}
	slog.Debug("Deleting user", "id", id)

	err = app.models.User.Delete(id)
	if err != nil {
		switch {
			case errors.Is(err, data.ErrRecordNotFound):
				app.notFoundResponse(w, r, id)
			default:
				app.serverErrorResponse(w, r, err, "Couldn't delete user")
		}
		return
	}
	slog.Debug("Deleted user", "id", id)

	err = app.writeJSON(w, &data.JSONResponse{
		Envelope: 		data.Envelope{"message": "User deleted successfully"},
		StatusCode:		http.StatusOK,
	})
	if err != nil {
		app.serverErrorResponse(w, r, err, "Failed to serialize response")
	}
}
