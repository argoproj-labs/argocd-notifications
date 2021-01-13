package templates

import (
	"fmt"

	"github.com/Masterminds/sprig"

	"github.com/argoproj-labs/argocd-notifications/pkg/services"
)

type Service interface {
	FormatNotification(vars map[string]interface{}, templates ...string) (*services.Notification, error)
}

type service struct {
	templaters map[string]services.Templater
}

func NewService(templates map[string]services.Notification) (*service, error) {
	f := sprig.TxtFuncMap()
	delete(f, "env")
	delete(f, "expandenv")

	svc := &service{templaters: map[string]services.Templater{}}
	for name, cfg := range templates {
		templater, err := cfg.GetTemplater(name, f)
		if err != nil {
			return nil, err
		}
		svc.templaters[name] = templater
	}
	return svc, nil
}

func (s *service) FormatNotification(vars map[string]interface{}, templates ...string) (*services.Notification, error) {
	var notification services.Notification
	for _, templateName := range templates {
		templater, ok := s.templaters[templateName]
		if !ok {
			return nil, fmt.Errorf("template '%s' is not supported", templateName)
		}

		if err := templater(&notification, vars); err != nil {
			return nil, err
		}
	}
	return &notification, nil
}
