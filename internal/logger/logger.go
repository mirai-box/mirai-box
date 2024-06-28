package logger

import (
	"log/slog"
	"os"

	"github.com/mirai-box/mirai-box/internal/config"
)

func Setup(conf *config.Config) {
	var logHandler slog.Handler

	if conf.IsLocal() {
		logHandler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: conf.LogLevel,
		})
	} else {
		logHandler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: conf.LogLevel,
		})
	}

	logger := slog.New(logHandler)
	slog.SetDefault(logger)
}
