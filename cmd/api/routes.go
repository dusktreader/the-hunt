package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (app *application) routes() http.Handler {
	router := httprouter.New()

	router.NotFound = http.HandlerFunc(app.notFoundResponse)
	router.MethodNotAllowed = http.HandlerFunc(app.notAllowedResponse)

	app.logger.Debug("Adding health routes")
	router.HandlerFunc(http.MethodGet, "/v1/health", app.healthHandler)

	app.logger.Debug("Adding company routes")
	router.HandlerFunc(http.MethodPost, "/v1/companies", app.createCompanyHandler)
	router.HandlerFunc(http.MethodGet, "/v1/companies", app.showCompanyListHandler)
	router.HandlerFunc(http.MethodGet, "/v1/companies/:id", app.showCompanyHandler)

	return app.recoverPanic(router)
}
