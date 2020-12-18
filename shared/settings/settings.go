package settings

import (
	"fmt"
	"strings"

	"github.com/argoproj-labs/argocd-notifications/pkg"
	"github.com/argoproj-labs/argocd-notifications/shared/argocd"
	"github.com/argoproj-labs/argocd-notifications/triggers"

	"gopkg.in/yaml.v2"
	v1 "k8s.io/api/core/v1"
)

type Config struct {
	pkg.Config
	Triggers         map[string]triggers.Trigger
	TriggersSettings []triggers.NotificationTrigger
	Context          map[string]string
	Subscriptions    DefaultSubscriptions
	Notifier         pkg.Notifier
	ArgoCDService    argocd.Service
}

// ParseConfigMap retrieves configured templates and triggers from the provided config map
func NewConfig(configMap *v1.ConfigMap, secret *v1.Secret, argocdService argocd.Service) (*Config, error) {
	c, err := pkg.ParseConfig(configMap, secret)
	if err != nil {
		return nil, err
	}
	notifier, err := pkg.NewNotifier(*c)
	if err != nil {
		return nil, err
	}
	cfg := Config{
		Config:   *c,
		Triggers: map[string]triggers.Trigger{},
		Notifier: notifier, ArgoCDService: argocdService,
		Context: map[string]string{
			"argocdUrl": "https://localhost:4000",
		},
	}
	// read all the keys in format of templates.%s and triggers.%s
	// to create config
	for k, v := range configMap.Data {
		parts := strings.Split(k, ".")
		switch {
		case k == "subscriptions":
			if err := yaml.Unmarshal([]byte(v), &cfg.Subscriptions); err != nil {
				return nil, err
			}
		case k == "context":
			ctx := map[string]string{}
			if err := yaml.Unmarshal([]byte(v), &ctx); err != nil {
				return nil, err
			}
			for k, v := range ctx {
				cfg.Context[k] = v
			}
		case strings.HasPrefix(k, "trigger."):
			name := strings.Join(parts[1:], ".")
			nt := triggers.NotificationTrigger{}
			if err := yaml.Unmarshal([]byte(v), &nt); err != nil {
				return nil, fmt.Errorf("Failed to unmarshal trigger %s: %v", name, err)
			}
			nt.Name = name
			trigger, err := triggers.NewTrigger(nt, argocdService)
			if err != nil {
				return nil, fmt.Errorf("Failed to create trigger %s: %v", name, err)
			}
			cfg.Triggers[name] = trigger
			cfg.TriggersSettings = append(cfg.TriggersSettings, nt)
		}
	}

	return &cfg, nil
}
