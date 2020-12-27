package legacy

import (
	"testing"

	"github.com/argoproj-labs/argocd-notifications/pkg"
	"github.com/argoproj-labs/argocd-notifications/pkg/services"

	"github.com/stretchr/testify/assert"
)

func TestGetSubscriptions(t *testing.T) {
	res := GetSubscriptions(map[string]string{
		"my-trigger.recipients.argocd-notifications.argoproj.io": "slack:my-channel",
	})
	assert.Equal(t, pkg.Subscriptions{
		"my-trigger": []services.Destination{{
			Recipient: "my-channel",
			Service:   "slack",
		},
		}}, res)
}

func TestGetSubscriptions_DefaultTrigger(t *testing.T) {
	res := GetSubscriptions(map[string]string{
		"recipients.argocd-notifications.argoproj.io": "slack:my-channel",
	}, "my-trigger")
	assert.Equal(t, pkg.Subscriptions{
		"my-trigger": []services.Destination{{
			Recipient: "my-channel",
			Service:   "slack",
		}},
	}, res)
}
