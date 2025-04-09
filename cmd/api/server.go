package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func (app *application) serve() error {
	srv := &http.Server{
		Addr:        fmt.Sprintf(":%d", app.config.APIPort),
		Handler:     app.routes(),
		IdleTimeout: 5 * time.Second, // These should probably be app settings
		ReadTimeout: 10 * time.Second,
		ErrorLog:    slog.NewLogLogger(slog.Default().Handler(), slog.LevelError),
	}

	shutdownError := make(chan error)

	go func() {
		slog.Debug("Configuring signal handler")
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		s := <-quit

		slog.Info("Received signal", "signal", s)

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		err := srv.Shutdown(ctx)
		if err != nil {
			shutdownError <- err
		}

		slog.Info("Completing background tasks")
		app.waiter.Wait()
		shutdownError <- nil
	}()

	slog.Info("Starting server", "Config", app.config)

	err := srv.ListenAndServe()

	if !errors.Is(err, http.ErrServerClosed) {
		slog.Error("Server closed with an unexpected error", "error", err)
		return err
	}

	err = <-shutdownError
	if err != nil {
		return err
	}

	slog.Info("Server shutdown successfully")

	return nil
}
