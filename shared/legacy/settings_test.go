package legacy

import (
	"testing"

	"github.com/argoproj-labs/argocd-notifications/shared/settings"

	"k8s.io/apimachinery/pkg/labels"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"

	"github.com/argoproj-labs/argocd-notifications/pkg"
	"github.com/argoproj-labs/argocd-notifications/pkg/services"
	"github.com/argoproj-labs/argocd-notifications/pkg/triggers"
)

func TestMergeLegacyConfig_DefaultTriggers(t *testing.T) {
	cfg := settings.Config{
		Config: pkg.Config{
			Services: map[string]pkg.ServiceFactory{},
			Triggers: map[string][]triggers.Condition{
				"my-trigger1": {{
					When: "true",
					Send: []string{"my-template1"},
				}},
				"my-trigger2": {{
					When: "false",
					Send: []string{"my-template2"},
				}},
			},
		},
		Context: map[string]string{},
	}
	configYAML := `
config.yaml:
triggers:
- name: my-trigger1
  enabled: true
`
	err := ApplyLegacyConfig(&cfg,
		&v1.ConfigMap{Data: map[string]string{"config.yaml": configYAML}},
		&v1.Secret{Data: map[string][]byte{}},
	)
	assert.NoError(t, err)
	assert.Equal(t, []string{"my-trigger1"}, cfg.DefaultTriggers)
}

func TestMergeLegacyConfig(t *testing.T) {
	cfg := settings.Config{
		Config: pkg.Config{
			Templates: map[string]services.Notification{"my-template1": {Message: "foo"}},
			Triggers: map[string][]triggers.Condition{
				"my-trigger1": {{
					When: "true",
					Send: []string{"my-template1"},
				}},
			},
			Services: map[string]pkg.ServiceFactory{},
		},
		Context:       map[string]string{"some": "value"},
		Subscriptions: []settings.DefaultSubscription{{Triggers: []string{"my-trigger1"}}},
	}

	configYAML := `
triggers:
- name: my-trigger1
  enabled: true
- name: my-trigger2
  condition: false
  template: my-template2
  enabled: true
templates:
- name: my-template1
  body: bar
- name: my-template2
  body: foo
context:
  other: value2
subscriptions:
- triggers:
  - my-trigger2
  selector: test=true
`
	notifiersYAML := `
slack:
  token: my-token
`
	err := ApplyLegacyConfig(&cfg,
		&v1.ConfigMap{Data: map[string]string{"config.yaml": configYAML}},
		&v1.Secret{Data: map[string][]byte{"notifiers.yaml": []byte(notifiersYAML)}},
	)

	assert.NoError(t, err)
	assert.Equal(t, map[string]services.Notification{
		"my-template1": {Message: "bar"},
		"my-template2": {Message: "foo"},
	}, cfg.Templates)

	assert.Equal(t, []triggers.Condition{{
		When: "true",
		Send: []string{"my-template1"},
	}}, cfg.Triggers["my-trigger1"])
	assert.Equal(t, []triggers.Condition{{
		When: "false",
		Send: []string{"my-template2"},
	}}, cfg.Triggers["my-trigger2"])

	label, err := labels.Parse("test=true")
	if !assert.NoError(t, err) {
		return
	}
	assert.Equal(t, settings.DefaultSubscriptions([]settings.DefaultSubscription{
		{Triggers: []string{"my-trigger2"}, Selector: label},
	}), cfg.Subscriptions)
	assert.NotNil(t, cfg.Services["slack"])
}
