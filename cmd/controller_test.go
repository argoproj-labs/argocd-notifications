package main

import (
	"context"
	"testing"

	"github.com/argoproj-labs/argocd-notifications/shared/settings"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"

	"github.com/argoproj-labs/argocd-notifications/shared/argocd/mocks"
	"github.com/argoproj-labs/argocd-notifications/shared/k8s"
)

func TestWatchConfig(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	configMap := &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      k8s.ConfigMapName,
			Namespace: "default",
		},
		Data: map[string]string{
			"context": `
argocdUrl: https://myargocd.com
`,
		},
	}
	secret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      k8s.SecretName,
			Namespace: "default",
		},
		Data: map[string][]byte{},
	}

	argocdService := mocks.NewMockService(ctrl)
	clientset := fake.NewSimpleClientset(configMap, secret)
	var parsedCfg *settings.Config
	watchConfig(ctx, argocdService, clientset, "default", func(cfg settings.Config) error {
		parsedCfg = &cfg
		return nil
	})

	if !assert.NotNil(t, parsedCfg) {
		return
	}

	assert.Equal(t, "https://myargocd.com", parsedCfg.Context["argocdUrl"])
}
