package settings

import (
	"context"
	"testing"

	"github.com/argoproj-labs/argocd-notifications/shared/argocd/mocks"
	"github.com/argoproj-labs/argocd-notifications/shared/k8s"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes/fake"
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
	assert.Equal(t, DefaultSubscriptions([]DefaultSubscription{
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

func TestWatchConfig_Named(t *testing.T) {
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
	cfgCn := make(chan Config)
	err := WatchConfig(ctx, argocdService, clientset, "default", func(cfg Config) error {
		cfgCn <- cfg
		return nil
	})

	if !assert.NoError(t, err) {
		return
	}

	parsedCfg := <-cfgCn

	assert.Equal(t, "https://myargocd.com", parsedCfg.Context["argocdUrl"])
}

func TestWatchConfig_Labeled(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	configMap1 := &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "random-name1",
			Namespace: "default",
			Labels: map[string]string{
				partOfLabel: "argocd-notifications",
			},
		},
		Data: map[string]string{
			"context": `
argocdUrl: https://myargocd.com
`,
		},
	}
	configMap2 := &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "random-name2",
			Namespace: "default",
			Labels: map[string]string{
				partOfLabel: "argocd-notifications",
			},
		},
		Data: map[string]string{
			"service.slack": `
token: $slackToken
`,
		},
	}

	secret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "random-name",
			Namespace: "default",
		},
		Data: map[string][]byte{
			"slackToken": []byte("abc"),
		},
	}

	argocdService := mocks.NewMockService(ctrl)
	clientset := fake.NewSimpleClientset(configMap1, configMap2, secret)
	cfgCn := make(chan Config)
	err := WatchConfig(ctx, argocdService, clientset, "default", func(cfg Config) error {
		cfgCn <- cfg
		return nil
	})

	if !assert.NoError(t, err) {
		return
	}

	parsedCfg := <-cfgCn

	assert.Equal(t, "https://myargocd.com", parsedCfg.Context["argocdUrl"])
	_, ok := parsedCfg.Services["slack"]
	assert.True(t, ok)
}
