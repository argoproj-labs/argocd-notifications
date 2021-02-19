package pkg

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/argoproj-labs/argocd-notifications/pkg/services"
	"github.com/argoproj-labs/argocd-notifications/pkg/triggers"

	"github.com/ghodss/yaml"
	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
)

// Subscriptions holds an information what explains when, where and how a notification should be sent
type Subscriptions map[string][]services.Destination

func (s Subscriptions) Merge(other Subscriptions) {
	for k := range other {
		s[k] = append(s[k], other[k]...)
	}
}

func (s Subscriptions) Dedup() Subscriptions {
	for k, v := range s {
		set := map[services.Destination]bool{}
		var dedup []services.Destination
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

type ServiceFactory func() (services.NotificationService, error)

// Config holds settings required to create new api
type Config struct {
	Services  map[string]ServiceFactory
	Triggers  map[string][]triggers.Condition
	Templates map[string]services.Notification
}

var keyPattern = regexp.MustCompile(`[$][\w-_]+`)

// replaceStringSecret checks if given string is a secret key reference ( starts with $ ) and returns corresponding value from provided map
func replaceStringSecret(val string, secretValues map[string][]byte) string {
	return keyPattern.ReplaceAllStringFunc(val, func(secretKey string) string {
		secretVal, ok := secretValues[secretKey[1:]]
		if !ok {
			log.Warnf("config referenced '%s', but key does not exist in secret", val)
			return secretKey
		}
		return string(secretVal)
	})
}

func replaceServiceConfigSecret(data map[string]interface{}, secretValues map[string][]byte) map[string]interface{} {
	result := map[string]interface{}{}
	for k, v := range data {
		switch v := v.(type) {
		case string:
			result[k] = replaceStringSecret(v, secretValues)
		case []string:
			list := make([]string, len(v))
			for i, item := range v {
				list[i] = replaceStringSecret(item, secretValues)
			}
			result[k] = list
		case map[string]interface{}:
			result[k] = replaceServiceConfigSecret(v, secretValues)
		case []map[string]interface{}:
			list := make([]map[string]interface{}, len(v))
			for i, item := range v {
				list[i] = replaceServiceConfigSecret(item, secretValues)
			}
			result[k] = list
		default:
			result[k] = v
		}
	}

	return result
}

// ParseConfig retrieves Config from given ConfigMap and Secret
func ParseConfig(configMap *v1.ConfigMap, secret *v1.Secret) (*Config, error) {
	cfg := Config{map[string]ServiceFactory{}, map[string][]triggers.Condition{}, map[string]services.Notification{}}
	for k, v := range configMap.Data {
		parts := strings.Split(k, ".")
		switch {
		case strings.HasPrefix(k, "template."):
			name := strings.Join(parts[1:], ".")
			template := services.Notification{}
			if err := yaml.Unmarshal([]byte(v), &template); err != nil {
				return nil, fmt.Errorf("failed to unmarshal template %s: %v", name, err)
			}
			cfg.Templates[name] = template
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

			serviceConfig := map[string]interface{}{}
			if err := yaml.Unmarshal([]byte(v), &serviceConfig); err != nil {
				return nil, fmt.Errorf("failed to unmarshal service %s: %v", serviceType, err)
			}

			serviceConfig = replaceServiceConfigSecret(serviceConfig, secret.Data)
			optsData, err := yaml.Marshal(serviceConfig)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal service %s: %v", serviceType, err)
			}

			cfg.Services[name] = func() (services.NotificationService, error) {
				return services.NewService(serviceType, optsData)
			}
		case strings.HasPrefix(k, "trigger."):
			name := strings.Join(parts[1:], ".")
			var trigger []triggers.Condition
			if err := yaml.Unmarshal([]byte(v), &trigger); err != nil {
				return nil, fmt.Errorf("failed to unmarshal trigger %s: %v", name, err)
			}
			cfg.Triggers[name] = trigger
		}
	}
	return &cfg, nil
}
