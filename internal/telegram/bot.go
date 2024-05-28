package telegram

import (
	"botTelegram/internal/sqlite"
	"context"
	"fmt"
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Bot struct {
	bot     *tgbotapi.BotAPI
	chatID  int64
	storage sqlite.Storage
}

func NewBot(bot *tgbotapi.BotAPI, chatID int64, storage *sqlite.Storage) *Bot {
	return &Bot{
		bot:     bot,
		chatID:  chatID,
		storage: *storage,
	}
}
func (b *Bot) Start() error {
	log.Printf("Authorized on account %s", b.bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := b.initUpdatesCannel()
	if err != nil {
		return err
	}
	b.storage.GetAdmins(context.Background())
	b.handleUpdates(updates)
	return nil
}
func (b *Bot) handleUpdates(updates tgbotapi.UpdatesChannel) {
	for update := range updates {
		fmt.Println(update)
		if update.Message == nil { // If we got a message

		}

		if update.EditedMessage != nil {
			continue
		}
		if update.ChannelPost != nil {
			continue
		}
		if update.Message.IsCommand() { // If we got a command
			b.Command(update.Message)
			continue
		}
	}
}
func (b *Bot) initUpdatesCannel() (tgbotapi.UpdatesChannel, error) {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	return b.bot.GetUpdatesChan(u), nil
}
