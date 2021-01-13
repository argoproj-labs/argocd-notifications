package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	texttemplate "text/template"

	"github.com/ghodss/yaml"
)

type Notification struct {
	Message  string                `json:"message,omitempty"`
	Email    *EmailNotification    `json:"email,omitempty"`
	Slack    *SlackNotification    `json:"slack,omitempty"`
	Webhook  WebhookNotifications  `json:"webhook,omitempty"`
	Opsgenie *OpsgenieNotification `json:"opsgenie,omitempty"`
}

// Destination holds notification destination details
type Destination struct {
	Service   string `json:"service"`
	Recipient string `json:"recipient"`
}

func (n *Notification) GetTemplater(name string, f texttemplate.FuncMap) (Templater, error) {
	var sources []TemplaterSource
	if n.Slack != nil {
		sources = append(sources, n.Slack)
	}
	if n.Email != nil {
		sources = append(sources, n.Email)
	}
	if n.Webhook != nil {
		sources = append(sources, n.Webhook)
	}

	if n.Opsgenie != nil {
		sources = append(sources, n.Opsgenie)
	}

	return n.getTemplater(name, f, sources)
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
	case "telegram":
		var opts TelegramOptions
		if err := yaml.Unmarshal(optsData, &opts); err != nil {
			return nil, err
		}
		return NewTelegramService(opts), nil
	default:
		return nil, fmt.Errorf("service type '%s' is not supported", serviceType)
	}
}

func (n *Notification) Preview() string {
	preview := ""
	switch {
	case n.Message != "":
		preview = n.Message
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

func (n *Notification) getTemplater(name string, f texttemplate.FuncMap, sources []TemplaterSource) (Templater, error) {
	message, err := texttemplate.New(name).Funcs(f).Parse(n.Message)
	if err != nil {
		return nil, err
	}

	templaters := []Templater{func(notification *Notification, vars map[string]interface{}) error {
		var messageData bytes.Buffer
		if err := message.Execute(&messageData, vars); err != nil {
			return err
		}
		if val := messageData.String(); val != "" {
			notification.Message = messageData.String()
		}

		return nil
	}}

	for _, src := range sources {
		t, err := src.GetTemplater(name, f)
		if err != nil {
			return nil, err
		}
		templaters = append(templaters, t)
	}

	return func(notification *Notification, vars map[string]interface{}) error {
		for _, t := range templaters {
			if err := t(notification, vars); err != nil {
				return err
			}
		}
		return nil
	}, nil
}

type Templater func(notification *Notification, vars map[string]interface{}) error

type TemplaterSource interface {
	GetTemplater(name string, f texttemplate.FuncMap) (Templater, error)
}
