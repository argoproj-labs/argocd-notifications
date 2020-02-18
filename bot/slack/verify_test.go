package slack

import (
	"context"
	"strconv"
	"testing"
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/tools/cache"

	"github.com/argoproj-labs/argocd-notifications/shared/settings"
	testingutil "github.com/argoproj-labs/argocd-notifications/testing"
	"github.com/stretchr/testify/assert"
)

func syncedInformer(t *testing.T, ctx context.Context, objects ...runtime.Object) cache.SharedIndexInformer {
	informer := settings.NewSecretInformer(fake.NewSimpleClientset(objects...), testingutil.TestNamespace)
	go informer.Run(ctx.Done())
	if !cache.WaitForCacheSync(context.Background().Done(), informer.HasSynced) {
		t.Fatal("Timed out waiting for caches to sync")
	}
	return informer
}

func notificationSecret(data map[string][]byte) *v1.Secret {
	return &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: settings.SecretName, Namespace: testingutil.TestNamespace},
		Data:       data,
	}
}

func TestNewVerifier_IncorrectConfig(t *testing.T) {
	testCases := map[string]struct {
		Secret *v1.Secret
		Error  string
	}{
		"NoSecret": {
			Secret: nil,
			Error:  "cannot find secret",
		},
		"IncorrectSecret": {
			Secret: notificationSecret(map[string][]byte{"notifiers.yaml": []byte("bad")}),
			Error:  "unable to parse slack configuration",
		},
		"NoSlack": {
			Secret: notificationSecret(map[string][]byte{"notifiers.yaml": []byte("email: {}")}),
			Error:  "slack is not configured",
		},
		"SlackWithoutSigningSecret": {
			Secret: notificationSecret(map[string][]byte{"notifiers.yaml": []byte("slack: {}")}),
			Error:  "slack signing secret is not configured",
		},
	}

	for k := range testCases {
		testCase := testCases[k]

		t.Run(k, func(t *testing.T) {

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			var objs []runtime.Object
			if testCase.Secret != nil {
				objs = append(objs, testCase.Secret)
			}
			verifier := NewVerifier(syncedInformer(t, ctx, objs...))

			err := verifier(nil, nil)

			assert.Error(t, err)
			assert.Contains(t, err.Error(), testCase.Error)
		})
	}
}

func TestNewVerifier_IncorrectSignature(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	verifier := NewVerifier(syncedInformer(t, ctx, notificationSecret(map[string][]byte{
		"notifiers.yaml": []byte(`slack: {signingSecret: "helloworld"}`),
	})))
	now := time.Now()
	data := "hello world"
	err := verifier([]byte(data), map[string][]string{
		"X-Slack-Request-Timestamp": {strconv.Itoa(int(now.UnixNano()))},
		"X-Slack-Signature":         {"v0=9e3753bb47fd3495894ab133c423ec93eff1ff30dd905ce39dda065e21ed9255"},
	})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Expected signing signature")
}
