package telegram

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (b *Bot) isAdministrator(message *tgbotapi.Message) bool {
	for _, admin := range b.storage.Admins {
		if message.From.ID == admin.ID {
			return true
		}
	}
	return false
}
func (b *Bot) getLink(message *tgbotapi.Message) (link string, err error) {
	chat := tgbotapi.ChatInviteLinkConfig{tgbotapi.ChatConfig{ChatID: b.chatID}}
	if link, err = b.bot.GetInviteLink(chat); err != nil {
		return "", err
	}
	return link, nil
}

func (b *Bot) kickChatMember(idUser int64) error {
	config := tgbotapi.KickChatMemberConfig{
		ChatMemberConfig: tgbotapi.ChatMemberConfig{
			ChatID: b.chatID,
			UserID: idUser,
		},
	}
	_, err := b.bot.Request(config)
	if err != nil {
		return err

	}
	return nil
}
