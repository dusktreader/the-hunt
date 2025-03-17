package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/caarlos0/env/v11"
	"github.com/dusktreader/the-hunt/internal/data"
	"github.com/joho/godotenv"
)

const version = "0.1.0"

type config struct {
	port	int
	env		string
}

type application struct {
	config	data.Config
	logger	*slog.Logger
	models	data.Models
}

func main() {
	// If the .env file is not found, we don't care. It's optional.
	_ = godotenv.Load()
	var cfg data.Config
	err := env.Parse(&cfg)
	MaybeDie(err)

	logOpts := &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}
	if cfg.APIEnv == "development" {
		logOpts.Level = slog.LevelDebug
	}
	logger := slog.New(slog.NewTextHandler(os.Stdout, logOpts))

	dsn := buildDSN(cfg)
	logger.Info("Attempting to connect to the database", "dsn", dsn)
	db, err := openDB(dsn, cfg)
	MaybeDie(err)
	defer db.Close()
	logger.Info("Database connection pool established")

	app := &application{
		config: cfg,
		logger: logger,
		models: data.NewModels(db),
	}

	srv := &http.Server{
		Addr:			fmt.Sprintf(":%d", cfg.APIPort),
		Handler:		app.routes(),
		IdleTimeout:	time.Minute,
		ReadTimeout:	5 * time.Second,
		WriteTimeout:	10 * time.Second,
		ErrorLog:		slog.NewLogLogger(logger.Handler(), slog.LevelError),
	}

	logger.Info("Starting server", "Config", cfg)

	err = srv.ListenAndServe()
	MaybeDie(err)
}
