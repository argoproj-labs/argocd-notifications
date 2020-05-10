package builtin

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"

	. "github.com/argoproj-labs/argocd-notifications/testing"
	"github.com/argoproj-labs/argocd-notifications/triggers"
)

func triggerWithTemplate(name string) (triggers.Trigger, error) {
	triggersByName, err := triggers.GetTriggers(Templates, []triggers.NotificationTrigger{{
		Name:      "test",
		Template:  name,
		Condition: "true",
	}}, nil)
	if err != nil {
		return nil, err
	}
	return triggersByName["test"], nil
}

func TestFormatNotification_SlackAttachment(t *testing.T) {
	for i := range Templates {
		trigger, err := triggerWithTemplate(Templates[i].Name)
		if !assert.NoError(t, err) {
			return
		}

		quotedString := `"wrong"`
		n, err := trigger.FormatNotification(
			NewApp(quotedString, WithConditions(quotedString, quotedString)), map[string]string{
				"argocdUrl":        quotedString,
				"notificationType": "slack",
			})

		if !assert.NoError(t, err) {
			return
		}

		var attachments []map[string]interface{}
		err = json.Unmarshal([]byte(n.Slack.Attachments), &attachments)
		assert.NoError(t, err)
	}
}
