package main

import (
	"expvar"
	"flag"
	"fmt"
	"log/slog"
	"runtime"
	"sync"
	"time"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"

	"github.com/dusktreader/the-hunt/internal/data"
	"github.com/dusktreader/the-hunt/internal/logs"
	"github.com/dusktreader/the-hunt/internal/mailer"
)

func main() {
	showVer := flag.Bool("version", false, "Display version and exit")
	flag.Parse()
	if *showVer {
		fmt.Println(Version())
		Close()
	}

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

	if cfg.APIEnv.IsDev() {
		expvar.NewString("version").Set(Version())
		expvar.Publish("goroutines", expvar.Func(func() any {
			return runtime.NumGoroutine()
		}))
		expvar.Publish("database", expvar.Func(func() any {
			return db.Stats()
		}))
		expvar.Publish("timestamp", expvar.Func(func() any {
			return time.Now().Format(time.RFC3339)
		}))
		expvar.Publish("config", expvar.Func(func() any {
			return cfg
		}))
	}

	app := &application{
		config: cfg,
		models: data.NewModels(db, data.NewModelConfig(cfg)),
		mailer: mailer,
		waiter: new(sync.WaitGroup),
	}

	MaybeDie(app.serve())
	Close("App finished")
}
