package main

import (
	"fmt"
	"net/http"

	"github.com/tomasen/realip"

	"github.com/dusktreader/the-hunt/internal/data"
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

func (app * application) rateLimit(next http.Handler) http.Handler {
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
