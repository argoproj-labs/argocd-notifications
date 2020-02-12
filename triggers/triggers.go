package triggers

//go:generate mockgen -destination=./mocks/triggers.go -package=mocks github.com/argoproj-labs/argocd-notifications/triggers Trigger

import (
	"bytes"
	"fmt"
	htmptemplate "html/template"

	"github.com/antonmedv/expr"
	"github.com/antonmedv/expr/vm"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/argoproj-labs/argocd-notifications/notifiers"
	"github.com/Masterminds/sprig"
	exprHelpers "github.com/argoproj-labs/argocd-notifications/triggers/expr"
)

type NotificationTrigger struct {
	Name        string `json:"name,omitempty"`
	Condition   string `json:"condition,omitempty"`
	Description string `json:"description,omitempty"`
	Template    string `json:"template,omitempty"`
	Enabled     *bool  `json:"enabled,omitempty"`
}

type NotificationTemplate struct {
	notifiers.Notification
	Name string `json:"name,omitempty"`
}

type Trigger interface {
	Triggered(app *unstructured.Unstructured) (bool, error)
	FormatNotification(app *unstructured.Unstructured, context map[string]string) (*notifiers.Notification, error)
}

type template struct {
	title            *htmptemplate.Template
	body             *htmptemplate.Template
	slackAttachments *htmptemplate.Template
	slackBlocks      *htmptemplate.Template
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

func spawnExprEnvs(opts map[string]interface{}) interface{} {
	envs := exprHelpers.Spawn()
	for name, env := range opts {
		envs[name] = env
	}

	return envs
}

func (t *trigger) Triggered(app *unstructured.Unstructured) (bool, error) {
	envs := map[string]interface{}{"app": app.Object}
	if res, err := expr.Run(t.condition, spawnExprEnvs(envs)); err != nil {
		return false, err
	} else if boolRes, ok := res.(bool); ok {
		return boolRes, nil
	}
	return false, nil
}

func (t *trigger) FormatNotification(app *unstructured.Unstructured, context map[string]string) (*notifiers.Notification, error) {
	vars := map[string]interface{}{
		"app":     app.Object,
		"context": context,
	}
	var title bytes.Buffer
	err := t.template.title.Execute(&title, vars)
	if err != nil {
		return nil, err
	}
	var body bytes.Buffer
	err = t.template.body.Execute(&body, vars)
	if err != nil {
		return nil, err
	}
	notification := &notifiers.Notification{Title: title.String(), Body: body.String()}
	if t.template.slackAttachments != nil || t.template.slackBlocks != nil {
		notification.Slack = &notifiers.SlackSpecific{}
	}
	if t.template.slackAttachments != nil {
		var slackAttachments bytes.Buffer
		err = t.template.slackAttachments.Execute(&slackAttachments, vars)
		if err != nil {
			return nil, err
		}
		notification.Slack.Attachments = slackAttachments.String()
	}
	if t.template.slackBlocks != nil {
		var slackBlocks bytes.Buffer
		err = t.template.slackBlocks.Execute(&slackBlocks, vars)
		if err != nil {
			return nil, err
		}
		notification.Slack.Blocks = slackBlocks.String()
	}
	return notification, nil
}

func parseTemplates(templates []NotificationTemplate) (map[string]template, error) {
	res := make(map[string]template)
	for _, nt := range templates {
		title, err := htmptemplate.New(nt.Name).Funcs(sprig.FuncMap()).Parse(nt.Title)
		if err != nil {
			return nil, err
		}
		body, err := htmptemplate.New(nt.Name).Funcs(sprig.FuncMap()).Parse(nt.Body)
		if err != nil {
			return nil, err
		}
		t := template{title: title, body: body}
		if nt.Slack != nil {
			slackAttachments, err := htmptemplate.New(nt.Name).Funcs(sprig.FuncMap()).Parse(nt.Slack.Attachments)
			if err != nil {
				return nil, err
			}
			t.slackAttachments = slackAttachments
			slackBlocks, err := htmptemplate.New(nt.Name).Funcs(sprig.FuncMap()).Parse(nt.Slack.Blocks)
			if err != nil {
				return nil, err
			}
			t.slackBlocks = slackBlocks
		}
		res[nt.Name] = t
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
