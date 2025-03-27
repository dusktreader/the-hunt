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

	slog.Debug("Inserting new company into database")

	err = app.models.User.Insert(u)
	if err != nil {
		switch {
			case errors.Is(err, data.ErrDuplicateKey):
				// TODO: We probably don't want to use this to avoid user enumeration
				app.duplicateKeyResponse(w, r)
			default:
				app.serverErrorResponse(w, r, err, "Couldn't add user")
		}
		return
	}

	slog.Debug("Serializing response")

	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/v1/users/%d", u.ID))

	err = app.writeJSON(w, &data.JSONResponse{
		Envelope: 		data.Envelope{"user": u},
		StatusCode:		http.StatusCreated,
		Headers:		headers,
	})
	if err != nil {
		app.serverErrorResponse(w, r, err, "Failed to serialize user data")
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


	pw, err := data.NewPassword(input.Password)
	if err != nil {
		app.serverErrorResponse(w, r, err, "Failed to hash password")
		return
	}
	if pw.Matches(input.Password) {
		app.unauthorizedResponse(w, r)
		return
	}
	u.Password = *pw

	slog.Debug("Validating updated company", "id", id, "company", u)

	v := validator.New()
	u.Validate(v)
	if !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors())
		return
	}

	slog.Debug("Updating company in database")

	err = app.models.Company.Update(u)
	if err != nil {
		switch {
			case errors.Is(err, data.ErrEditConflict):
				app.editConflictResponse(w, r)
			case errors.Is(err, data.ErrDuplicateKey):
				app.duplicateKeyResponse(w, r)
			default:
				app.serverErrorResponse(w, r, err, "Couldn't update company")
		}
		return
	}

	slog.Debug("Serializing response")

	err = app.writeJSON(w, &data.JSONResponse{
		Envelope: 		data.Envelope{"company": u},
		StatusCode:		http.StatusOK,
	})
	if err != nil {
		app.serverErrorResponse(w, r, err, "Failed to serialize company data")
		return
	}
}

func (app *application) updatePartialCompanyHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.parseIdParam(r)
	if err != nil {
		app.badIdResponse(w, r, err)
		return
	}

	slog.Debug("Partially updating company", "id", id)

	version, err := app.models.Company.GetVersion(id)
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

	pc := data.PartialCompany{}

	err = app.readJSON(w, r, &pc)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	slog.Debug("Updating company with request payload", "id", id, "input", pc)

	slog.Debug("Validating partial company", "id", id, "partial_company", pc)
	v := validator.New()
	pc.Validate(v)
	if !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors())
		return
	}

	slog.Debug("Updating company in database")
	c, err := app.models.Company.PartialUpdate(id, version, &pc)
	if err != nil {
		switch {
			case errors.Is(err, data.ErrEditConflict):
				app.editConflictResponse(w, r)
			default:
				app.serverErrorResponse(w, r, err, "Couldn't update company")
		}
		return
	}

	slog.Debug("Serializing response")

	err = app.writeJSON(w, &data.JSONResponse{
		Envelope: 		data.Envelope{"company": c},
		StatusCode:		http.StatusOK,
	})
	if err != nil {
		app.serverErrorResponse(w, r, err, "Failed to serialize company data")
		return
	}
}

func (app *application) deleteCompanyHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.parseIdParam(r)
	if err != nil {
		app.badIdResponse(w, r, err)
		return
	}
	slog.Debug("Deleting company", "id", id)

	err = app.models.Company.Delete(id)
	if err != nil {
		switch {
			case errors.Is(err, data.ErrRecordNotFound):
				app.notFoundResponse(w, r, id)
			default:
				app.serverErrorResponse(w, r, err, "Couldn't delete company")
		}
		return
	}
	slog.Debug("Deleted company", "id", id)

	err = app.writeJSON(w, &data.JSONResponse{
		Envelope: 		data.Envelope{"message": "Company deleted successfully"},
		StatusCode:		http.StatusOK,
	})
	if err != nil {
		app.serverErrorResponse(w, r, err, "Failed to serialize response")
	}
}
