package main

import (
	"botTelegram/internal/config"
	"botTelegram/internal/sqlite"
	"botTelegram/internal/telegram"
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	sqliteStoragePath = "data" + string(os.PathSeparator) + "sqlite.db"
)

var (
	cfg = config.Get()
)

var version string

func init() {
	cmd := exec.Command("git", "describe", "--tags")
	output, err := cmd.Output()
	if err != nil {
		fmt.Println("Error getting version:", err)
		os.Exit(1)
	}
	version = strings.TrimSpace(string(output))
}
func main() {
	fmt.Printf("Version: %s\n", version)
	os.Exit(1)
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

	botApi.Debug = true

	bot := telegram.NewBot(botApi, cfg.TelegramChannel, s)
	// go func() {
	// 	for {
	// 		time.Sleep(24 * time.Hour)

	// 	}
	// }()
	bot.Start()

}
