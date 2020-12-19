package settings

import (
	"testing"

	"k8s.io/apimachinery/pkg/labels"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	"k8s.io/utils/pointer"

	"github.com/argoproj-labs/argocd-notifications/pkg"
	"github.com/argoproj-labs/argocd-notifications/pkg/services"
	"github.com/argoproj-labs/argocd-notifications/pkg/templates"
	"github.com/argoproj-labs/argocd-notifications/triggers"
)

func TestMergeLegacyConfig_DefaultTriggers(t *testing.T) {
	cfg := Config{
		Config: pkg.Config{
			Services: map[string]pkg.ServiceFactory{},
		},
		TriggersSettings: []triggers.NotificationTrigger{
			{Name: "my-trigger1", Condition: "true"},
			{Name: "my-trigger2", Condition: "false"},
		},
		Context: map[string]string{},
	}
	configYAML := `
config.yaml:
triggers:
- name: my-trigger1
  enabled: true
`
	err := mergeLegacyConfig(&cfg,
		&v1.ConfigMap{Data: map[string]string{"config.yaml": configYAML}},
		&v1.Secret{Data: map[string][]byte{}},
	)
	assert.NoError(t, err)
	assert.Equal(t, []string{"my-trigger1"}, cfg.DefaultTriggers)
}

func TestMergeLegacyConfig(t *testing.T) {
	cfg := Config{
		Config: pkg.Config{
			Templates: []templates.NotificationTemplate{{Name: "my-template1", Notification: services.Notification{Body: "foo"}}},
			Services:  map[string]pkg.ServiceFactory{},
		},
		TriggersSettings: []triggers.NotificationTrigger{{Name: "my-trigger1", Condition: "true", Enabled: pointer.BoolPtr(false)}},
		Context:          map[string]string{"some": "value"},
		Subscriptions:    []Subscription{{Triggers: []string{"my-trigger1"}}},
	}

	configYAML := `
triggers:
- name: my-trigger1
  enabled: true
- name: my-trigger2
  condition: false
  enabled: true
templates:
- name: my-template1
  body: bar
- name: my-template2
  body: foo
subscriptions:
- triggers:
  - my-trigger2
  selector: test=true
`
	notifiersYAML := `
slack:
  token: my-token
`
	err := mergeLegacyConfig(&cfg,
		&v1.ConfigMap{Data: map[string]string{"config.yaml": configYAML}},
		&v1.Secret{Data: map[string][]byte{"notifiers.yaml": []byte(notifiersYAML)}},
	)

	assert.NoError(t, err)
	assert.Equal(t, []templates.NotificationTemplate{
		{Name: "my-template1", Notification: services.Notification{Body: "bar"}},
		{Name: "my-template2", Notification: services.Notification{Body: "foo"}},
	}, cfg.Templates)
	assert.Equal(t, []triggers.NotificationTrigger{
		{Name: "my-trigger1", Condition: "true", Enabled: pointer.BoolPtr(true)},
		{Name: "my-trigger2", Condition: "false", Enabled: pointer.BoolPtr(true)},
	}, cfg.TriggersSettings)
	label, err := labels.Parse("test=true")
	if !assert.NoError(t, err) {
		return
	}
	assert.Equal(t, DefaultSubscriptions([]Subscription{
		{Triggers: []string{"my-trigger2"}, Selector: label},
	}), cfg.Subscriptions)
	assert.NotNil(t, cfg.Services["slack"])
}
