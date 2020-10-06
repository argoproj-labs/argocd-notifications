package main

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"

	"github.com/argoproj-labs/argocd-notifications/notifiers"
	"github.com/argoproj-labs/argocd-notifications/shared/argocd/mocks"
	"github.com/argoproj-labs/argocd-notifications/shared/settings"
	"github.com/argoproj-labs/argocd-notifications/triggers"
)

func TestWatchConfig(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	builtin := &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      settings.ConfigMapBuildInName,
			Namespace: "default",
		},
		Data: map[string]string{
			"config.yaml": `
triggers: []
templates: []
`,
		},
	}
	configMap := &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      settings.ConfigMapName,
			Namespace: "default",
		},
		Data: map[string]string{
			"config.yaml": `
triggers:
  - name: on-sync-status-unknown
    condition: "app.status.sync.status == 'Unknown'"
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
	argocdService := mocks.NewMockService(ctrl)
	clientset := fake.NewSimpleClientset(builtin, configMap, secret)
	watchConfig(ctx, argocdService, clientset, "default", func(t map[string]triggers.Trigger, n map[string]notifiers.Notifier, cfg *settings.Config) error {
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
