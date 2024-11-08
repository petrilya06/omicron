package logger

import (
	"log/slog"
	"os"
)

func InitLogger() *slog.Logger {
	if os.Getenv("DEBUG") == "true" {
		logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.Level(slog.LevelDebug),
		}))
		return logger
	} else {
		f, err := os.OpenFile(os.Getenv("LOG_PATH"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			panic(err)
		}
		defer f.Close()

		logger := slog.New(slog.NewJSONHandler(f, &slog.HandlerOptions{
			Level: slog.Level(slog.LevelInfo),
		}))
		return logger
	}
}
