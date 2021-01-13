package templates

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/argoproj-labs/argocd-notifications/pkg/services"
)

func TestFormat_Message(t *testing.T) {
	svc, err := NewService(map[string]services.Notification{
		"test": {
			Message: "{{.foo}}",
		},
	})

	if !assert.NoError(t, err) {
		return
	}

	notification, err := svc.FormatNotification(map[string]interface{}{
		"foo": "hello",
	}, "test")

	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t, "hello", notification.Message)
}
