package pkg

import (
	"fmt"
	"strings"

	"github.com/argoproj-labs/argocd-notifications/pkg/services"
	"github.com/argoproj-labs/argocd-notifications/pkg/templates"

	log "github.com/sirupsen/logrus"
)

//go:generate mockgen -destination=./mocks/mocks.go -package=mocks github.com/argoproj-labs/argocd-notifications/pkg Notifier

type Notifier interface {
	Send(vars map[string]interface{}, template string, serviceType string, recipient string) error
	AddService(name string, service services.NotificationService)
	GetServices() map[string]services.NotificationService
}

const (
	serviceTypeVarName = "serviceType"
)

type notifier struct {
	services  map[string]services.NotificationService
	templates map[string]templates.Template
}

func (n *notifier) AddService(name string, service services.NotificationService) {
	n.services[name] = service
}

func (n *notifier) GetServices() map[string]services.NotificationService {
	return n.services
}

func (n *notifier) Send(vars map[string]interface{}, templateName string, serviceType string, recipient string) error {
	service, ok := n.services[serviceType]
	if !ok {
		return fmt.Errorf("service '%s' is not supported", serviceType)
	}
	template, ok := n.templates[templateName]
	if !ok {
		return fmt.Errorf("template '%s' is not supported", templateName)
	}

	in := make(map[string]interface{})
	for k := range vars {
		in[k] = vars[k]
	}
	in[serviceTypeVarName] = serviceType
	notification, err := template.FormatNotification(in)
	if err != nil {
		return err
	}
	return service.Send(*notification, recipient)
}

// replaceStringSecret checks if given string is a secret key reference ( starts with $ ) and returns corresponding value from provided map
func replaceStringSecret(val string, secretValues map[string]string) string {
	if val == "" || !strings.HasPrefix(val, "$") {
		return val
	}
	secretKey := val[1:]
	secretVal, ok := secretValues[secretKey]
	if !ok {
		log.Warnf("config referenced '%s', but key does not exist in secret", val)
		return val
	}
	return secretVal
}

func NewNotifier(cfg Config) (*notifier, error) {
	n := notifier{map[string]services.NotificationService{}, map[string]templates.Template{}}
	for k, v := range cfg.Services {
		svc, err := v()
		if err != nil {
			return nil, err
		}
		n.services[k] = svc
	}

	for _, v := range cfg.Templates {
		tmpl, err := templates.NewTemplate(v)
		if err != nil {
			return nil, err
		}
		n.templates[v.Name] = tmpl
	}

	return &n, nil
}
