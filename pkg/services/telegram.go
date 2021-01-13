package services

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"

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
	_, err = bot.Send(tgbotapi.NewMessageToChannel("@"+dest.Recipient, notification.Message))
	return err
}
