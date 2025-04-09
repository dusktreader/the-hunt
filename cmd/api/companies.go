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

func (app *application) createCompanyHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Name      string   `json:"name"`
		URL       string   `json:"url"`
		TechStack []string `json:"tech_stack"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	slog.Debug("Creating a new company", "input", input)

	v := validator.New()

	c := &types.Company{
		Name:      input.Name,
		URL:       input.URL,
		TechStack: input.TechStack,
	}

	slog.Debug("Validating new company")

	c.Validate(v)
	if !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors())
		return
	}

	slog.Debug("Inserting new company into database")

	err = app.models.Company.Insert(c)
	if err != nil {
		switch {
		case errors.Is(err, types.ErrDuplicateKey):
			app.duplicateKeyResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err, "Couldn't add company")
		}
		return
	}

	slog.Debug("Serializing response")

	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/v1/companies/%d", c.ID))

	err = app.writeJSON(w, &data.JSONResponse{
		Envelope:   data.Envelope{"company": c},
		StatusCode: http.StatusCreated,
		Headers:    headers,
	})
	if err != nil {
		app.serverErrorResponse(w, r, err, "Failed to serialize company data")
		return
	}
}

func (app *application) readCompanyHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.parseIdParam(r)
	if err != nil {
		app.badIdResponse(w, r, err)
		return
	}
	slog.Debug("Fetching company details", "id", id)

	c, err := app.models.Company.GetOne(id)
	if err != nil {
		switch {
		case errors.Is(err, types.ErrRecordNotFound):
			app.notFoundResponse(w, r, id)
		default:
			app.serverErrorResponse(w, r, err, "Couldn't retrieve company")
		}
		return
	}
	slog.Debug("Retrieved company", "Company", *c)

	err = app.writeJSON(w, &data.JSONResponse{
		Envelope:   data.Envelope{"company": c},
		StatusCode: http.StatusOK,
	})
	if err != nil {
		app.serverErrorResponse(w, r, err, "Failed to serialize company data")
	}
}

func (app *application) readManyCompaniesHandler(w http.ResponseWriter, r *http.Request) {
	slog.Debug("Fetching company list")

	qs := r.URL.Query()
	v := validator.New()
	filters := data.ParseFilters(
		qs,
		v,
		data.FilterConstraints{
			Search: data.CompanySearchFields.Check,
			Sort:   data.CompanySortFields.Check,
			In:     data.CompanyInFields.Check,
		},
	)
	if !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors())
		return
	}

	slog.Debug("Retrieved filters", "filters", filters)

	companies, metadata, err := app.models.Company.GetMany(filters)
	if err != nil {
		app.serverErrorResponse(w, r, err, "Couldn't retrieve companies")
	}
	slog.Debug("Fetched companies", "metadata", metadata)

	err = app.writeJSON(w, &data.JSONResponse{
		StatusCode: http.StatusOK,
		Envelope: data.Envelope{
			"companies": companies,
			"metadata":  metadata,
		},
	})
	if err != nil {
		app.serverErrorResponse(w, r, err, "Failed to serialize company data")
	}
}

func (app *application) updateCompanyHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.parseIdParam(r)
	if err != nil {
		app.badIdResponse(w, r, err)
		return
	}

	slog.Debug("Updating company", "id", id)

	c, err := app.models.Company.GetOne(id)
	if err != nil {
		switch {
		case errors.Is(err, types.ErrRecordNotFound):
			app.notFoundResponse(w, r, id)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}
	slog.Debug("Retrieved company", "Company", *c)

	var input struct {
		Name      string   `json:"name"`
		URL       string   `json:"url"`
		TechStack []string `json:"tech_stack"`
	}

	err = app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	slog.Debug("Updating company with request payload", "id", id, "input", input)

	c.Name = input.Name
	c.URL = input.URL
	c.TechStack = input.TechStack

	slog.Debug("Validating updated company", "id", id, "company", c)

	v := validator.New()
	c.Validate(v)
	if !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors())
		return
	}

	slog.Debug("Updating company in database")

	err = app.models.Company.Update(c)
	if err != nil {
		switch {
		case errors.Is(err, types.ErrEditConflict):
			app.editConflictResponse(w, r)
		case errors.Is(err, types.ErrDuplicateKey):
			app.duplicateKeyResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err, "Couldn't update company")
		}
		return
	}

	slog.Debug("Serializing response")

	err = app.writeJSON(w, &data.JSONResponse{
		Envelope:   data.Envelope{"company": c},
		StatusCode: http.StatusOK,
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
		case errors.Is(err, types.ErrRecordNotFound):
			app.notFoundResponse(w, r, id)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}
	slog.Debug("Retrieved version", "Version", version)

	pc := types.PartialCompany{}

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
		case errors.Is(err, types.ErrEditConflict):
			app.editConflictResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err, "Couldn't update company")
		}
		return
	}

	slog.Debug("Serializing response")

	err = app.writeJSON(w, &data.JSONResponse{
		Envelope:   data.Envelope{"company": c},
		StatusCode: http.StatusOK,
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
		case errors.Is(err, types.ErrRecordNotFound):
			app.notFoundResponse(w, r, id)
		default:
			app.serverErrorResponse(w, r, err, "Couldn't delete company")
		}
		return
	}
	slog.Debug("Deleted company", "id", id)

	err = app.writeJSON(w, &data.JSONResponse{
		Envelope:   data.Envelope{"message": "Company deleted successfully"},
		StatusCode: http.StatusOK,
	})
	if err != nil {
		app.serverErrorResponse(w, r, err, "Failed to serialize response")
	}
}
