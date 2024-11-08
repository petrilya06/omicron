package main

import (
	"log/slog"
	"omicron/logger"

	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		panic("can`t load .env, " + err.Error())
	}

	slog.SetDefault(logger.InitLogger())
	slog.Info("starting the application")
}
