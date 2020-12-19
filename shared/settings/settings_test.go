package settings

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/labels"

	v1 "k8s.io/api/core/v1"
)

var (
	emptySecret = &v1.Secret{Data: map[string][]byte{}}
)

func TestNewConfig_Subscriptions(t *testing.T) {
	cfg, err := NewConfig(&v1.ConfigMap{
		Data: map[string]string{
			"subscriptions": `
- selector: test=true
  triggers:
  - my-trigger2`,
		},
	}, emptySecret, nil)

	if !assert.NoError(t, err) {
		return
	}

	label, err := labels.Parse("test=true")
	if !assert.NoError(t, err) {
		return
	}
	assert.Equal(t, DefaultSubscriptions([]Subscription{
		{Triggers: []string{"my-trigger2"}, Selector: label},
	}), cfg.Subscriptions)
}

func TestNewSettings_Context(t *testing.T) {
	cfg, err := NewConfig(&v1.ConfigMap{
		Data: map[string]string{
			"context": `{hello: world}`,
		},
	}, emptySecret, nil)

	if !assert.NoError(t, err) {
		return
	}
	assert.Equal(t, map[string]string{"argocdUrl": "https://localhost:4000", "hello": "world"}, cfg.Context)
}

func TestNewSettings_DefaultTriggers(t *testing.T) {
	cfg, err := NewConfig(&v1.ConfigMap{
		Data: map[string]string{
			"defaultTriggers": `[trigger1, trigger2]`,
		},
	}, emptySecret, nil)

	if !assert.NoError(t, err) {
		return
	}
	assert.Equal(t, []string{"trigger1", "trigger2"}, cfg.DefaultTriggers)
}
