package triggers

//go:generate mockgen -destination=./mocks/triggers.go -package=mocks github.com/argoproj-labs/argocd-notifications/triggers Trigger

import (
	"bytes"
	"fmt"
	texttemplate "text/template"

	"github.com/antonmedv/expr"
	"github.com/antonmedv/expr/vm"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/Masterminds/sprig"
	"github.com/argoproj-labs/argocd-notifications/notifiers"
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

type webhookTemplate struct {
	body   *texttemplate.Template
	path   *texttemplate.Template
	method string
}

type template struct {
	title            *texttemplate.Template
	body             *texttemplate.Template
	slackAttachments *texttemplate.Template
	slackBlocks      *texttemplate.Template
	webhooks         map[string]webhookTemplate
}

func (tmpl template) formatNotification(app *unstructured.Unstructured, context map[string]string) (*notifiers.Notification, error) {
	vars := map[string]interface{}{
		"app":     app.Object,
		"context": context,
	}
	for k, v := range exprHelpers.Spawn() {
		vars[k] = v
	}
	var title bytes.Buffer

	err := tmpl.title.Execute(&title, vars)
	if err != nil {
		return nil, err
	}
	var body bytes.Buffer
	err = tmpl.body.Execute(&body, vars)
	if err != nil {
		return nil, err
	}
	notification := &notifiers.Notification{Title: title.String(), Body: body.String()}
	if tmpl.slackAttachments != nil || tmpl.slackBlocks != nil {
		notification.Slack = &notifiers.SlackNotification{}
	}
	if tmpl.slackAttachments != nil {
		var slackAttachments bytes.Buffer
		err = tmpl.slackAttachments.Execute(&slackAttachments, vars)
		if err != nil {
			return nil, err
		}
		notification.Slack.Attachments = slackAttachments.String()
	}
	if tmpl.slackBlocks != nil {
		var slackBlocks bytes.Buffer
		err = tmpl.slackBlocks.Execute(&slackBlocks, vars)
		if err != nil {
			return nil, err
		}
		notification.Slack.Blocks = slackBlocks.String()
	}
	notification.Webhook = map[string]notifiers.WebhookNotification{}
	for k, v := range tmpl.webhooks {
		var body bytes.Buffer
		err = tmpl.webhooks[k].body.Execute(&body, vars)
		if err != nil {
			return nil, err
		}
		var path bytes.Buffer
		err = tmpl.webhooks[k].path.Execute(&path, vars)
		if err != nil {
			return nil, err
		}
		notification.Webhook[k] = notifiers.WebhookNotification{
			Method: v.method,
			Body:   body.String(),
			Path:   path.String(),
		}
	}
	return notification, nil
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
	return t.template.formatNotification(app, context)
}

func parseTemplates(templates []NotificationTemplate) (map[string]template, error) {
	res := make(map[string]template)
	f := sprig.TxtFuncMap()
	delete(f, "env")
	delete(f, "expandenv")

	for _, nt := range templates {
		title, err := texttemplate.New(nt.Name).Funcs(f).Parse(nt.Title)
		if err != nil {
			return nil, err
		}
		body, err := texttemplate.New(nt.Name).Funcs(f).Parse(nt.Body)
		if err != nil {
			return nil, err
		}
		t := template{title: title, body: body}
		if nt.Slack != nil {
			slackAttachments, err := texttemplate.New(nt.Name).Funcs(f).Parse(nt.Slack.Attachments)
			if err != nil {
				return nil, err
			}
			t.slackAttachments = slackAttachments
			slackBlocks, err := texttemplate.New(nt.Name).Funcs(f).Parse(nt.Slack.Blocks)
			if err != nil {
				return nil, err
			}
			t.slackBlocks = slackBlocks
		}

		t.webhooks = map[string]webhookTemplate{}
		for k, v := range nt.Webhook {
			body, err := texttemplate.New(k).Funcs(f).Parse(v.Body)
			if err != nil {
				return nil, err
			}
			path, err := texttemplate.New(k).Funcs(f).Parse(v.Path)
			if err != nil {
				return nil, err
			}
			t.webhooks[k] = webhookTemplate{body: body, method: v.Method, path: path}
		}

		res[nt.Name] = t
	}
	return res, nil
}

func parseTriggers(triggers []NotificationTrigger, templates map[string]template) (map[string]Trigger, error) {
	res := make(map[string]Trigger)
	for _, t := range triggers {
		if t.Enabled != nil && !*t.Enabled {
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
