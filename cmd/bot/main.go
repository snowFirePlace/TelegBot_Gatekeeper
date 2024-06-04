package main

import (
	"botTelegram/internal/config"
	"botTelegram/internal/sqlite"
	"botTelegram/internal/telegram"
	"context"
	"log"
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	sqliteStoragePath = "data" + string(os.PathSeparator) + "sqlite.db"
)

var (
	version string
	cfg     = config.Get()
)

func main() {
	config.Version = version

	s, err := sqlite.New(sqliteStoragePath)
	if err != nil {
		log.Fatalf("can't connect to storage: %s", err)
	}
	if err := s.Init(context.Background(), cfg.Admin); err != nil {
		log.Fatal("can't init storage: ", err)
	}

	botApi, err := tgbotapi.NewBotAPI(cfg.TelegramBotToken)
	if err != nil {
		log.Panic(err)
	}

	bot := telegram.NewBot(botApi, cfg.TelegramChannel, s)
	// go func() {
	// 	for {
	// 		time.Sleep(24 * time.Hour)

	bot.Start()
}
