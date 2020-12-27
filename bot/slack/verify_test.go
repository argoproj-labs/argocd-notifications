package slack

import (
	"strconv"
	"testing"
	"time"

	"github.com/argoproj-labs/argocd-notifications/pkg"
	"github.com/argoproj-labs/argocd-notifications/pkg/services"
	"github.com/argoproj-labs/argocd-notifications/shared/settings"

	"github.com/stretchr/testify/assert"
)

func TestNewVerifier_IncorrectConfig(t *testing.T) {
	testCases := map[string]struct {
		Services map[string]services.NotificationService
		Error    string
	}{
		"NoSlack": {
			Services: map[string]services.NotificationService{},
			Error:    "slack is not configured",
		},
		"SlackWithoutSigningSecret": {
			Services: map[string]services.NotificationService{"slack": services.NewSlackService(services.SlackOptions{})},
			Error:    "slack signing secret is not configured",
		},
	}

	for k := range testCases {
		testCase := testCases[k]

		t.Run(k, func(t *testing.T) {

			api, err := pkg.NewAPI(pkg.Config{})
			if !assert.NoError(t, err) {
				return
			}
			for k, v := range testCase.Services {
				api.AddNotificationService(k, v)
			}
			verifier := NewVerifier(settings.Config{API: api})

			_, err = verifier(nil, nil)

			assert.Error(t, err)
			assert.Contains(t, err.Error(), testCase.Error)
		})
	}
}

func TestNewVerifier_IncorrectSignature(t *testing.T) {
	api, err := pkg.NewAPI(pkg.Config{})
	if !assert.NoError(t, err) {
		return
	}
	api.AddNotificationService("slack", services.NewSlackService(services.SlackOptions{SigningSecret: "hello world"}))
	verifier := NewVerifier(settings.Config{API: api})

	now := time.Now()
	data := "hello world"
	_, err = verifier([]byte(data), map[string][]string{
		"X-Slack-Request-Timestamp": {strconv.Itoa(int(now.UnixNano()))},
		"X-Slack-Signature":         {"v0=9e3753bb47fd3495894ab133c423ec93eff1ff30dd905ce39dda065e21ed9255"},
	})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Expected signing signature")
}
