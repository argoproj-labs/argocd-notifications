package notifiers

import "gomodules.xyz/notify/smtp"

type Config struct {
	Email *smtp.Options `json:"email"`
}

type Notifier interface {
	Send(title string, body string, recipient string) error
}

func GetAll(config Config) map[string]Notifier {
	res := make(map[string]Notifier)
	if config.Email != nil {
		res["email"] = NewEmailNotifier(*config.Email)
	}
	return res
}
