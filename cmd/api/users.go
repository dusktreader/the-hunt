package main

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/dusktreader/the-hunt/internal/data"
	"github.com/dusktreader/the-hunt/internal/types"
	"github.com/dusktreader/the-hunt/internal/validator"
)

func (app *application) createUserHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: Can we use the partial struct here?
	var input struct {
		Name		string			`json:"name"`
		Email		types.Email		`json:"email"`
		Password	types.PlainPW	`json:"password"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	slog.Debug("Creating a new user", "input", input)

	v := validator.New()

	u := &types.User{
		Name:			input.Name,
		Email:			input.Email,
		Activated:		false,
		PlainPassword:	input.Password,
	}

	slog.Debug("Validating new user")

	u.Validate(v)
	if !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors())
		return
	}

	hp, err := types.NewHashPW(input.Password)
	if err != nil {
		app.serverErrorResponse(w, r, err, "Failed to hash password")
		return
	}
	u.HashedPassword = hp

	slog.Debug("Inserting new user into database")

	err = app.models.User.Insert(u)
	if err != nil {
		slog.Debug("Got an error on user insert", "err", err)
		switch {
			case errors.Is(err, types.ErrDuplicateKey):
				// TODO: We probably don't want to use this to avoid user enumeration
				app.duplicateKeyResponse(w, r)
			default:
				app.serverErrorResponse(w, r, err, "Couldn't add user")
		}
		return
	}

	t, err := app.models.Token.New(u.ID, app.config.ActivationTTL, types.ScopeActivation, false)
	if err != nil {
		slog.Debug("Got an error on token insert", "err", err)
		app.serverErrorResponse(w, r, err, "Couldn't create activation token")
		return
	}

	err = app.models.Permission.AddForUser(u.ID, types.CompanyRead, types.CompanyWrite)
	if err != nil {
		slog.Debug("Got an error from adding user permissions", "err", err)
		app.serverErrorResponse(w, r, err, "Couldn't grant default permissions to user")
		return
	}

	templateData := map[string]any{
		"user":		u,
		"token":	t,
	}

	slog.Debug("Starting mail sender go routine")
	close := func() error { return app.mailer.Send(u.Email, "user_welcome.tmpl", templateData) }
	app.background(close)

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
	slog.Debug("Processing request for user activation")

	var input struct {
		PlainToken types.PlainToken `json:"token"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	slog.Debug("Validating JSON payload")
	v := validator.New()
	input.PlainToken.Validate(v)
	if !v.Valid() {
		slog.Debug("Token validation failed", "errors", v.Errors())
		app.invalidTokenResponse(w, r, types.ScopeActivation)
		return
	}

	slog.Debug("Looking up token in database")
	t, err := app.models.Token.GetOne(input.PlainToken, types.ScopeActivation)
	if err != nil {
		switch {
			case errors.Is(err, types.ErrRecordNotFound):
				app.invalidTokenResponse(w, r, types.ScopeActivation)
			default:
				app.serverErrorResponse(w, r, err, "Couldn't parse activation token")
		}
		return
	}

	slog.Debug("Activating user", "id", t.UserID)
	id, err := app.models.User.Activate(*t)
	if err != nil {
		switch {
			case errors.Is(err, types.ErrRecordNotFound):
				app.invalidTokenResponse(w, r, types.ScopeActivation)
			default:
				app.serverErrorResponse(w, r, err, "Couldn't activate user")
		}
		return
	}
	slog.Debug("Activated user", "id", id)

	slog.Debug("Starting token cleanup go routine")
	close := func() error { return app.models.User.DeleteTokensForUser(id, "activation") }
	app.background(close)

	err = app.writeJSON(w, &data.JSONResponse{
		Envelope: 		data.Envelope{"message": fmt.Sprintf("Activated user %d", id)},
		StatusCode:		http.StatusOK,
	})
	if err != nil {
		app.serverErrorResponse(w, r, err, "Failed to serialize response")
		return
	}
}

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
			case errors.Is(err, types.ErrRecordNotFound):
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
			case errors.Is(err, types.ErrRecordNotFound):
				app.notFoundResponse(w, r, id)
			default:
				app.serverErrorResponse(w, r, err)
		}
		return
	}
	slog.Debug("Retrieved user", "User", *u)

	var input struct {
		Name		string			`json:"name"`
		Email		types.Email		`json:"email"`
		Password	types.PlainPW	`json:"password"`
	}

	err = app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	slog.Debug("Updating user with request payload", "id", id, "input", input)

	u.Name = input.Name
	u.Email = input.Email


	slog.Debug("Validating updated company", "id", id, "company", u)

	v := validator.New()
	u.Validate(v)
	if !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors())
		return
	}

	const genericMessage = "Couldn't update user"

	slog.Debug("Hashing password")
	hp, err := types.NewHashPW(input.Password)
	if err != nil {
		app.serverErrorResponse(w, r, err, genericMessage)
		return
	}
	u.HashedPassword = hp

	slog.Debug("Updating company in database")

	err = app.models.User.Update(u)
	if err != nil {
		switch {
			case errors.Is(err, types.ErrEditConflict):
				app.editConflictResponse(w, r)
			case errors.Is(err, types.ErrDuplicateKey):
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
			case errors.Is(err, types.ErrRecordNotFound):
				app.notFoundResponse(w, r, id)
			default:
				app.serverErrorResponse(w, r, err)
		}
		return
	}
	slog.Debug("Retrieved version", "Version", version)

	pu := types.PartialUser{}

	err = app.readJSON(w, r, &pu)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	slog.Debug("Updating user with request payload", "id", id, "input", pu)

	slog.Debug("Validating partial user", "id", id, "partial_user", pu)
	v := validator.New()
	pu.Validate(v)
	if !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors())
		return
	}

	if pu.PlainPassword != nil {
		slog.Debug("Hashing password")
		hp, err := types.NewHashPW(*pu.PlainPassword)
		if err != nil {
			app.serverErrorResponse(w, r, err, "Couldn't update user")
			return
		}
		pu.HashedPassword = &hp
	}

	slog.Debug("Updating user in database")
	c, err := app.models.User.PartialUpdate(id, version, &pu)
	if err != nil {
		switch {
			case errors.Is(err, types.ErrEditConflict):
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
			case errors.Is(err, types.ErrRecordNotFound):
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
