package triggers

import (
	"testing"
	"time"

	"github.com/argoproj-labs/argocd-notifications/notifiers"

	testingutil "github.com/argoproj-labs/argocd-notifications/testing"

	"github.com/stretchr/testify/assert"
)

func TestGetTriggers_FailsIfReferencesNonExistingTemplate(t *testing.T) {
	_, err := GetTriggers([]NotificationTemplate{}, []NotificationTrigger{{
		Name:      "test",
		Template:  "bad",
		Condition: "true",
	}}, nil)
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
	}}, nil)
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
	}}, nil)
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
			Slack: &notifiers.SlackNotification{
				Attachments: "Application {{.app.metadata.name}} Application details: {{.context.argocdUrl}}/applications/{{.app.metadata.name}}.",
				Blocks:      "Application {{.app.metadata.name}} Application details: {{.context.argocdUrl}}/applications/{{.app.metadata.name}}.",
			},
		},
	}}, []NotificationTrigger{{
		Name:      "trigger",
		Template:  "template",
		Condition: "app.metadata.name == 'foo'",
	}}, nil)
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

func TestSpawnExprEnvs(t *testing.T) {
	opts := map[string]interface{}{"app": "dummy"}
	envs, ok := spawnExprEnvs(testingutil.NewApp("test"), opts, nil).(map[string]interface{})
	assert.True(t, ok)

	_, hasApp := envs["app"]
	assert.True(t, hasApp)
}

func TestGetTriggers_UsingExprVm(t *testing.T) {
	triggers, err := GetTriggers([]NotificationTemplate{{
		Name: "template",
		Notification: notifiers.Notification{
			Title: "the title: {{.app.metadata.name}}",
			Body:  "the body: {{.app.metadata.name}}",
		},
	}}, []NotificationTrigger{{
		Name:      "trigger",
		Template:  "template",
		Condition: "app.metadata.name == 'foo' && app.status.operationState.phase in ['Running'] && time.Now().Sub(time.Parse(app.status.operationState.startedAt)).Minutes() >= 5",
	}}, nil)
	assert.NoError(t, err)

	trigger, ok := triggers["trigger"]
	assert.True(t, ok)

	before2Minute := time.Now().Add(-2 * time.Minute)
	ok, err = trigger.Triggered(testingutil.NewApp("bar",
		testingutil.WithSyncOperationPhase("Running"),
		testingutil.WithSyncOperationStartAt(before2Minute)))
	assert.NoError(t, err)
	assert.False(t, ok)

	before5Minute := time.Now().Add(-5 * time.Minute)
	ok, err = trigger.Triggered(testingutil.NewApp("foo",
		testingutil.WithSyncOperationPhase("Running"),
		testingutil.WithSyncOperationStartAt(before5Minute)))
	assert.NoError(t, err)
	assert.True(t, ok)
}

func TestTrigger_FormatWebhookNotification(t *testing.T) {
	templates, err := parseTemplates([]NotificationTemplate{{
		Name: "myTemplate",
		Notification: notifiers.Notification{
			Webhook: map[string]notifiers.WebhookNotification{
				"test": {
					Method: "get",
					Body:   "hello {{.app.metadata.name}}",
				},
			},
		},
	}})
	assert.NoError(t, err)

	testTemplate, ok := templates["myTemplate"]
	if !assert.True(t, ok) {
		return
	}

	nt, err := testTemplate.formatNotification(testingutil.NewApp("world"), map[string]string{}, nil)
	if !assert.NoError(t, err) {
		return
	}

	hook, ok := nt.Webhook["test"]
	if !assert.True(t, ok) {
		return
	}
	assert.Equal(t, hook.Body, "hello world")
}
