package logs

import (
	"log/slog"
	"os"

	"github.com/dusktreader/the-hunt/internal/data"
	"github.com/dusktreader/the-hunt/internal/types"
)

func InitLogger(cfg data.Config) {
	logOpts := &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}
	if cfg.APIEnv == types.EnvDev {
		logOpts.Level = slog.LevelDebug
	}
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, logOpts)))
}
