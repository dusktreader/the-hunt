package main

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/tomasen/realip"

	"github.com/dusktreader/the-hunt/internal/data"
	"github.com/dusktreader/the-hunt/internal/types"
	"github.com/dusktreader/the-hunt/internal/validator"
)

func (app *application) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				w.Header().Set("Connection", "close")
				app.errorResponse(w, r, &data.ErrorPackage{
					Error:		fmt.Errorf("%+v", err),
					Message:	"An unexpected error occurred",
					LogMessage:	"PANIC RECOVER",
					StatusCode:	http.StatusInternalServerError,
				})
			}
		}()

		next.ServeHTTP(w, r)
	})
}

func (app *application) rateLimit(next http.Handler) http.Handler {
	if app.config.APIEnv.IsDev() {
		return next
	}

	clients	:= data.NewClientMap(app.config)
	go clients.CleanCycle()

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if app.config.LimitEnabled {
			if !clients.IsIpAllowed(realip.FromRequest(r)) {
				app.rateLimitExceededResponse(w, r)
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}

func (app *application) authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		slog.Debug("Authenticating request")

		w.Header().Add("Vary", "Authorization")

		slog.Debug("Checking for authorization header")
		authorizationHeader := r.Header.Get("Authorization")
		if authorizationHeader == "" {
			slog.Debug("No authorization header provided. Binding AnonymousUser to the request")
			r = app.contextSetUser(r, types.AnonymousUser)
			next.ServeHTTP(w, r)
			return
		}

		slog.Debug("Parsing token from auth header")
		headerParts := strings.Split(authorizationHeader, " ")
		if len(headerParts) != 2 || headerParts[0] != "Bearer" {
			slog.Debug("Bad header", "parts", headerParts)
			app.invalidTokenResponse(w, r, types.ScopeAuthentication)
			return
		}

		plainToken := types.PlainToken(headerParts[1])
		slog.Debug("Parsed token", "token", plainToken)

		slog.Debug("Validating token")
		v := validator.New()
		plainToken.Validate(v)
		if !v.Valid() {
			app.invalidTokenResponse(w, r, types.ScopeAuthentication)
			return
		}

		slog.Debug("Looking up token in database")
		t, err := app.models.Token.GetOne(plainToken, types.ScopeAuthentication)
		if err != nil {
			switch {
				case errors.Is(err, types.ErrRecordNotFound):
					app.invalidTokenResponse(w, r, types.ScopeAuthentication)
				default:
					app.serverErrorResponse(w, r, err, "Couldn't parse authentication token")
			}
			return
		}

		if t.IsAdmin {
			slog.Debug("Token is admin token. Binding admin to the request")
			r = app.contextSetAdmin(r)
		} else {
			slog.Debug("Token is not admin token. Fetching and binding user", "id", t.UserID)
			u, err := app.models.User.GetOne(t.UserID)
			if err != nil {
				switch {
					case errors.Is(err, types.ErrRecordNotFound):
						app.notFoundResponse(w, r, t.UserID)
					default:
						app.serverErrorResponse(w, r, err, "Couldn't find user for token")
				}
				return
			}

			slog.Debug("Authenticated user from auth token. Binding user to the request", "user", u)
			r = app.contextSetUser(r, u)

			slog.Debug("Binding user permissions")
			perms, err := app.models.Permission.GetForUser(t.UserID)
			if err != nil {
				app.serverErrorResponse(w, r, err, "Couldn't retrieve user permissions")
				return
			}
			r = app.contextSetPerms(r, perms)
		}
		next.ServeHTTP(w, r)
	})
}

func (app *application) requireAuthorization(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		slog.Debug("Requiring authorization for request")

		isAdmin := app.contextGetAdmin(r, true)

		if !isAdmin {
			slog.Debug("Request was not made with admin token. Checking for user token")
			user := app.contextGetUser(r, true)

			if user.IsAnonymous() {
				app.unauthorizedResponse(w, r)
				return
			}
			slog.Debug("User is authorized!")
		}

		next.ServeHTTP(w, r)
	})
}

func (app *application) requirePermissions(
	next http.HandlerFunc,
	strategy types.PermissionStrategy,
	perms ...types.PermCode,
) http.HandlerFunc {
	fn := func(w http.ResponseWriter, r *http.Request) {
		slog.Debug("Checking permissions for request", "perms", perms)
		if len(perms) > 0 {

			isAdmin := app.contextGetAdmin(r, true)

			if !isAdmin {
				slog.Debug(
					"Request was not made with admin token. Checking for user permissions",
					"strategy", strategy,
					"perms", perms,
				)
				userPerms := app.contextGetPerms(r, true)
				slog.Debug("User perms", "perms", userPerms)
				if !types.HasPerms(userPerms, strategy, perms...) {
					app.forbiddenResponse(w, r)
					return
				}
				slog.Debug("User has required permissions", "perms", userPerms)
			}
		}

		next.ServeHTTP(w, r)
	}
	return app.requireAuthorization(fn)
}
