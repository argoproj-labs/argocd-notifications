package main

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"

	"github.com/argoproj-labs/argocd-notifications/notifiers"
	"github.com/argoproj-labs/argocd-notifications/shared/settings"
	"github.com/argoproj-labs/argocd-notifications/triggers"
)

func TestWatchConfig(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	configMap := &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      settings.ConfigMapName,
			Namespace: "default",
		},
		Data: map[string]string{
			"config.yaml": `
triggers:
  - name: on-sync-status-unknown
    template: app-sync-status
    enabled: true
templates:
  - name: app-sync-status
    title: updated
    body: updated"`,
		},
	}
	secret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      settings.SecretName,
			Namespace: "default",
		},
		Data: map[string][]byte{
			"notifiers.yaml": []byte(`slack: {token: my-token}`),
		},
	}

	triggersMap := make(map[string]triggers.Trigger)
	notifiersMap := make(map[string]notifiers.Notifier)
	clientset := fake.NewSimpleClientset(configMap, secret)
	watchConfig(ctx, clientset, "default", func(t map[string]triggers.Trigger, n map[string]notifiers.Notifier, ctx map[string]string) error {
		triggersMap = t
		notifiersMap = n
		return nil
	})

	assert.Len(t, triggersMap, 1)

	_, ok := triggersMap["on-sync-status-unknown"]
	assert.True(t, ok)

	_, ok = notifiersMap["slack"]
	assert.True(t, ok)
}
