package notifiers

type Config struct {
	Email *EmailOptions `json:"email"`
	Slack *SlackOptions `json:"slack"`
}

type Notifier interface {
	Send(title string, body string, recipient string) error
}

func GetAll(config Config) map[string]Notifier {
	res := make(map[string]Notifier)
	if config.Email != nil {
		res["email"] = NewEmailNotifier(*config.Email)
	}
	if config.Slack != nil {
		res["slack"] = NewSlackNotifier(*config.Slack)
	}
	return res
}
