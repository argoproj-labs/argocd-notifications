package triggers

import (
	"testing"

	"github.com/argoproj-labs/argocd-notifications/notifiers"

	testingutil "github.com/argoproj-labs/argocd-notifications/testing"

	"github.com/stretchr/testify/assert"
)

func TestGetTriggers_FailsIfReferencesNonExistingTemplate(t *testing.T) {
	_, err := GetTriggers([]NotificationTemplate{}, []NotificationTrigger{{
		Name:      "test",
		Template:  "bad",
		Condition: "true",
	}})
	assert.EqualError(t, err, "trigger test references unknown template bad")
}

func TestGetTriggers(t *testing.T) {
	triggers, err := GetTriggers([]NotificationTemplate{{
		Name: "template",
		Notification: notifiers.Notification{
			Title: "the title: {{.app.metadata.name}}",
			Body:  "the body: {{.app.metadata.name}}",
		},
	}}, []NotificationTrigger{{
		Name:      "trigger",
		Template:  "template",
		Condition: "app.metadata.name == 'foo'",
	}})
	assert.NoError(t, err)

	trigger, ok := triggers["trigger"]
	assert.True(t, ok)

	ok, err = trigger.Triggered(testingutil.NewApp("foo"))
	assert.NoError(t, err)
	assert.True(t, ok)

	ok, err = trigger.Triggered(testingutil.NewApp("bar"))
	assert.NoError(t, err)
	assert.False(t, ok)

	notification, err := trigger.FormatNotification(testingutil.NewApp("test"), map[string]string{})
	assert.NoError(t, err)
	assert.Equal(t, "the title: test", notification.Title)
	assert.Equal(t, "the body: test", notification.Body)
}

func TestGetTriggers_UsingContext(t *testing.T) {
	triggers, err := GetTriggers([]NotificationTemplate{{
		Name: "template",
		Notification: notifiers.Notification{
			Title: "the title: {{.app.metadata.name}}",
			Body:  "the body: {{.app.metadata.name}} argocd url: {{.context.argocdUrl}}",
		},
	}}, []NotificationTrigger{{
		Name:      "trigger",
		Template:  "template",
		Condition: "app.metadata.name == 'foo'",
	}})
	assert.NoError(t, err)

	trigger, ok := triggers["trigger"]
	assert.True(t, ok)

	ok, err = trigger.Triggered(testingutil.NewApp("foo"))
	assert.NoError(t, err)
	assert.True(t, ok)

	ok, err = trigger.Triggered(testingutil.NewApp("bar"))
	assert.NoError(t, err)
	assert.False(t, ok)

	notification, err := trigger.FormatNotification(testingutil.NewApp("test"), map[string]string{"argocdUrl": "testUrl"})
	assert.NoError(t, err)
	assert.Equal(t, "the title: test", notification.Title)
	assert.Equal(t, "the body: test argocd url: testUrl", notification.Body)
}

func TestGetTriggers_UsingSlack(t *testing.T) {
	triggers, err := GetTriggers([]NotificationTemplate{{
		Name: "template",
		Notification: notifiers.Notification{
			Title: "the title: {{.app.metadata.name}}",
			Body:  "the body: {{.app.metadata.name}}",
			Slack: &notifiers.SlackSpecific{
				Attachments: "Application {{.app.metadata.name}} Application details: {{.context.argocdUrl}}/applications/{{.app.metadata.name}}.",
				Blocks:      "Application {{.app.metadata.name}} Application details: {{.context.argocdUrl}}/applications/{{.app.metadata.name}}.",
			},
		},
	}}, []NotificationTrigger{{
		Name:      "trigger",
		Template:  "template",
		Condition: "app.metadata.name == 'foo'",
	}})
	assert.NoError(t, err)

	trigger, ok := triggers["trigger"]
	assert.True(t, ok)

	ok, err = trigger.Triggered(testingutil.NewApp("foo"))
	assert.NoError(t, err)
	assert.True(t, ok)

	ok, err = trigger.Triggered(testingutil.NewApp("bar"))
	assert.NoError(t, err)
	assert.False(t, ok)

	notification, err := trigger.FormatNotification(testingutil.NewApp("test"), map[string]string{"argocdUrl": "testUrl"})
	assert.NoError(t, err)
	assert.Equal(t, "the title: test", notification.Title)
	assert.Equal(t, "the body: test", notification.Body)
	assert.Equal(t, "Application test Application details: testUrl/applications/test.", notification.Slack.Attachments)
	assert.Equal(t, "Application test Application details: testUrl/applications/test.", notification.Slack.Blocks)
}
