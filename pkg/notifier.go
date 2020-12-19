package pkg

import (
	"fmt"

	"github.com/argoproj-labs/argocd-notifications/pkg/services"
	"github.com/argoproj-labs/argocd-notifications/pkg/templates"
)

const (
	serviceTypeVarName = "serviceType"
)

//go:generate mockgen -destination=./mocks/mocks.go -package=mocks github.com/argoproj-labs/argocd-notifications/pkg Notifier

// Notifier provides high level interface to send notifications and manage notification services
type Notifier interface {
	Send(vars map[string]interface{}, template string, dest services.Destination) error
	AddService(name string, service services.NotificationService)
	GetServices() map[string]services.NotificationService
}

type notifier struct {
	services  map[string]services.NotificationService
	templates map[string]templates.Template
}

// AddService adds new service with the specified name
func (n *notifier) AddService(name string, service services.NotificationService) {
	n.services[name] = service
}

// GetServices returns map of registered services
func (n *notifier) GetServices() map[string]services.NotificationService {
	return n.services
}

// Send sends notification using specified service and template to the specified destination
func (n *notifier) Send(vars map[string]interface{}, templateName string, dest services.Destination) error {
	service, ok := n.services[dest.Service]
	if !ok {
		return fmt.Errorf("service '%s' is not supported", dest.Service)
	}
	template, ok := n.templates[templateName]
	if !ok {
		return fmt.Errorf("template '%s' is not supported", templateName)
	}

	in := make(map[string]interface{})
	for k := range vars {
		in[k] = vars[k]
	}
	in[serviceTypeVarName] = dest.Service
	notification, err := template.FormatNotification(in)
	if err != nil {
		return err
	}
	return service.Send(*notification, dest)
}

// NewNotifier creates new notifier instance using provided config
func NewNotifier(cfg Config) (*notifier, error) {
	n := notifier{
		map[string]services.NotificationService{},
		map[string]templates.Template{},
	}
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
