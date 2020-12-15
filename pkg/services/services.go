package services

import (
	"fmt"

	"github.com/ghodss/yaml"
)

type Notification struct {
	Title   string                         `json:"title,omitempty"`
	Body    string                         `json:"body,omitempty"`
	Slack   *SlackNotification             `json:"slack,omitempty"`
	Webhook map[string]WebhookNotification `json:"webhook,omitempty" patchStrategy:"replace"`
}

type NotificationService interface {
	Send(notification Notification, recipient string) error
}

func NewService(serviceType string, optsData []byte) (NotificationService, error) {
	switch serviceType {
	case "email":
		var opts EmailOptions
		if err := yaml.Unmarshal(optsData, &opts); err != nil {
			return nil, err
		}
		return NewEmailService(opts), nil
	case "slack":
		var opts SlackOptions
		if err := yaml.Unmarshal(optsData, &opts); err != nil {
			return nil, err
		}
		return NewSlackService(opts), nil
	case "grafana":
		var opts GrafanaOptions
		if err := yaml.Unmarshal(optsData, &opts); err != nil {
			return nil, err
		}
		return NewGrafanaService(opts), nil
	case "opsgenie":
		var opts OpsgenieOptions
		if err := yaml.Unmarshal(optsData, &opts); err != nil {
			return nil, err
		}
		return NewOpsgenieService(opts), nil
	case "webhook":
		var opts WebhookOptions
		if err := yaml.Unmarshal(optsData, &opts); err != nil {
			return nil, err
		}
		return NewWebhookService(opts), nil
	default:
		return nil, fmt.Errorf("service type '%s' is not supported", serviceType)
	}
}
