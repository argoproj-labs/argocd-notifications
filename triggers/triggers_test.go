package triggers

import (
	"testing"

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
		Name:  "template",
		Title: "the title: {{.app.metadata.name}}",
		Body:  "the body: {{.app.metadata.name}}",
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

	title, body, err := trigger.FormatNotification(testingutil.NewApp("test"), map[string]string{})
	assert.NoError(t, err)
	assert.Equal(t, "the title: test", title)
	assert.Equal(t, "the body: test", body)
}
