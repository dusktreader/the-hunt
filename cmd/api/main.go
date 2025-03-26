package main

import (
	"log/slog"
	"sync"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"

	"github.com/dusktreader/the-hunt/internal/data"
	"github.com/dusktreader/the-hunt/internal/logs"
	"github.com/dusktreader/the-hunt/internal/mailer"
)

const version = "0.1.0"

type config struct {
	port	int
	env		string
}

type application struct {
	config	data.Config
	models	data.Models
	mailer	*mailer.Mailer
	waiter	*sync.WaitGroup
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

	mailer, err := mailer.New(cfg)
	MaybeDie(err)

	app := &application{
		config: cfg,
		models: data.NewModels(db, data.NewModelConfig(cfg)),
		mailer: mailer,
		waiter: new(sync.WaitGroup),
	}

	MaybeDie(app.serve())
	Close("App finished")
}
