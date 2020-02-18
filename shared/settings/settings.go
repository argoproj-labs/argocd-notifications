package settings

import (
	"github.com/argoproj-labs/argocd-notifications/notifiers"
	"github.com/argoproj-labs/argocd-notifications/triggers"

	"github.com/ghodss/yaml"
	v1 "k8s.io/api/core/v1"
)

type Config struct {
	Triggers  []triggers.NotificationTrigger  `json:"triggers"`
	Templates []triggers.NotificationTemplate `json:"templates"`
	Context   map[string]string               `json:"context"`
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

func coalesce(first string, other ...string) string {
	res := first
	for i := range other {
		if res != "" {
			break
		}
		res = other[i]
	}
	return res
}

func (cfg *Config) Merge(other *Config) *Config {
	triggersMap := map[string]triggers.NotificationTrigger{}
	for i := range cfg.Triggers {
		triggersMap[cfg.Triggers[i].Name] = cfg.Triggers[i]
	}
	for _, item := range other.Triggers {
		if existing, ok := triggersMap[item.Name]; ok {
			existing.Condition = coalesce(item.Condition, existing.Condition)
			existing.Template = coalesce(item.Template, existing.Template)
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
			existing.Body = coalesce(item.Body, existing.Body)
			existing.Title = coalesce(item.Title, existing.Title)
			if item.Slack != nil {
				if existing.Slack == nil {
					existing.Slack = &notifiers.SlackSpecific{Blocks: item.Slack.Blocks, Attachments: item.Slack.Attachments}
				} else {
					existing.Slack.Attachments = coalesce(item.Slack.Attachments, existing.Slack.Attachments)
					existing.Slack.Blocks = coalesce(item.Slack.Blocks, existing.Slack.Blocks)
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
