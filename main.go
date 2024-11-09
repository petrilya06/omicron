package main

import (
	"log/slog"
	"omicron/bot"
	"omicron/db"
	"omicron/logger"

	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		panic("can`t load .env, " + err.Error())
	}

	slog.SetDefault(logger.InitLogger())

	db.MustInitDB()

	bot.RunBot()
}
