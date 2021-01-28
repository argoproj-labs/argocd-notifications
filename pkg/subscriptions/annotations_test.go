package subscriptions

import (
	"testing"

	"github.com/argoproj-labs/argocd-notifications/pkg/services"

	"github.com/stretchr/testify/assert"

	"github.com/argoproj-labs/argocd-notifications/pkg"
)

func TestIterate(t *testing.T) {
	tests := []struct {
		annotations map[string]string
		trigger     string
		service     string
		recipients  []string
		key         string
	}{
		{
			annotations: map[string]string{
				"notifications.argoproj.io/subscribe.my-trigger.slack": "my-channel",
			},
			trigger:    "my-trigger",
			service:    "slack",
			recipients: []string{"my-channel"},
			key:        "notifications.argoproj.io/subscribe.my-trigger.slack",
		},
		{
			annotations: map[string]string{
				"notifications.argoproj.io/subscribe..slack": "my-channel",
			},
			trigger:    "",
			service:    "slack",
			recipients: []string{"my-channel"},
			key:        "notifications.argoproj.io/subscribe..slack",
		},
		{
			annotations: map[string]string{
				"notifications.argoproj.io/subscribe.slack": "my-channel",
			},
			trigger:    "",
			service:    "slack",
			recipients: []string{"my-channel"},
			key:        "notifications.argoproj.io/subscribe.slack",
		},
	}

	for _, tt := range tests {
		a := Annotations(tt.annotations)
		a.iterate(func(trigger, service string, recipients []string, key string) {
			assert.Equal(t, tt.trigger, trigger)
			assert.Equal(t, tt.service, service)
			assert.Equal(t, tt.recipients, recipients)
			assert.Equal(t, tt.key, key)
		})
	}
}

func TestGetAll(t *testing.T) {
	a := Annotations(map[string]string{
		"notifications.argoproj.io/subscribe.my-trigger.slack": "my-channel",
	})
	subscriptions := a.GetAll()
	assert.Equal(t, pkg.Subscriptions{
		"my-trigger": []services.Destination{{
			Service:   "slack",
			Recipient: "my-channel",
		}},
	}, subscriptions)
}

func TestSubscribe(t *testing.T) {
	a := Annotations(map[string]string{})
	a.Subscribe("my-trigger", "slack", "my-channel1")

	assert.Equal(t, a["notifications.argoproj.io/subscribe.my-trigger.slack"], "my-channel1")
}

func TestSubscribe_AddSecondRecipient(t *testing.T) {
	a := Annotations(map[string]string{
		"notifications.argoproj.io/subscribe.my-trigger.slack": "my-channel1",
	})
	a.Subscribe("my-trigger", "slack", "my-channel2")

	assert.Equal(t, a["notifications.argoproj.io/subscribe.my-trigger.slack"], "my-channel1;my-channel2")
}

func TestUnsubscribe(t *testing.T) {
	a := Annotations(map[string]string{
		"notifications.argoproj.io/subscribe.my-trigger.slack": "my-channel1;my-channel2",
	})
	a.Unsubscribe("my-trigger", "slack", "my-channel1")
	assert.Equal(t, a["notifications.argoproj.io/subscribe.my-trigger.slack"], "my-channel2")
	a.Unsubscribe("my-trigger", "slack", "my-channel2")
	_, ok := a["notifications.argoproj.io/subscribe.my-trigger.slack"]
	assert.False(t, ok)
}
