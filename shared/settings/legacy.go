package settings

import (
	"encoding/json"

	"github.com/ghodss/yaml"
	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/strategicpatch"

	"github.com/argoproj-labs/argocd-notifications/pkg"
	"github.com/argoproj-labs/argocd-notifications/pkg/services"
	"github.com/argoproj-labs/argocd-notifications/pkg/templates"
	"github.com/argoproj-labs/argocd-notifications/triggers"
)

type legacyConfig struct {
	Triggers      []triggers.NotificationTrigger   `json:"triggers,omitempty" patchStrategy:"merge" patchMergeKey:"name"`
	Templates     []templates.NotificationTemplate `json:"templates,omitempty" patchStrategy:"merge" patchMergeKey:"name"`
	Context       map[string]string                `json:"context,omitempty"`
	Subscriptions DefaultSubscriptions             `json:"subscriptions,omitempty"`
}

type legacyServicesConfig struct {
	Email    *services.EmailOptions    `json:"email"`
	Slack    *services.SlackOptions    `json:"slack"`
	Opsgenie *services.OpsgenieOptions `json:"opsgenie"`
	Grafana  *services.GrafanaOptions  `json:"grafana"`
	Webhook  *services.WebhookOptions  `json:"webhook"`
}

func (c legacyConfig) merge(cfg *Config) error {
	origData, err := json.Marshal(&legacyConfig{
		Triggers:      cfg.TriggersSettings,
		Templates:     cfg.Templates,
		Context:       cfg.Context,
		Subscriptions: cfg.Subscriptions,
	})
	if err != nil {
		return err
	}
	otherData, err := json.Marshal(c)
	if err != nil {
		return err
	}

	mergedData, err := strategicpatch.StrategicMergePatch(origData, otherData, &legacyConfig{})
	if err != nil {
		return err
	}
	merged := &legacyConfig{}
	err = json.Unmarshal(mergedData, merged)
	if err != nil {
		return err
	}
	cfg.Templates = merged.Templates
	cfg.TriggersSettings = merged.Triggers
	cfg.Subscriptions = merged.Subscriptions
	cfg.Context = merged.Context
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
	if c.Webhook != nil {
		cfg.Services["webhook"] = func() (services.NotificationService, error) {
			return services.NewWebhookService(*c.Webhook), nil
		}
	}
}

// mergeLegacyConfig settings specified using deprecated config map and secret keys
func mergeLegacyConfig(cfg *Config, cm *v1.ConfigMap, secret *v1.Secret) error {
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
