package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/dusktreader/the-hunt/internal/data"
	"github.com/dusktreader/the-hunt/internal/validator"
	"github.com/julienschmidt/httprouter"
)

func (app *application) showCompanyListHandler(w http.ResponseWriter, _ *http.Request) {
	fmt.Fprintln(w, "get all companies")
}

func (app *application) createCompanyHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Name		string		`json:"name"`
		URL			string		`json:"url"`
		TechStack	[]string	`json:"tech_stack"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	v := validator.New()

	c := &data.Company{
		ID:			13,
		CreatedAt:	time.Now(),
		Name:		input.Name,
		URL:		input.URL,
		TechStack:	input.TechStack,
	}

	c.Validate(v)
	if !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors())
		return
	}

	fmt.Fprintf(w, "%+v\n", input)
}

func (app *application) showCompanyHandler(w http.ResponseWriter, r *http.Request) {
	_ = httprouter.ParamsFromContext(r.Context())
	id, err := app.parseIdParam(r)
	if err != nil {
		app.errorResponse(w, r, &data.ErrorPackage{
			StatusCode:	http.StatusBadRequest,
			Message:	"Invalid ID format",
			Error:		err,
		})
		return
	}
	app.logger.Debug("Showing the company details", "id", id)

	company := &data.Company{
		ID:			id,
		CreatedAt:	time.Now(),
		Name:		"Test Company",
		URL:		"https://example.com",
		TechStack:  []string{"Go", "PostgreSQL", "Docker"},
	}

	err = app.writeJSON(w, &data.JSONResponse{
		Data:			company,
		StatusCode:		http.StatusBadRequest,
		EnvelopeKey:	"company",
	})
	if err != nil {
		app.errorResponse(w, r, &data.ErrorPackage{
			StatusCode:	http.StatusInternalServerError,
			Message:	"Failed to serialize company data",
			Error:		err,
		})
	}
}
