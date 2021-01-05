package templates

import (
	"bytes"
	texttemplate "text/template"

	"github.com/Masterminds/sprig"

	"github.com/argoproj-labs/argocd-notifications/pkg/services"
)

type compiledWebhookTemplate struct {
	body   *texttemplate.Template
	path   *texttemplate.Template
	method string
}

type compiledTemplate struct {
	title            *texttemplate.Template
	body             *texttemplate.Template
	slackAttachments *texttemplate.Template
	slackBlocks      *texttemplate.Template
	webhooks         map[string]compiledWebhookTemplate
}

func (tmpl compiledTemplate) formatNotification(vars map[string]interface{}, notification *services.Notification) error {

	var title bytes.Buffer

	err := tmpl.title.Execute(&title, vars)
	if err != nil {
		return err
	}
	if val := title.String(); val != "" {
		notification.Title = val
	}

	var body bytes.Buffer
	err = tmpl.body.Execute(&body, vars)
	if err != nil {
		return err
	}
	if val := body.String(); val != "" {
		notification.Body = val
	}

	if tmpl.slackAttachments != nil || tmpl.slackBlocks != nil {
		notification.Slack = &services.SlackNotification{}
	}
	if tmpl.slackAttachments != nil {
		var slackAttachments bytes.Buffer
		err = tmpl.slackAttachments.Execute(&slackAttachments, vars)
		if err != nil {
			return err
		}
		notification.Slack.Attachments = slackAttachments.String()
	}
	if tmpl.slackBlocks != nil {
		var slackBlocks bytes.Buffer
		err = tmpl.slackBlocks.Execute(&slackBlocks, vars)
		if err != nil {
			return err
		}
		notification.Slack.Blocks = slackBlocks.String()
	}
	for k, v := range tmpl.webhooks {
		if notification.Webhook == nil {
			notification.Webhook = map[string]services.WebhookNotification{}
		}
		var body bytes.Buffer
		err = tmpl.webhooks[k].body.Execute(&body, vars)
		if err != nil {
			return err
		}
		var path bytes.Buffer
		err = tmpl.webhooks[k].path.Execute(&path, vars)
		if err != nil {
			return err
		}
		notification.Webhook[k] = services.WebhookNotification{
			Method: v.method,
			Body:   body.String(),
			Path:   path.String(),
		}
	}
	return nil
}

func compileTemplate(name string, cfg services.Notification) (*compiledTemplate, error) {
	f := sprig.TxtFuncMap()
	delete(f, "env")
	delete(f, "expandenv")

	title, err := texttemplate.New(name).Funcs(f).Parse(cfg.Title)
	if err != nil {
		return nil, err
	}
	body, err := texttemplate.New(name).Funcs(f).Parse(cfg.Body)
	if err != nil {
		return nil, err
	}
	t := compiledTemplate{title: title, body: body}
	if cfg.Slack != nil {
		slackAttachments, err := texttemplate.New(name).Funcs(f).Parse(cfg.Slack.Attachments)
		if err != nil {
			return nil, err
		}
		t.slackAttachments = slackAttachments
		slackBlocks, err := texttemplate.New(name).Funcs(f).Parse(cfg.Slack.Blocks)
		if err != nil {
			return nil, err
		}
		t.slackBlocks = slackBlocks
	}

	t.webhooks = map[string]compiledWebhookTemplate{}
	for k, v := range cfg.Webhook {
		body, err := texttemplate.New(k).Funcs(f).Parse(v.Body)
		if err != nil {
			return nil, err
		}
		path, err := texttemplate.New(k).Funcs(f).Parse(v.Path)
		if err != nil {
			return nil, err
		}
		t.webhooks[k] = compiledWebhookTemplate{body: body, method: v.Method, path: path}
	}
	return &t, nil
}
