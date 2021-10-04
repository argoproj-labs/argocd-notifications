package services

import (
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

type TelegramOptions struct {
	Token string `json:"token"`
}

func NewTelegramService(opts TelegramOptions) NotificationService {
	return &telegramService{opts: opts}
}

type telegramService struct {
	opts TelegramOptions
}

func (s telegramService) Send(notification Notification, dest Destination) error {
	bot, err := tgbotapi.NewBotAPI(s.opts.Token)
	if err != nil {
		return err
	}

	if strings.HasPrefix(dest.Recipient, "-") {
		chatID, err := strconv.ParseInt(dest.Recipient, 10, 64)
		if err != nil {
			return err
		}

		_, err = bot.Send(tgbotapi.NewMessage(chatID, notification.Message))
		if err != nil {
			return err
		}
	} else {
		_, err := bot.Send(tgbotapi.NewMessageToChannel("@"+dest.Recipient, notification.Message))
		if err != nil {
			return err
		}
	}

	return nil
}
