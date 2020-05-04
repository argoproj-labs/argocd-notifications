package tools

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/argoproj-labs/argocd-notifications/notifiers"
	"github.com/argoproj-labs/argocd-notifications/shared/settings"
	testingutil "github.com/argoproj-labs/argocd-notifications/testing"
	"github.com/argoproj-labs/argocd-notifications/triggers"
)

func TestTemplateNotifyConsole(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	ctx, closer, err := newTestContext(&stdout, &stderr, settings.Config{
		Triggers: []triggers.NotificationTrigger{{
			Name:      "my-trigger",
			Condition: "app.metadata.name == 'guestbook'",
			Template:  "my-template",
		}},
		Templates: []triggers.NotificationTemplate{{
			Name: "my-template",
			Notification: notifiers.Notification{
				Title: "hello {{.app.metadata.name}}",
			},
		}},
	}, testingutil.NewApp("guestbook"))
	if !assert.NoError(t, err) {
		return
	}
	defer closer()

	command := newTemplateNotifyCommand(ctx)
	err = command.RunE(command, []string{"my-template", "guestbook"})
	assert.NoError(t, err)
	assert.Empty(t, stderr.String())
	assert.Contains(t, stdout.String(), "hello guestbook")
}

func TestTemplateGet(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	ctx, closer, err := newTestContext(&stdout, &stderr, settings.Config{
		Templates: []triggers.NotificationTemplate{{
			Name: "my-template1",
		}, {
			Name: "my-template2",
		}},
	})
	if !assert.NoError(t, err) {
		return
	}
	defer closer()

	command := newTemplateGetCommand(ctx)
	err = command.RunE(command, nil)
	assert.NoError(t, err)
	assert.Empty(t, stderr.String())
	assert.Contains(t, stdout.String(), "my-template1")
	assert.Contains(t, stdout.String(), "my-template2")
}
