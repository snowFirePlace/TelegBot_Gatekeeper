package telegram

import (
	"fmt"
	"log"
	"net/http"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"golang.org/x/net/proxy"
	"snowfireplace.com/config"
	"snowfireplace.com/data"
	p "snowfireplace.com/proxy"
)

func Run() {
	var bot *tgbotapi.BotAPI
	var err error
	proxyList := p.GetList()
	if config.Proxy {
		var dialSocksProxy proxy.Dialer
		for _, prox := range proxyList {
			dialSocksProxy, err = proxy.SOCKS5("tcp", prox, nil, proxy.Direct)
			if err != nil {
				log.Printf(fmt.Sprintf("|ERROR| Connecting to proxy: %s\r\n", err))
				continue
			}
			tr := &http.Transport{Dial: dialSocksProxy.Dial}
			myClient := &http.Client{
				Transport: tr,
			}
			bot, err = tgbotapi.NewBotAPIWithClient("", myClient)

			if err != nil {
				log.Printf(fmt.Sprintf("|ERROR| %s\r\n", err))
				continue
			}
			log.Printf("Authorized on account %s with proxy on %s", bot.Self.UserName, prox)
			break
		}
	} else {
		bot, err = tgbotapi.NewBotAPI("")
		if err != nil {
			log.Panic(fmt.Sprintf("|ERROR| %s\r\n", err))
		}
		log.Printf("Authorized on account %s", bot.Self.UserName)
	}

	bot.Debug = false

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)
	go func() {
		for {
			select {
			case m := <-data.MessageChan:
				bot.Send(tgbotapi.NewMessage(-(config.IDGroup), m))
			}
		}
	}()
	for update := range updates {
		var reply string
		if update.Message == nil { // ignore any non-Message Updates
			continue
		}
		if update.Message.Chat.IsGroup() {
			continue
		}
		switch update.Message.Command() {
		case "version":
			reply = fmt.Sprintf("Версия: %s", data.Version)
		case "about":
			reply = fmt.Sprintf("Бот")
		case "state":
			reply = fmt.Sprintf("Бот работает с %s", data.StartTime.Format("15:04 02.01.2006"))
		}
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, reply)
		bot.Send(msg)
	}
}
