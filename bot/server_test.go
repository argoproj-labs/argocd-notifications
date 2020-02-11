package bot

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/argoproj-labs/argocd-notifications/shared/recipients"
	testingutil "github.com/argoproj-labs/argocd-notifications/testing"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic/fake"
)

func TestListRecipients_NoSubscriptions(t *testing.T) {
	client := fake.NewSimpleDynamicClient(runtime.NewScheme())
	s := NewServer(client, testingutil.TestNamespace)

	response, err := s.listSubscriptions("slack:general")

	assert.NoError(t, err)

	assert.Contains(t, response, "The slack:general has no subscriptions.")
}

func TestListRecipients_HasAppSubscription(t *testing.T) {
	client := fake.NewSimpleDynamicClient(runtime.NewScheme(),
		testingutil.NewApp("foo"),
		testingutil.NewApp("bar",
			testingutil.WithAnnotations(map[string]string{recipients.RecipientsAnnotation: "slack:general"})))
	s := NewServer(client, testingutil.TestNamespace)

	response, err := s.listSubscriptions("slack:general")

	assert.NoError(t, err)

	assert.Contains(t, response, "Applications: default/bar")
}

func TestListRecipients_HasAppProjSubscription(t *testing.T) {
	client := fake.NewSimpleDynamicClient(runtime.NewScheme(),
		testingutil.NewApp("foo"),
		testingutil.NewProject("bar",
			testingutil.WithAnnotations(map[string]string{recipients.RecipientsAnnotation: "slack:general"})))
	s := NewServer(client, testingutil.TestNamespace)

	response, err := s.listSubscriptions("slack:general")

	assert.NoError(t, err)

	assert.Contains(t, response, "Projects: default/bar")
}
