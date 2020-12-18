package pkg

import (
	"fmt"
	"strings"

	"github.com/argoproj-labs/argocd-notifications/pkg/services"
	"github.com/argoproj-labs/argocd-notifications/pkg/templates"
	"github.com/ghodss/yaml"
	v1 "k8s.io/api/core/v1"
)

// NotificationSubscription holds an information what explains when, where and how a notification should be sent
type NotificationSubscription struct {
	When string                 `json:"when"`
	Send string                 `json:"send"`
	To   []services.Destination `json:"to"`
}

type ServiceFactory func() (services.NotificationService, error)

// Config holds settings required to create new notifier
type Config struct {
	Templates []templates.NotificationTemplate
	Services  map[string]ServiceFactory
}

// ParseConfig retrieves Config from given ConfigMap and Secret
func ParseConfig(configMap *v1.ConfigMap, secret *v1.Secret) (*Config, error) {
	for k, v := range configMap.Data {
		configMap.Data[k] = replaceStringSecret(v, secret.StringData)
	}

	cfg := Config{Services: map[string]ServiceFactory{}}
	for k, v := range configMap.Data {
		parts := strings.Split(k, ".")
		switch {
		case strings.HasPrefix(k, "template."):
			name := strings.Join(parts[1:], ".")
			nt := templates.NotificationTemplate{}
			if err := yaml.Unmarshal([]byte(v), &nt); err != nil {
				return nil, fmt.Errorf("failed to unmarshal template %s: %v", name, err)
			}
			nt.Name = name
			cfg.Templates = append(cfg.Templates, nt)
		case strings.HasPrefix(k, "service."):
			name := ""
			serviceType := ""
			parts := strings.Split(k, ".")
			if len(parts) == 3 {
				serviceType, name = parts[1], parts[2]
			} else if len(parts) == 2 {
				serviceType, name = parts[1], parts[1]
			} else {
				return nil, fmt.Errorf("invalid service key; expected 'service.<type>(.<name>)' but got '%s'", k)
			}

			cfg.Services[name] = func() (services.NotificationService, error) {
				return services.NewService(serviceType, []byte(v))
			}
		}
	}
	return &cfg, nil
}
