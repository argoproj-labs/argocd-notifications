package legacy

import (
	"encoding/json"
	"fmt"

	"github.com/argoproj-labs/argocd-notifications/shared/settings"

	"github.com/argoproj-labs/argocd-notifications/pkg/triggers"

	jsonpatch "github.com/evanphx/json-patch"

	"github.com/ghodss/yaml"
	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"

	"github.com/argoproj-labs/argocd-notifications/pkg"
	"github.com/argoproj-labs/argocd-notifications/pkg/services"
)

type legacyTemplate struct {
	Name  string `json:"name,omitempty"`
	Title string `json:"subject,omitempty"`
	Body  string `json:"body,omitempty"`
	services.Notification
}

type legacyTrigger struct {
	Name        string `json:"name,omitempty"`
	Condition   string `json:"condition,omitempty"`
	Description string `json:"description,omitempty"`
	Template    string `json:"template,omitempty"`
	Enabled     *bool  `json:"enabled,omitempty"`
}

type legacyConfig struct {
	Triggers      []legacyTrigger               `json:"triggers,omitempty"`
	Templates     []legacyTemplate              `json:"templates,omitempty"`
	Context       map[string]string             `json:"context,omitempty"`
	Subscriptions settings.DefaultSubscriptions `json:"subscriptions,omitempty"`
}

type legacyWebhookOptions struct {
	services.WebhookOptions
	Name string `json:"name"`
}

type legacyServicesConfig struct {
	Email    *services.EmailOptions    `json:"email"`
	Slack    *services.SlackOptions    `json:"slack"`
	Opsgenie *services.OpsgenieOptions `json:"opsgenie"`
	Grafana  *services.GrafanaOptions  `json:"grafana"`
	Webhook  []legacyWebhookOptions    `json:"webhook"`
}

func mergePatch(orig interface{}, other interface{}) error {
	origData, err := json.Marshal(orig)
	if err != nil {
		return err
	}
	otherData, err := json.Marshal(other)
	if err != nil {
		return err
	}

	if string(otherData) == "null" {
		return nil
	}

	mergedData, err := jsonpatch.MergePatch(origData, otherData)
	if err != nil {
		return err
	}
	return json.Unmarshal(mergedData, orig)
}

func (legacy legacyConfig) merge(cfg *settings.Config) error {
	if err := mergePatch(&cfg.Context, &legacy.Context); err != nil {
		return err
	}
	if err := mergePatch(&cfg.Subscriptions, &legacy.Subscriptions); err != nil {
		return err
	}

	for _, template := range legacy.Templates {
		t, ok := cfg.Templates[template.Name]
		if ok {
			if err := mergePatch(&t, &template.Notification); err != nil {
				return err
			}
		}
		if template.Title != "" {
			if template.Notification.Email == nil {
				template.Notification.Email = &services.EmailNotification{}
			}
			template.Notification.Email.Subject = template.Title
		}
		if template.Body != "" {
			template.Notification.Message = template.Body
		}
		cfg.Templates[template.Name] = template.Notification
	}

	for _, trigger := range legacy.Triggers {
		if trigger.Enabled != nil && *trigger.Enabled {
			cfg.DefaultTriggers = append(cfg.DefaultTriggers, trigger.Name)
		}
		var firstCondition triggers.Condition
		t, ok := cfg.Triggers[trigger.Name]
		if !ok || len(t) == 0 {
			t = []triggers.Condition{firstCondition}
		} else {
			firstCondition = t[0]
		}

		if trigger.Condition != "" {
			firstCondition.When = trigger.Condition
		}
		if trigger.Template != "" {
			firstCondition.Send = []string{trigger.Template}
		}
		if trigger.Description != "" {
			firstCondition.Description = trigger.Description
		}
		t[0] = firstCondition
		cfg.Triggers[trigger.Name] = t
	}

	return nil
}

func (c *legacyServicesConfig) merge(cfg *pkg.Config) {
	if c.Email != nil {
		cfg.Services["email"] = func() (services.NotificationService, error) {
			return services.NewEmailService(*c.Email), nil
		}
	}
	if c.Slack != nil {
		cfg.Services["slack"] = func() (services.NotificationService, error) {
			return services.NewSlackService(*c.Slack), nil
		}
	}
	if c.Grafana != nil {
		cfg.Services["grafana"] = func() (services.NotificationService, error) {
			return services.NewGrafanaService(*c.Grafana), nil
		}
	}
	if c.Opsgenie != nil {
		cfg.Services["opsgenie"] = func() (services.NotificationService, error) {
			return services.NewOpsgenieService(*c.Opsgenie), nil
		}
	}
	for i := range c.Webhook {
		opts := c.Webhook[i]
		cfg.Services[fmt.Sprintf(opts.Name)] = func() (services.NotificationService, error) {
			return services.NewWebhookService(opts.WebhookOptions), nil
		}
	}
}

// ApplyLegacyConfig settings specified using deprecated config map and secret keys
func ApplyLegacyConfig(cfg *settings.Config, cm *v1.ConfigMap, secret *v1.Secret) error {
	if notifiersData, ok := secret.Data["notifiers.yaml"]; ok && len(notifiersData) > 0 {
		log.Warn("Key 'notifiers.yaml' in Secret is deprecated, please migrate to new settings")
		legacyServices := &legacyServicesConfig{}
		err := yaml.Unmarshal(notifiersData, legacyServices)
		if err != nil {
			return err
		}
		legacyServices.merge(&cfg.Config)
	}

	if configData, ok := cm.Data["config.yaml"]; ok && configData != "" {
		log.Warn("Key 'config.yaml' in ConfigMap is deprecated, please migrate to new settings")
		legacyCfg := &legacyConfig{}
		err := yaml.Unmarshal([]byte(configData), legacyCfg)
		if err != nil {
			return err
		}
		err = legacyCfg.merge(cfg)
		if err != nil {
			return err
		}
	}
	return nil
}
