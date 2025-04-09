package main

import (
	"expvar"
	"log/slog"
	"net/http"

	"github.com/dusktreader/the-hunt/internal/types"
	"github.com/julienschmidt/httprouter"
)

type Route struct {
	method	string
	path	string
	handler	http.HandlerFunc
}

type RouteList []Route

func (app *application) routes() http.Handler {
	router := httprouter.New()

	router.NotFound = http.HandlerFunc(app.routeNotFoundResponse)
	router.MethodNotAllowed = http.HandlerFunc(app.notAllowedResponse)

	// auth := app.requireAuthorization
	perms := app.requirePermissions

	routes := RouteList{
		{http.MethodGet,		"/health",				app.healthHandler},

		{http.MethodPost,		"/v1/companies",		perms(app.createCompanyHandler, types.All, types.CompanyWrite)},
		{http.MethodGet,		"/v1/companies",		perms(app.readManyCompaniesHandler, types.All, types.CompanyRead)},
		{http.MethodGet,		"/v1/companies/:id",	perms(app.readCompanyHandler, types.All, types.CompanyRead)},
		{http.MethodPut,		"/v1/companies/:id",	perms(app.updateCompanyHandler, types.All, types.CompanyWrite)},
		{http.MethodPatch,		"/v1/companies/:id",	perms(app.updatePartialCompanyHandler, types.All, types.CompanyWrite)},
		{http.MethodDelete,		"/v1/companies/:id",	perms(app.deleteCompanyHandler, types.All, types.CompanyWrite)},

		{http.MethodPost,		"/v1/users",			perms(app.createUserHandler, types.All, types.UserWrite)},
		{http.MethodGet,		"/v1/users",			perms(app.readManyUsersHandler, types.All, types.UserRead)},
		{http.MethodGet,		"/v1/users/:id",		perms(app.readUserHandler, types.All, types.UserRead)},
		{http.MethodPut,		"/v1/users/:id",		perms(app.updateUserHandler, types.All, types.UserWrite)},
		{http.MethodPatch,		"/v1/users/:id",		perms(app.updatePartialUserHandler, types.All, types.UserWrite)},
		{http.MethodDelete,		"/v1/users/:id",		perms(app.deleteUserHandler, types.All, types.UserWrite)},
		{http.MethodPost,		"/v1/users/activate",	app.activateUserHandler},

		{http.MethodPost,		"/v1/login",			app.loginHandler},

	}

	slog.Debug("Adding routes")
	for _, r := range routes {
		router.HandlerFunc(r.method, r.path, r.handler)
	}

	if app.config.APIEnv.IsDev() {
		router.Handler(http.MethodGet, "/metrics", expvar.Handler())
	}

	return chainMiddleware(
		router,
		app.recoverPanic,
		app.metrics,
		app.enableCORS,
		app.rateLimit,
		app.authenticate,
	)
}
