package recipients

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/labels"

	"github.com/argoproj-labs/argocd-notifications/pkg"
	"github.com/argoproj-labs/argocd-notifications/pkg/services"
	"github.com/argoproj-labs/argocd-notifications/shared/settings"
)

func TestGetGlobalRecipients_GlobalTriggers(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	recipients, err := GetGlobalRecipients(map[string]string{}, settings.DefaultSubscriptions{{
		Selector:   labels.Everything(),
		Recipients: []string{"email:alex@gmail.com"},
	}}, getTriggers(ctrl), []string{"trigger1", "trigger2"})

	assert.NoError(t, err)

	assert.ElementsMatch(t, recipients.GetNotificationSubscriptions(), []pkg.NotificationSubscription{{
		When: "trigger1",
		Send: "trigger1-template",
		To: []services.Destination{{
			Service: "email", Recipient: "alex@gmail.com",
		}},
	}, {
		When: "trigger2",
		Send: "trigger2-template",
		To: []services.Destination{{
			Service: "email", Recipient: "alex@gmail.com",
		}},
	}})
}
