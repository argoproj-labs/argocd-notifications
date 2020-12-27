package subscriptions

import (
	"testing"

	"github.com/argoproj-labs/argocd-notifications/pkg/services"

	"github.com/stretchr/testify/assert"

	"github.com/argoproj-labs/argocd-notifications/pkg"
)

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
