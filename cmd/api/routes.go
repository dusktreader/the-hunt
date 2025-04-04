package main

import (
	"log/slog"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (app *application) routes() http.Handler {
	router := httprouter.New()

	router.NotFound = http.HandlerFunc(app.routeNotFoundResponse)
	router.MethodNotAllowed = http.HandlerFunc(app.notAllowedResponse)

	slog.Debug("Adding health routes")
	router.HandlerFunc(http.MethodGet, "/health", app.healthHandler)

	slog.Debug("Adding company routes")
	router.HandlerFunc(http.MethodPost,		"/v1/companies",		app.createCompanyHandler)
	router.HandlerFunc(http.MethodGet,		"/v1/companies",		app.readManyCompaniesHandler)
	router.HandlerFunc(http.MethodGet,		"/v1/companies/:id",	app.readCompanyHandler)
	router.HandlerFunc(http.MethodPut,		"/v1/companies/:id",	app.updateCompanyHandler)
	router.HandlerFunc(http.MethodPatch,	"/v1/companies/:id",	app.updatePartialCompanyHandler)
	router.HandlerFunc(http.MethodDelete,	"/v1/companies/:id",	app.deleteCompanyHandler)

	slog.Debug("Adding user routes")
	router.HandlerFunc(http.MethodPost,		"/v1/users",			app.createUserHandler)
	router.HandlerFunc(http.MethodGet,		"/v1/users",			app.readManyUsersHandler)
	router.HandlerFunc(http.MethodGet,		"/v1/users/:id",		app.readUserHandler)
	router.HandlerFunc(http.MethodPut,		"/v1/users/:id",		app.updateUserHandler)
	router.HandlerFunc(http.MethodPatch,	"/v1/users/:id",		app.updatePartialUserHandler)
	router.HandlerFunc(http.MethodDelete,	"/v1/users/:id",		app.deleteUserHandler)
	router.HandlerFunc(http.MethodPost,		"/v1/users/activate",	app.activateUserHandler)

	return app.recoverPanic(app.rateLimit(router))
}
