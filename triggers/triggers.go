package triggers

//go:generate mockgen -destination=./mocks/triggers.go -package=mocks github.com/argoproj-labs/argocd-notifications/triggers Trigger

import (
	"bytes"
	"fmt"
	htmptemplate "html/template"

	"github.com/antonmedv/expr"
	"github.com/antonmedv/expr/vm"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type NotificationTrigger struct {
	Name        string `json:"name,omitempty"`
	Condition   string `json:"condition,omitempty"`
	Description string `json:"description,omitempty"`
	Template    string `json:"template,omitempty"`
	Enabled     *bool  `json:"enabled,omitempty"`
}

type NotificationTemplate struct {
	Name  string `json:"name,omitempty"`
	Title string `json:"title,omitempty"`
	Body  string `json:"body,omitempty"`
}

type Trigger interface {
	Triggered(app *unstructured.Unstructured) (bool, error)
	FormatNotification(app *unstructured.Unstructured, context map[string]string) (string, string, error)
}

type template struct {
	title *htmptemplate.Template
	body  *htmptemplate.Template
}

type trigger struct {
	condition *vm.Program
	template  template
}

func GetTriggers(templatesCfg []NotificationTemplate, triggersCfg []NotificationTrigger) (map[string]Trigger, error) {
	templates, err := parseTemplates(templatesCfg)
	if err != nil {
		return nil, err
	}
	return parseTriggers(triggersCfg, templates)
}

func (t *trigger) Triggered(app *unstructured.Unstructured) (bool, error) {
	if res, err := expr.Run(t.condition, map[string]interface{}{"app": app.Object}); err != nil {
		return false, err
	} else if boolRes, ok := res.(bool); ok {
		return boolRes, nil
	}
	return false, nil
}

func (t *trigger) FormatNotification(app *unstructured.Unstructured, context map[string]string) (string, string, error) {
	vars := map[string]interface{}{
		"app":     app.Object,
		"context": context,
	}
	var title bytes.Buffer
	err := t.template.title.Execute(&title, vars)
	if err != nil {
		return "", "", err
	}
	var body bytes.Buffer
	err = t.template.body.Execute(&body, vars)
	if err != nil {
		return "", "", err
	}
	return title.String(), body.String(), nil
}

func parseTemplates(templates []NotificationTemplate) (map[string]template, error) {
	res := make(map[string]template)
	for _, nt := range templates {
		title, err := htmptemplate.New(nt.Name).Parse(nt.Title)
		if err != nil {
			return nil, err
		}
		body, err := htmptemplate.New(nt.Name).Parse(nt.Body)
		if err != nil {
			return nil, err
		}
		res[nt.Name] = template{title: title, body: body}
	}
	return res, nil
}

func parseTriggers(triggers []NotificationTrigger, templates map[string]template) (map[string]Trigger, error) {
	res := make(map[string]Trigger)
	for _, t := range triggers {
		if t.Enabled != nil && *t.Enabled == false {
			continue
		}
		condition, err := expr.Compile(t.Condition)
		if err != nil {
			return nil, err
		}
		template, ok := templates[t.Template]
		if !ok {
			return nil, fmt.Errorf("trigger %s references unknown template %s", t.Name, t.Template)
		}
		res[t.Name] = &trigger{condition: condition, template: template}
	}
	return res, nil
}
