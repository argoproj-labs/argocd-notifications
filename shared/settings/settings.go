package settings

import (
	"encoding/json"

	"github.com/argoproj-labs/argocd-notifications/notifiers"
	"github.com/argoproj-labs/argocd-notifications/shared/text"
	"github.com/argoproj-labs/argocd-notifications/triggers"

	"github.com/ghodss/yaml"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
)

type rawSubscription struct {
	Recipients []string
	Triggers   []string
	Selector   string
}

// DefaultSubscription holds recipients that receives notification by default.
type Subscription struct {
	// Recipients comma separated list of recipients
	Recipients []string
	// Optional trigger name
	Triggers []string
	// Options label selector that limits applied applications
	Selector labels.Selector
}

func (s *Subscription) MatchesTrigger(trigger string) bool {
	if len(s.Triggers) == 0 {
		return true
	}
	for i := range s.Triggers {
		if s.Triggers[i] == trigger {
			return true
		}
	}
	return false
}

func (s *Subscription) UnmarshalJSON(data []byte) error {
	raw := rawSubscription{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	s.Triggers = raw.Triggers
	s.Recipients = raw.Recipients
	selector, err := labels.Parse(raw.Selector)
	if err != nil {
		return err
	}
	s.Selector = selector
	return nil
}

type DefaultSubscriptions []Subscription

// Returns list of recipients for the specified trigger
func (subscriptions DefaultSubscriptions) GetRecipients(trigger string, labels map[string]string) []string {
	var result []string
	for _, s := range subscriptions {
		if s.MatchesTrigger(trigger) && s.Selector.Matches(fields.Set(labels)) {
			result = append(result, s.Recipients...)
		}
	}
	return result
}

type Config struct {
	Triggers      []triggers.NotificationTrigger  `json:"triggers"`
	Templates     []triggers.NotificationTemplate `json:"templates"`
	Context       map[string]string               `json:"context"`
	Subscriptions DefaultSubscriptions            `json:"subscriptions"`
}

// ParseSecret retrieves configured notification services from the provided secret
func ParseSecret(secret *v1.Secret) (notifiersConfig notifiers.Config, err error) {
	notifiersData := secret.Data["notifiers.yaml"]
	err = yaml.Unmarshal(notifiersData, &notifiersConfig)
	if err != nil {
		return notifiers.Config{}, err
	}
	return notifiersConfig, nil
}

// ParseSecret retrieves configured templates and triggers from the provided config map
func ParseConfigMap(configMap *v1.ConfigMap) (cfg *Config, err error) {
	cfg = &Config{}
	if data, ok := configMap.Data["config.yaml"]; ok {
		err = yaml.Unmarshal([]byte(data), &cfg)
		if err != nil {
			return cfg, err
		}
	}
	return cfg, nil
}

func (cfg *Config) Merge(other *Config) *Config {
	triggersMap := map[string]triggers.NotificationTrigger{}
	for i := range cfg.Triggers {
		triggersMap[cfg.Triggers[i].Name] = cfg.Triggers[i]
	}
	for _, item := range other.Triggers {
		if existing, ok := triggersMap[item.Name]; ok {
			existing.Condition = text.Coalesce(item.Condition, existing.Condition)
			existing.Template = text.Coalesce(item.Template, existing.Template)
			if item.Enabled != nil {
				existing.Enabled = item.Enabled
			}
			triggersMap[item.Name] = existing
		} else {
			triggersMap[item.Name] = item
		}
	}

	templatesMap := map[string]triggers.NotificationTemplate{}
	for i := range cfg.Templates {
		templatesMap[cfg.Templates[i].Name] = cfg.Templates[i]
	}
	for _, item := range other.Templates {
		if existing, ok := templatesMap[item.Name]; ok {
			existing.Body = text.Coalesce(item.Body, existing.Body)
			existing.Title = text.Coalesce(item.Title, existing.Title)
			if item.Slack != nil {
				if existing.Slack == nil {
					existing.Slack = &notifiers.SlackNotification{Blocks: item.Slack.Blocks, Attachments: item.Slack.Attachments}
				} else {
					existing.Slack.Attachments = text.Coalesce(item.Slack.Attachments, existing.Slack.Attachments)
					existing.Slack.Blocks = text.Coalesce(item.Slack.Blocks, existing.Slack.Blocks)
				}
			}
			templatesMap[item.Name] = existing
		} else {
			templatesMap[item.Name] = item
		}
	}

	contextValues := map[string]string{}
	for k, v := range cfg.Context {
		contextValues[k] = v
	}
	for k, v := range other.Context {
		contextValues[k] = v
	}
	res := &Config{}
	for k := range triggersMap {
		res.Triggers = append(res.Triggers, triggersMap[k])
	}
	for k := range templatesMap {
		res.Templates = append(res.Templates, templatesMap[k])
	}
	res.Context = contextValues
	return res
}
