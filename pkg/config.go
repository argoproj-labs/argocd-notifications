package pkg

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/argoproj-labs/argocd-notifications/pkg/services"
	"github.com/argoproj-labs/argocd-notifications/pkg/triggers"

	"github.com/ghodss/yaml"
	log "github.com/sirupsen/logrus"
	yaml3 "gopkg.in/yaml.v3"
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
