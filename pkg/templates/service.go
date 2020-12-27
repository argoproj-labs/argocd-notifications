package templates

import (
	"fmt"

	"github.com/argoproj-labs/argocd-notifications/pkg/services"
)

type Service interface {
	FormatNotification(vars map[string]interface{}, templates ...string) (*services.Notification, error)
}

type service struct {
	compiledTemplates map[string]*compiledTemplate
}

func NewService(templates map[string]services.Notification) (*service, error) {
	svc := &service{compiledTemplates: map[string]*compiledTemplate{}}
	for name, cfg := range templates {
		compiled, err := compileTemplate(name, cfg)
		if err != nil {
			return nil, err
		}
		svc.compiledTemplates[name] = compiled
	}
	return svc, nil
}

func (s *service) FormatNotification(vars map[string]interface{}, templates ...string) (*services.Notification, error) {
	var notification services.Notification
	for _, templateName := range templates {
		template, ok := s.compiledTemplates[templateName]
		if !ok {
			return nil, fmt.Errorf("template '%s' is not supported", templateName)
		}

		if err := template.formatNotification(vars, &notification); err != nil {
			return nil, err
		}
	}
	return &notification, nil
}
