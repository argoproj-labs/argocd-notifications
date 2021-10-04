package api

import (
	"fmt"

	"github.com/argoproj/notifications-engine/pkg/services"
	"github.com/argoproj/notifications-engine/pkg/templates"
	"github.com/argoproj/notifications-engine/pkg/triggers"
)

const (
	serviceTypeVarName = "serviceType"
	recipientVarName   = "recipient"
)

//go:generate mockgen -destination=../mocks/api.go -package=mocks github.com/argoproj/notifications-engine/pkg/api API

type GetVars func(obj map[string]interface{}, dest services.Destination) map[string]interface{}

// API provides high level interface to send notifications and manage notification services
type API interface {
	Send(obj map[string]interface{}, templates []string, dest services.Destination) error
	RunTrigger(triggerName string, vars map[string]interface{}) ([]triggers.ConditionResult, error)
	AddNotificationService(name string, service services.NotificationService)
	GetNotificationServices() map[string]services.NotificationService
	GetConfig() Config
}

type api struct {
	notificationServices map[string]services.NotificationService
	templatesService     templates.Service
	triggersService      triggers.Service
	getVars              GetVars
	config               Config
}

func (n *api) GetConfig() Config {
	return n.config
}

// AddService adds new service with the specified name
func (n *api) AddNotificationService(name string, service services.NotificationService) {
	n.notificationServices[name] = service
}

// GetServices returns map of registered services
func (n *api) GetNotificationServices() map[string]services.NotificationService {
	return n.notificationServices
}

// Send sends notification using specified service and template to the specified destination
func (n *api) Send(obj map[string]interface{}, templates []string, dest services.Destination) error {
	notificationService, ok := n.notificationServices[dest.Service]
	if !ok {
		return fmt.Errorf("notification service '%s' is not supported", dest.Service)
	}

	vars := n.getVars(obj, dest)

	in := make(map[string]interface{})
	for k := range vars {
		in[k] = vars[k]
	}
	in[serviceTypeVarName] = dest.Service
	in[recipientVarName] = dest.Recipient
	notification, err := n.templatesService.FormatNotification(in, templates...)
	if err != nil {
		return err
	}

	return notificationService.Send(*notification, dest)
}

func (n *api) RunTrigger(triggerName string, obj map[string]interface{}) ([]triggers.ConditionResult, error) {
	vars := n.getVars(obj, services.Destination{})
	return n.triggersService.Run(triggerName, vars)
}

// NewAPI creates new api instance using provided config
func NewAPI(cfg Config, getVars GetVars) (*api, error) {
	notificationServices := map[string]services.NotificationService{}
	for k, v := range cfg.Services {
		svc, err := v()
		if err != nil {
			return nil, err
		}
		notificationServices[k] = svc
	}
	triggersService, err := triggers.NewService(cfg.Triggers)
	if err != nil {
		return nil, err
	}
	templatesService, err := templates.NewService(cfg.Templates)
	if err != nil {
		return nil, err
	}

	return &api{notificationServices, templatesService, triggersService, getVars, cfg}, nil
}
