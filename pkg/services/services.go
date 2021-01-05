package services

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/ghodss/yaml"
)

type Notification struct {
	Title   string                         `json:"title,omitempty"`
	Body    string                         `json:"body,omitempty"`
	Slack   *SlackNotification             `json:"slack,omitempty"`
	Webhook map[string]WebhookNotification `json:"webhook,omitempty" patchStrategy:"replace"`
}

func (n *Notification) Preview() string {
	preview := ""
	switch {
	case n.Title != "":
		preview = n.Title
	case n.Body != "":
		preview = n.Body
	default:
		if yamlData, err := json.Marshal(n); err != nil {
			preview = "failed to generate preview"
		} else {
			preview = string(yamlData)
		}
	}
	preview = strings.Split(preview, "\n")[0]
	if len(preview) > 100 {
		preview = preview[:99] + "..."
	}
	return preview
}

// Destination holds notification destination details
type Destination struct {
	Service   string `json:"service"`
	Recipient string `json:"recipient"`
}

//go:generate mockgen -destination=./mocks/mocks.go -package=mocks github.com/argoproj-labs/argocd-notifications/pkg/services NotificationService

// NotificationService defines notification service interface
type NotificationService interface {
	Send(notification Notification, dest Destination) error
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
