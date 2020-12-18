package slack

import (
	"context"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/tools/cache"

	"github.com/argoproj-labs/argocd-notifications/shared/k8s"
	testingutil "github.com/argoproj-labs/argocd-notifications/testing"
)

func syncedInformers(t *testing.T, ctx context.Context, objects ...runtime.Object) (cache.SharedIndexInformer, cache.SharedIndexInformer) {
	objects = append(objects, &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: k8s.SecretName, Namespace: testingutil.TestNamespace},
	})
	clientset := fake.NewSimpleClientset(objects...)
	cmInformer := k8s.NewConfigMapInformer(clientset, testingutil.TestNamespace)
	secretInformer := k8s.NewSecretInformer(clientset, testingutil.TestNamespace)
	go cmInformer.Run(ctx.Done())
	go secretInformer.Run(ctx.Done())
	if !cache.WaitForCacheSync(context.Background().Done(), cmInformer.HasSynced, secretInformer.HasSynced) {
		t.Fatal("Timed out waiting for caches to sync")
	}

	return cmInformer, secretInformer
}

func notificationConfigMap(data map[string]string) *v1.ConfigMap {
	return &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{Name: k8s.ConfigMapName, Namespace: testingutil.TestNamespace},
		Data:       data,
	}
}

func TestNewVerifier_IncorrectConfig(t *testing.T) {
	testCases := map[string]struct {
		ConfigMap *v1.ConfigMap
		Error     string
	}{
		"NoConfigMap": {
			ConfigMap: nil,
			Error:     "cannot find config map",
		},
		"IncorrectConfigMap": {
			ConfigMap: notificationConfigMap(map[string]string{"service.slack": "bad"}),
			Error:     "unable to parse slack configuration",
		},
		"NoSlack": {
			ConfigMap: notificationConfigMap(map[string]string{"service.email": "{}"}),
			Error:     "slack is not configured",
		},
		"SlackWithoutSigningSecret": {
			ConfigMap: notificationConfigMap(map[string]string{"service.slack": "{}"}),
			Error:     "slack signing secret is not configured",
		},
	}

	for k := range testCases {
		testCase := testCases[k]

		t.Run(k, func(t *testing.T) {

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			var objs []runtime.Object
			if testCase.ConfigMap != nil {
				objs = append(objs, testCase.ConfigMap)
			}
			cmInformer, secretInformer := syncedInformers(t, ctx, objs...)
			verifier := NewVerifier(cmInformer, secretInformer)

			err := verifier(nil, nil)

			assert.Error(t, err)
			assert.Contains(t, err.Error(), testCase.Error)
		})
	}
}

func TestNewVerifier_IncorrectSignature(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cmInformer, secretInformer := syncedInformers(t, ctx, notificationConfigMap(map[string]string{
		"service.slack": `{signingSecret: "helloworld"}`,
	}))

	verifier := NewVerifier(cmInformer, secretInformer)
	now := time.Now()
	data := "hello world"
	err := verifier([]byte(data), map[string][]string{
		"X-Slack-Request-Timestamp": {strconv.Itoa(int(now.UnixNano()))},
		"X-Slack-Signature":         {"v0=9e3753bb47fd3495894ab133c423ec93eff1ff30dd905ce39dda065e21ed9255"},
	})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Expected signing signature")
}
