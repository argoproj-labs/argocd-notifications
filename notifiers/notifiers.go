package notifiers

type Config struct {
	Email    *EmailOptions    `json:"email"`
	Slack    *SlackOptions    `json:"slack"`
	Opsgenie *OpsgenieOptions `json:"opsgenie"`
	Grafana  *GrafanaOptions  `json:"grafana"`
	Webhook  *WebhookOptions  `json:"webhook"`
}

type SlackSpecific struct {
	Attachments string `json:"attachments,omitempty"`
	Blocks      string `json:"blocks,omitempty"`
}

type Notification struct {
	Title   string `json:"title,omitempty"`
	Body    string `json:"body,omitempty"`
	Slack   *SlackNotification
	Webhook map[string]WebhookNotification `json:"webhook,omitempty" patchStrategy:"replace"`
}

//go:generate mockgen -destination=./mocks/notifiers.go -package=mocks github.com/argoproj-labs/argocd-notifications/notifiers Notifier

type Notifier interface {
	Send(notification Notification, recipient string) error
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

	if config.Grafana != nil {
		res["grafana"] = NewGrafanaNotifier(*config.Grafana)
	}

	if config.Webhook != nil {
		res["webhook"] = NewWebhookNotifier(*config.Webhook)
	}
	return res
}
