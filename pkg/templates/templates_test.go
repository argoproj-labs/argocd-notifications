package templates

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/argoproj-labs/argocd-notifications/pkg/services"
)

func TestFormat_BodyAndTitle(t *testing.T) {
	templ, err := NewTemplate(NotificationTemplate{Notification: services.Notification{
		Title: "{{.foo}}",
		Body:  "{{.bar}}",
	}})

	if !assert.NoError(t, err) {
		return
	}

	notification, err := templ.FormatNotification(map[string]interface{}{
		"foo": "hello",
		"bar": "world",
	})

	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t, "hello", notification.Title)
	assert.Equal(t, "world", notification.Body)
}

func TestFormat_Slack(t *testing.T) {
	templ, err := NewTemplate(NotificationTemplate{Notification: services.Notification{
		Slack: &services.SlackNotification{
			Attachments: "{{.foo}}",
			Blocks:      "{{.bar}}",
		},
	}})

	if !assert.NoError(t, err) {
		return
	}

	notification, err := templ.FormatNotification(map[string]interface{}{
		"foo": "hello",
		"bar": "world",
	})

	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t, "hello", notification.Slack.Attachments)
	assert.Equal(t, "world", notification.Slack.Blocks)
}

func TestFormat_Webhook(t *testing.T) {
	templ, err := NewTemplate(NotificationTemplate{Notification: services.Notification{
		Webhook: map[string]services.WebhookNotification{
			"github": {
				Method: "POST",
				Body:   "{{.foo}}",
				Path:   "{{.bar}}",
			},
		},
	}})

	if !assert.NoError(t, err) {
		return
	}

	notification, err := templ.FormatNotification(map[string]interface{}{
		"foo": "hello",
		"bar": "world",
	})

	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t, notification.Webhook["github"].Method, "POST")
	assert.Equal(t, notification.Webhook["github"].Body, "hello")
	assert.Equal(t, notification.Webhook["github"].Path, "world")
}
