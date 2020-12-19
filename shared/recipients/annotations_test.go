package recipients

import (
	"testing"

	"github.com/argoproj-labs/argocd-notifications/pkg"
	"github.com/argoproj-labs/argocd-notifications/pkg/services"
	"github.com/argoproj-labs/argocd-notifications/triggers"
	"github.com/argoproj-labs/argocd-notifications/triggers/mocks"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func getTriggers(ctrl *gomock.Controller) map[string]triggers.Trigger {
	trigger1 := mocks.NewMockTrigger(ctrl)
	trigger1.EXPECT().GetTemplate().Return("trigger1-template").AnyTimes()
	trigger2 := mocks.NewMockTrigger(ctrl)
	trigger2.EXPECT().GetTemplate().Return("trigger2-template").AnyTimes()
	return map[string]triggers.Trigger{"trigger1": trigger1, "trigger2": trigger2}
}

func TestGetRecipientsFromAnnotations_DefaultTriggers(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	recipients, err := GetRecipientsFromAnnotations(map[string]string{
		"recipients.argocd-notifications.argoproj.io": "email:alex@gmail.com, slack:my-channel",
	}, getTriggers(ctrl), []string{"trigger1", "trigger2"})

	if !assert.NoError(t, err) {
		return
	}

	assert.ElementsMatch(t, recipients.GetNotificationSubscriptions(), []pkg.NotificationSubscription{{
		When: "trigger1",
		Send: "trigger1-template",
		To: []services.Destination{{
			Service: "email", Recipient: "alex@gmail.com",
		}, {
			Service: "slack", Recipient: "my-channel",
		}},
	}, {
		When: "trigger2",
		Send: "trigger2-template",
		To: []services.Destination{{
			Service: "email", Recipient: "alex@gmail.com",
		}, {
			Service: "slack", Recipient: "my-channel",
		}},
	}})
}

func TestGetRecipientsFromAnnotations_SpecifiedTrigger(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	recipients, err := GetRecipientsFromAnnotations(map[string]string{
		"trigger1.recipients.argocd-notifications.argoproj.io": "email:alex@gmail.com, slack:my-channel",
	}, getTriggers(ctrl), []string{"trigger1", "trigger2"})

	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t, recipients.GetNotificationSubscriptions(), []pkg.NotificationSubscription{{
		When: "trigger1",
		Send: "trigger1-template",
		To: []services.Destination{{
			Service: "email", Recipient: "alex@gmail.com",
		}, {
			Service: "slack", Recipient: "my-channel",
		}},
	}})
}
