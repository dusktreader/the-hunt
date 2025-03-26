package main

import (
	"context"
	"net/http"
	"log/slog"

	"github.com/dusktreader/the-hunt/internal/types"
)

type contextKey string

const userContextKey = contextKey("user")
const permsContextKey = contextKey("perms")
const adminContextKey = contextKey("admin")

func (app *application) contextSetUser(r *http.Request, user *types.User) *http.Request {
	ctx := context.WithValue(r.Context(), userContextKey, user)
	return r.WithContext(ctx)
}

func (app *application) contextGetUser(r *http.Request, dontPanic ...bool) *types.User {
	user, ok := r.Context().Value(userContextKey).(*types.User)
	if !ok {
		if len(dontPanic) > 0 && dontPanic[0] {
			slog.Debug("User not found in context, but not panicking")
			return nil
		} else {
			panic("could not find user in request context")
		}
	}
	return user
}

func (app *application) contextSetPerms(r *http.Request, perms *types.PermissionSet) *http.Request {
	ctx := context.WithValue(r.Context(), permsContextKey, perms)
	return r.WithContext(ctx)
}

func (app *application) contextGetPerms(r *http.Request, dontPanic ...bool) *types.PermissionSet {
	perms, ok := r.Context().Value(permsContextKey).(*types.PermissionSet)
	if !ok {
		if len(dontPanic) > 0 && dontPanic[0] {
			slog.Debug("Permissions not found in context, but not panicking")
			return nil
		} else {
			panic("could not find permissions in request context")
		}
	}
	return perms
}

func (app *application) contextSetAdmin(r *http.Request) *http.Request {
	ctx := context.WithValue(r.Context(), adminContextKey, true)
	return r.WithContext(ctx)
}

func (app *application) contextGetAdmin(r *http.Request, dontPanic ...bool) bool {
	isAdmin, ok := r.Context().Value(adminContextKey).(bool)
	if !ok {
		if len(dontPanic) > 0 && dontPanic[0] {
			slog.Debug("Admin not found in context, but not panicking")
			return false
		} else {
			panic("could not find admin in request context")
		}
	}
	return isAdmin
}
