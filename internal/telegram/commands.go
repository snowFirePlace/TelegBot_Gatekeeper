package telegram

import (
	"botTelegram/internal/config"
	"context"
	"fmt"
	"strings"
	"unicode/utf8"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	commandStart        = "start"
	commandRegistration = "reg"
	commandHelp         = "help"
	commandAddUser      = "add"
	commandShowUser     = "list"
	commandDeleteUser   = "del"
	commandAddAdmin     = "addadmin"
	commandShowAdmins   = "listadmin"
	commandDeleteAdmin  = "deladmin"
	commandVersion      = "ver"
)

func (b *Bot) Command(message *tgbotapi.Message) error {
	level := b.isAdministrator(message)
	msg := tgbotapi.NewMessage(message.Chat.ID, "")
	if level {
		switch message.Command() {
		case commandStart:
			msg.Text = msgWelcome + "\n" + msgHelpAdmin
		case commandHelp:
			msg.Text = msgHelpAdmin
		case commandAddUser:
			if err := b.addUser(message); err != nil {
				msg.Text = err.Error()
			} else {
				msg.Text = "Пользователь успешно добавлен"
			}
		case commandShowUser:
			userList, err := b.getUser(message)
			if err != nil {
				msg.Text = err.Error()
			} else {
				if utf8.RuneCountInString(userList) > 4096 {
					a := separateMessage(userList)
					for i, m := range a {
						msg.Text = m
						msg.ParseMode = "HTML"
						if i != len(a)-1 {

							b.bot.Send(msg)
						}
					}
				} else {
					msg.Text = userList
					msg.ParseMode = "HTML"
				}

			}
		case commandDeleteUser:
			if err := b.deleteUser(message); err != nil {
				msg.Text = err.Error()
			} else {
				msg.Text = "Пользователь удален"
			}
		case commandAddAdmin:
			if err := b.addAdmin(message); err != nil {
				msg.Text = err.Error()
			}
		case commandShowAdmins:
			adminList, err := b.showAdmins(message)
			if err != nil {
				msg.Text = err.Error()
			} else {
				msg.Text = adminList
			}
		case commandDeleteAdmin:
			if err := b.delAdmin(message); err != nil {
				msg.Text = err.Error()
			}
		case commandVersion:
			msg.Text = fmt.Sprintf("Версия Gatekeeper bot: %s", config.Version)
		default:
			msg.Text = "Неизвестная команда" + "\n" + msgHelpAdmin
		}
	} else {

		switch message.Command() {
		case commandStart:
			msg.Text = msgWelcome + "\n" + msgHelp
		case commandRegistration:

			if link, err := b.registration(message); err != nil {
				msg.Text = err.Error()
			} else {
				msg.Text = link
			}

		default:
			msg.Text = "Неизвестная команда" + "\n" + msgHelp
		}

	}
	b.bot.Send(msg)
	return nil
}

func (b *Bot) getUser(message *tgbotapi.Message) (string, error) {
	str, err := b.storage.ShowUsers(context.Background())
	if err != nil {
		return "", err
	}

	return str, nil
}
func (b *Bot) addUser(message *tgbotapi.Message) (err error) {
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

	if err = b.storage.AddUser(context.Background(), message.From.ID, fio, branch, unit, phone); err != nil {
		return err
	}
	return nil
}

func (b *Bot) deleteUser(message *tgbotapi.Message) error {
	q := message.CommandArguments()
	if len(strings.Replace(q, " ", "", -1)) == 0 {
		return fmt.Errorf("Не верный запрос.\n" + msgHelpDelUser)
	}
	fio := strings.TrimSpace(q)
	if idUser, err := b.storage.DelUser(context.Background(), message.From.ID, fio); err != nil {
		return err
	} else {
		if idUser == 0 {
			return fmt.Errorf("Пользователь %s удален", fio)
		}
		err := b.kickChatMember(idUser)
		if err != nil {
			return err
		}
	}
	return nil
}
func (b *Bot) getAdmins(message *tgbotapi.Message) (string, error) {
	str, err := b.storage.ShowAdmins(context.Background())
	if err != nil {
		return "", err
	}
	return str, nil
}
func (b *Bot) addAdmin(message *tgbotapi.Message) error {
	q := message.CommandArguments()
	if len(strings.Replace(q, " ", "", -1)) == 0 {
		return fmt.Errorf("Не верный запрос.\n" + msgHelpAddAdmin)
	}
	fio := strings.TrimSpace(q)
	if err := b.storage.AddAdmin(context.Background(), message.From.ID, fio); err != nil {
		return err
	}
	// TODO reflash admin list
	if err := b.storage.GetAdmins(context.Background()); err != nil {
		return err
	}
	return nil
}
func (b *Bot) showAdmins(message *tgbotapi.Message) (string, error) {
	str, err := b.storage.ShowAdmins(context.Background())
	if err != nil {
		return "", err
	}
	return str, nil
}
func (b *Bot) delAdmin(message *tgbotapi.Message) error {
	q := message.CommandArguments()
	if len(strings.Replace(q, " ", "", -1)) == 0 {
		return fmt.Errorf("Не верный запрос.\n" + msgHelpDelAdmin)
	}
	fio := strings.TrimSpace(q)
	if err := b.storage.DelAdmin(context.Background(), message.From.ID, fio); err != nil {
		return err
	}
	// TODO reflash admin list
	if err := b.storage.GetAdmins(context.Background()); err != nil {
		return err
	}
	return nil
}
func (b *Bot) registration(message *tgbotapi.Message) (link string, err error) {
	q := message.CommandArguments()
	if len(strings.Replace(q, " ", "", -1)) == 0 {
		return "", fmt.Errorf("Не верный запрос.\n" + msgHelp)
	}
	a := strings.Split(q, ",")
	if len(a) != 2 {
		err := fmt.Errorf("Не верный запрос.\n" + msgHelp)
		return "", err
	}
	fio := strings.TrimSpace(a[0])
	phone := strings.TrimSpace(a[1])
	username := message.From.FirstName + " " + message.From.LastName
	if err := b.storage.Registration(context.Background(), fio, phone, message.From.ID, username); err != nil {
		return "", err
	}

	link, err = b.getLink(message)
	if err != nil {
		return "", err
	}

	return
}

func separateMessage(message string) []string {
	if utf8.RuneCountInString(message) <= 4096 {
		return []string{message}
	}

	var separatedMessages []string
	for len(message) > 0 {
		if len(message) < 4096 {
			separatedMessages = append(separatedMessages, message)
			break
		}
		lastLineIndex := strings.LastIndex(message[:4095], "\n")
		separatedMessages = append(separatedMessages, message[:lastLineIndex+1])
		message = message[lastLineIndex+1:]
	}

	return separatedMessages
}
