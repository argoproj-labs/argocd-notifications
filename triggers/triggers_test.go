package triggers

import (
	"testing"

	testingutil "github.com/argoproj-labs/argocd-notifications/testing"

	"github.com/stretchr/testify/assert"
)

func TestTriggered(t *testing.T) {
	trigger, err := NewTrigger(NotificationTrigger{
		Name:      "trigger",
		Template:  "template",
		Condition: "app.metadata.name == 'foo'",
	}, nil)
	assert.NoError(t, err)

	ok, err := trigger.Triggered(testingutil.NewApp("foo"))
	assert.NoError(t, err)
	assert.True(t, ok)

	ok, err = trigger.Triggered(testingutil.NewApp("bar"))
	assert.NoError(t, err)
	assert.False(t, ok)
}
