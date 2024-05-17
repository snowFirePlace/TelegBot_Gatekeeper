package telegram

import (
	"context"
	"fmt"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	commandStart       = "start"
	commandHelp        = "help"
	commandAddUser     = "add"
	commandUsers       = "list"
	commandDeleteUser  = "del"
	commandAddAdmin    = "addadmin"
	commandShowAdmins  = "listadmin"
	commandDeleteAdmin = "deladmin"
	commandVersion     = "ver"
)

func (b *Bot) Command(message *tgbotapi.Message) error {
	level := false
	for _, admin := range b.storage.Admins {
		if message.From.ID == admin.ID {
			level = true
		}
	}
	msg := tgbotapi.NewMessage(message.Chat.ID, "")
	if level {
		switch message.Command() {
		case commandStart:
			msg.Text = msgHelpAddUser
		case commandHelp:
			msg.Text = msgHelpAddUser
		case commandAddUser:
			if err := b.commandAddUser(message); err != nil {
				msg.Text = err.Error()
			}
		case commandUsers:
			s, err := b.commandGetUser(message)
			if err != nil {
				msg.Text = err.Error()
			} else {
				msg.Text = s
			}
		case commandDeleteUser:
			if err := b.commandDeleteUser(message); err != nil {
				msg.Text = err.Error()
			}
		case commandAddAdmin:
			if err := b.commandAddAdmin(message); err != nil {
				msg.Text = err.Error()
			}
		case commandShowAdmins:
			s, err := b.commandShowAdmins(message)
			if err != nil {
				msg.Text = err.Error()
			} else {
				msg.Text = s
			}
		case commandVersion:

		default:
			msg.Text = "Неизвестная команда" + msgHelp

		}

	} else {
		switch message.Command() {
		case commandStart:

		default:

		}

	}
	b.bot.Send(msg)

	return nil

}
func (b *Bot) commandGetUser(message *tgbotapi.Message) (string, error) {
	str, err := b.storage.Show(context.Background())
	if err != nil {
		return "", err
	}
	return str, nil
}
func (b *Bot) commandAddUser(message *tgbotapi.Message) (err error) {
	q := message.CommandArguments()
	if len(strings.Replace(q, " ", "", -1)) == 0 {
		return fmt.Errorf("Не верный запрос.\n" + msgHelpAddUser)
	}
	a := strings.Split(q, ",")
	if len(a) != 4 {
		err = fmt.Errorf("Не верный запрос.\n" + msgHelpAddUser)
		return err
	}
	fio := strings.TrimSpace(a[0])
	branch := strings.TrimSpace(a[1])
	unit := strings.TrimSpace(a[2])
	phone := strings.TrimSpace(a[3])

	if err = b.storage.Add(context.Background(), fio, branch, unit, phone); err != nil {
		return err
	}
	return nil
}

func (b *Bot) commandDeleteUser(message *tgbotapi.Message) error {
	q := message.CommandArguments()
	if len(strings.Replace(q, " ", "", -1)) == 0 {
		return fmt.Errorf("Не верный запрос.\n" + msgHelpDelUser)
	}
	fio := strings.TrimSpace(q)
	if err := b.storage.Delete(context.Background(), fio); err != nil {
		return err
	}
	return nil
}
func (b *Bot) commandGetAdmins(message *tgbotapi.Message) (string, error) {
	str, err := b.storage.ShowAdmins(context.Background())
	if err != nil {
		return "", err
	}
	return str, nil
}
func (b *Bot) commandAddAdmin(message *tgbotapi.Message) error {
	q := message.CommandArguments()
	if len(strings.Replace(q, " ", "", -1)) == 0 {
		return fmt.Errorf("Не верный запрос.\n" + msgHelpAddAdmin)
	}
	fio := strings.TrimSpace(q)
	if err := b.storage.AddAdmin(context.Background(), fio, message.From.ID); err != nil {
		return err
	}
	// TODO reflash admin list
	if err := b.storage.GetAdmins(context.Background()); err != nil {
		return err
	}
	return nil
}
func (b *Bot) commandShowAdmins(message *tgbotapi.Message) (string, error) {
	str, err := b.storage.ShowAdmins(context.Background())
	if err != nil {
		return "", err
	}
	return str, nil
}
func (b *Bot) commandDelAdmin(message *tgbotapi.Message) error {
	q := message.CommandArguments()
	if len(strings.Replace(q, " ", "", -1)) == 0 {
		return fmt.Errorf("Не верный запрос.\n" + msgHelpDelAdmin)
	}
	fio := strings.TrimSpace(q)
	if err := b.storage.DelAdmin(context.Background(), fio); err != nil {
		return err
	}
	// TODO reflash admin list
	if err := b.storage.GetAdmins(context.Background()); err != nil {
		return err
	}
	return nil
}
