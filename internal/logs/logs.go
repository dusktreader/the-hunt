package logs

import (
	"log/slog"
	"os"

	"github.com/dusktreader/the-hunt/internal/data"
)

func InitLogger(cfg data.Config) {
	logOpts := &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}
	if cfg.APIEnv == "development" {
		logOpts.Level = slog.LevelDebug
	}
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, logOpts)))
}
