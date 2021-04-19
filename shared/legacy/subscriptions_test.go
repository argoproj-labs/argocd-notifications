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
	}, []string{}, nil)
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
	}, []string{"my-trigger"}, nil)
	assert.Equal(t, pkg.Subscriptions{
		"my-trigger": []services.Destination{{
			Recipient: "my-channel",
			Service:   "slack",
		}},
	}, res)
}

func TestGetSubscriptions_ServiceDefaultTriggers(t *testing.T) {
	res := GetSubscriptions(map[string]string{
		"recipients.argocd-notifications.argoproj.io": "slack:my-channel",
	}, []string{}, map[string][]string{
		"slack": {
			"trigger-a",
			"trigger-b",
		},
	})
	assert.Equal(t, pkg.Subscriptions{
		"trigger-a": []services.Destination{{
			Recipient: "my-channel",
			Service:   "slack",
		}},
		"trigger-b": []services.Destination{{
			Recipient: "my-channel",
			Service:   "slack",
		}},
	}, res)
}
