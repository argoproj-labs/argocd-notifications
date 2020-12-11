package notifiers

type Notification struct {
	Title   string                         `json:"title,omitempty"`
	Body    string                         `json:"body,omitempty"`
	Slack   *SlackNotification             `json:"slack,omitempty"`
	Webhook map[string]WebhookNotification `json:"webhook,omitempty" patchStrategy:"replace"`
}

//go:generate mockgen -destination=./mocks/notifiers.go -package=mocks github.com/argoproj-labs/argocd-notifications/notifiers Notifier

type Notifier interface {
	Send(notification Notification, recipient string) error
}
