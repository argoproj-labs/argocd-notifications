package api

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/argoproj/notifications-engine/pkg/services"
	"github.com/argoproj/notifications-engine/pkg/subscriptions"
	"github.com/argoproj/notifications-engine/pkg/triggers"

	"github.com/ghodss/yaml"
	log "github.com/sirupsen/logrus"
	yaml3 "gopkg.in/yaml.v3"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/fields"
)

type ServiceFactory func() (services.NotificationService, error)

// Config holds settings required to create new api
type Config struct {
	Services  map[string]ServiceFactory
	Triggers  map[string][]triggers.Condition
	Templates map[string]services.Notification
	// Subscriptions holds list of default application subscriptions
	Subscriptions subscriptions.DefaultSubscriptions
	// DefaultTriggers holds list of triggers that is used by default if subscriber don't specify trigger
	DefaultTriggers []string
	// ServiceDefaultTriggers holds list of default triggers per service
	ServiceDefaultTriggers map[string][]string
}

// Returns list of destinations for the specified trigger
func (cfg Config) GetGlobalDestinations(labels map[string]string) services.Destinations {
	dests := services.Destinations{}
	for _, s := range cfg.Subscriptions {
		triggers := s.Triggers
		if len(triggers) == 0 {
			triggers = cfg.DefaultTriggers
		}
		for _, trigger := range triggers {
			if s.MatchesTrigger(trigger) && s.Selector.Matches(fields.Set(labels)) {
				for _, recipient := range s.Recipients {
					parts := strings.Split(recipient, ":")
					dest := services.Destination{Service: parts[0]}
					if len(parts) > 1 {
						dest.Recipient = parts[1]
					}
					dests[trigger] = append(dests[trigger], dest)
				}
			}
		}
	}
	return dests
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

// ParseConfig retrieves Config from given ConfigMap and Secret
func ParseConfig(configMap *v1.ConfigMap, secret *v1.Secret) (*Config, error) {
	cfg := Config{
		Services:               map[string]ServiceFactory{},
		Triggers:               map[string][]triggers.Condition{},
		ServiceDefaultTriggers: map[string][]string{},
		Templates:              map[string]services.Notification{},
	}
	if subscriptionYaml, ok := configMap.Data["subscriptions"]; ok {
		if err := yaml.Unmarshal([]byte(subscriptionYaml), &cfg.Subscriptions); err != nil {
			return nil, err
		}
	}

	if defaultTriggersYaml, ok := configMap.Data["defaultTriggers"]; ok {
		if err := yaml.Unmarshal([]byte(defaultTriggersYaml), &cfg.DefaultTriggers); err != nil {
			return nil, err
		}
	}

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

			optsData, err := replaceServiceConfigSecrets(v, secret)
			if err != nil {
				return nil, fmt.Errorf("failed to render service configuration %s: %v", serviceType, err)
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
		case strings.HasPrefix(k, "defaultTriggers."):
			name := strings.Join(parts[1:], ".")
			var defaultTriggers []string
			if err := yaml.Unmarshal([]byte(v), &defaultTriggers); err != nil {
				return nil, fmt.Errorf("failed to unmarshal default trigger %s: %v", name, err)
			}
			cfg.ServiceDefaultTriggers[name] = defaultTriggers
		}
	}
	return &cfg, nil
}

func replaceServiceConfigSecrets(inputYaml string, secret *v1.Secret) ([]byte, error) {
	var node yaml3.Node
	err := yaml3.Unmarshal([]byte(inputYaml), &node)
	if err != nil {
		return nil, err
	}

	walkYamlDocument(&node, func(visitedNode *yaml3.Node) {
		if visitedNode.Kind == yaml3.ScalarNode && visitedNode.Tag == "!!str" {
			visitedNode.Value = replaceStringSecret(visitedNode.Value, secret.Data)
		}
	})

	if result, err := yaml3.Marshal(&node); err != nil {
		return nil, err
	} else {
		return result, nil
	}
}

func walkYamlDocument(node *yaml3.Node, visitor func(*yaml3.Node)) {
	visitor(node)

	for _, node := range node.Content {
		walkYamlDocument(node, visitor)
	}
}
