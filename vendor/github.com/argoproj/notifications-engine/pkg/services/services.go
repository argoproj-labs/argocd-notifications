package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	texttemplate "text/template"
	_ "time/tzdata"

	"github.com/ghodss/yaml"
)

type Notification struct {
	Message    string                  `json:"message,omitempty"`
	Email      *EmailNotification      `json:"email,omitempty"`
	Slack      *SlackNotification      `json:"slack,omitempty"`
	Mattermost *MattermostNotification `json:"mattermost,omitempty"`
	RocketChat *RocketChatNotification `json:"rocketchat,omitempty"`
	Teams      *TeamsNotification      `json:"teams,omitempty"`
	Webhook    WebhookNotifications    `json:"webhook,omitempty"`
	Opsgenie   *OpsgenieNotification   `json:"opsgenie,omitempty"`
	GitHub     *GitHubNotification     `json:"github,omitempty"`
}

// Destinations holds notification destinations group by trigger
type Destinations map[string][]Destination

func (s Destinations) Merge(other Destinations) {
	for k := range other {
		s[k] = append(s[k], other[k]...)
	}
}

func (s Destinations) Dedup() Destinations {
	for k, v := range s {
		set := map[Destination]bool{}
		var dedup []Destination
		for _, dest := range v {
			if !set[dest] {
				set[dest] = true
				dedup = append(dedup, dest)
			}
		}
		s[k] = dedup
	}
	return s
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
	if n.Mattermost != nil {
		sources = append(sources, n.Mattermost)
	}
	if n.RocketChat != nil {
		sources = append(sources, n.RocketChat)
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
	if n.GitHub != nil {
		sources = append(sources, n.GitHub)
	}

	if n.Teams != nil {
		sources = append(sources, n.Teams)
	}

	return n.getTemplater(name, f, sources)
}

//go:generate mockgen -destination=./mocks/mocks.go -package=mocks github.com/argoproj/notifications-engine/pkg/services NotificationService

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
	case "mattermost":
		var opts MattermostOptions
		if err := yaml.Unmarshal(optsData, &opts); err != nil {
			return nil, err
		}
		return NewMattermostService(opts), nil
	case "rocketchat":
		var opts RocketChatOptions
		if err := yaml.Unmarshal(optsData, &opts); err != nil {
			return nil, err
		}
		return NewRocketChatService(opts), nil
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
	case "github":
		var opts GitHubOptions
		if err := yaml.Unmarshal(optsData, &opts); err != nil {
			return nil, err
		}
		return NewGitHubService(opts)
	case "teams":
		var opts TeamsOptions
		if err := yaml.Unmarshal(optsData, &opts); err != nil {
			return nil, err
		}
		return NewTeamsService(opts), nil
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
