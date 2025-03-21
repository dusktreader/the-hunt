package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/caarlos0/env/v11"
	"github.com/dusktreader/the-hunt/internal/data"
	"github.com/dusktreader/the-hunt/internal/logs"
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

	logs.InitLogger(cfg)

	dsn := buildDSN(cfg)
	slog.Info("Attempting to connect to the database", "dsn", dsn)
	db, err := openDB(dsn, cfg)
	MaybeDie(err)
	defer db.Close()
	slog.Info("Database connection pool established")

	app := &application{
		config: cfg,
		models: data.NewModels(db, data.NewModelConfig(cfg)),
	}

	srv := &http.Server{
		Addr:			fmt.Sprintf(":%d", cfg.APIPort),
		Handler:		app.routes(),
		IdleTimeout:	time.Minute,
		ReadTimeout:	5 * time.Second,
		WriteTimeout:	10 * time.Second,
		ErrorLog:		slog.NewLogLogger(slog.Default().Handler(), slog.LevelError),
	}

	slog.Info("Starting server", "Config", cfg)

	err = srv.ListenAndServe()
	MaybeDie(err)
}
