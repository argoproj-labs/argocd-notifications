package notifiers

type Config struct {
	Email *EmailOptions       `json:"email"`
	Slack *SlackOptions       `json:"slack"`
	Opsgenie *OpsgenieOptions `json:"opsgenie"`
}

//go:generate mockgen -destination=./mocks/notifiers.go -package=mocks github.com/argoproj-labs/argocd-notifications/notifiers Notifier

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
	if config.Opsgenie != nil {
		res["opsgenie"] = NewOpsgenieNotifier(*config.Opsgenie)
	}
	return res
}
