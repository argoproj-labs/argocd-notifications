package templates

import (
	"bytes"
	texttemplate "text/template"

	"github.com/Masterminds/sprig"

	"github.com/argoproj-labs/argocd-notifications/pkg/services"
)

type NotificationTemplate struct {
	services.Notification
	Name string `json:"name,omitempty"`
}

type Template interface {
	GetName() string
	FormatNotification(vars map[string]interface{}) (*services.Notification, error)
}

type webhookTemplate struct {
	body   *texttemplate.Template
	path   *texttemplate.Template
	method string
}

type template struct {
	name             string
	title            *texttemplate.Template
	body             *texttemplate.Template
	slackAttachments *texttemplate.Template
	slackBlocks      *texttemplate.Template
	webhooks         map[string]webhookTemplate
}

func (tmpl template) GetName() string {
	return tmpl.name
}

func (tmpl template) FormatNotification(vars map[string]interface{}) (*services.Notification, error) {

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
	notification := &services.Notification{Title: title.String(), Body: body.String()}
	if tmpl.slackAttachments != nil || tmpl.slackBlocks != nil {
		notification.Slack = &services.SlackNotification{}
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
	notification.Webhook = map[string]services.WebhookNotification{}
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
		notification.Webhook[k] = services.WebhookNotification{
			Method: v.method,
			Body:   body.String(),
			Path:   path.String(),
		}
	}
	return notification, nil
}

func NewTemplate(nt NotificationTemplate) (*template, error) {
	f := sprig.TxtFuncMap()
	delete(f, "env")
	delete(f, "expandenv")

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
	return &t, nil
}
