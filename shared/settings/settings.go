package settings

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/argoproj-labs/argocd-notifications/notifiers"
	"github.com/argoproj-labs/argocd-notifications/shared/argocd"
	"github.com/argoproj-labs/argocd-notifications/triggers"

	"github.com/ghodss/yaml"
	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/strategicpatch"
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

func (s *Subscription) MarshalJSON() ([]byte, error) {
	raw := rawSubscription{
		Triggers:   s.Triggers,
		Recipients: s.Recipients,
	}
	if s.Selector != nil {
		raw.Selector = s.Selector.String()
	}
	return json.Marshal(raw)
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
	Triggers      []triggers.NotificationTrigger  `json:"triggers,omitempty" patchStrategy:"merge" patchMergeKey:"name"`
	Templates     []triggers.NotificationTemplate `json:"templates,omitempty" patchStrategy:"merge" patchMergeKey:"name"`
	Context       map[string]string               `json:"context,omitempty"`
	Subscriptions DefaultSubscriptions            `json:"subscriptions,omitempty"`
}

func createNotifier(notifierType string, config []byte) (notifiers.Notifier, error) {
	switch notifierType {
	case "email":
		var opts notifiers.EmailOptions
		if err := yaml.Unmarshal(config, &opts); err != nil {
			return nil, err
		}
		return notifiers.NewEmailNotifier(opts), nil
	case "slack":
		var opts notifiers.SlackOptions
		if err := yaml.Unmarshal(config, &opts); err != nil {
			return nil, err
		}
		return notifiers.NewSlackNotifier(opts), nil
	case "grafana":
		var opts notifiers.GrafanaOptions
		if err := yaml.Unmarshal(config, &opts); err != nil {
			return nil, err
		}
		return notifiers.NewGrafanaNotifier(opts), nil
	case "opsgenie":
		var opts notifiers.OpsgenieOptions
		if err := yaml.Unmarshal(config, &opts); err != nil {
			return nil, err
		}
		return notifiers.NewOpsgenieNotifier(opts), nil
	case "webhook":
		var opts notifiers.WebhookOptions
		if err := yaml.Unmarshal(config, &opts); err != nil {
			return nil, err
		}
		return notifiers.NewWebhookNotifier(opts), nil
	default:
		return nil, fmt.Errorf("notifier type '%s' is not supported", notifierType)
	}
}

// ParseSecret retrieves configured notification services from the provided secret
func ParseSecret(secret *v1.Secret) (map[string]notifiers.Notifier, error) {
	res := map[string]notifiers.Notifier{}
	if notifiersData, ok := secret.Data["notifiers.yaml"]; ok {
		var legacyConf legacyConfig
		err := yaml.Unmarshal(notifiersData, &legacyConf)
		if err != nil {
			return nil, err
		}
		legacyConf.addNotifiers(res)
	}
	for k, v := range secret.Data {
		if strings.HasPrefix(k, "notifier.") {
			name := ""
			notifierType := ""
			parts := strings.Split(k, ".")
			if len(parts) == 3 {
				notifierType, name = parts[1], parts[2]
			} else if len(parts) == 2 {
				notifierType, name = parts[1], parts[1]
			} else {
				return nil, fmt.Errorf("invalid notifier key; expected 'notifier.<type>(.<name>)' but got '%s'", k)
			}

			n, err := createNotifier(notifierType, v)
			if err != nil {
				return nil, err
			}
			res[name] = n
		}
	}
	return res, nil
}

// ParseConfigMap retrieves configured templates and triggers from the provided config map
func ParseConfigMap(configMap *v1.ConfigMap) (*Config, error) {
	legacyCfg := &Config{}
	if data, ok := configMap.Data["config.yaml"]; ok {
		err := yaml.Unmarshal([]byte(data), &legacyCfg)
		if err != nil {
			return legacyCfg, fmt.Errorf("Failed to read config.yaml key from configmap: %v", err)
		}
	}

	cfg := &Config{}
	// read all the keys in format of templates.%s and triggers.%s
	// to create config
	for k, v := range configMap.Data {
		parts := strings.Split(k, ".")
		switch {
		case k == "subscriptions":
			if err := yaml.Unmarshal([]byte(v), &cfg.Subscriptions); err != nil {
				return cfg, err
			}
		case k == "context":
			if err := yaml.Unmarshal([]byte(v), &cfg.Context); err != nil {
				return cfg, err
			}
		case strings.HasPrefix(k, "template."):
			name := strings.Join(parts[1:], ".")
			tmpl := triggers.NotificationTemplate{}
			if err := yaml.Unmarshal([]byte(v), &tmpl); err != nil {
				return cfg, fmt.Errorf("Failed to unmarshal template %s: %v", name, err)
			}
			tmpl.Name = name
			cfg.Templates = append(cfg.Templates, tmpl)
		case strings.HasPrefix(k, "trigger."):
			name := strings.Join(parts[1:], ".")
			trigger := triggers.NotificationTrigger{}
			if err := yaml.Unmarshal([]byte(v), &trigger); err != nil {
				return cfg, fmt.Errorf("Failed to unmarshal trigger %s: %v", name, err)
			}
			trigger.Name = name
			cfg.Triggers = append(cfg.Triggers, trigger)
		default:
			log.Warnf("Key %s does not match to any pattern, ignored", k)
		}
	}

	return legacyCfg.Merge(cfg)
}

func (cfg *Config) Merge(other *Config) (*Config, error) {
	origData, err := json.Marshal(cfg)
	if err != nil {
		return nil, err
	}
	otherData, err := json.Marshal(other)
	if err != nil {
		return nil, err
	}

	mergedData, err := strategicpatch.StrategicMergePatch(origData, otherData, &Config{})
	if err != nil {
		return nil, err
	}

	res := &Config{}
	err = json.Unmarshal(mergedData, res)
	if err != nil {
		return nil, err
	}
	return res, nil
}

// ParseConfig parses notifications configuration from the provided config map and secret.
func ParseConfig(configMap *v1.ConfigMap, secret *v1.Secret, defaultCfg Config, argocdService argocd.Service) (map[string]triggers.Trigger, map[string]notifiers.Notifier, *Config, error) {
	cfg, err := ParseConfigMap(configMap)
	if err != nil {
		return nil, nil, nil, err
	}
	cfg, err = defaultCfg.Merge(cfg)
	if err != nil {
		return nil, nil, nil, err
	}
	t, err := triggers.GetTriggers(cfg.Templates, cfg.Triggers, argocdService)
	if err != nil {
		return nil, nil, nil, err
	}
	n, err := ParseSecret(secret)
	if err != nil {
		return nil, nil, nil, err
	}
	return t, n, cfg, nil
}
