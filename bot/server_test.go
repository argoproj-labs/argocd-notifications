package bot

import (
	"fmt"
	"testing"

	"github.com/argoproj-labs/argocd-notifications/shared/recipients"
	. "github.com/argoproj-labs/argocd-notifications/testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic/fake"
)

func TestListRecipients_NoSubscriptions(t *testing.T) {
	client := fake.NewSimpleDynamicClient(runtime.NewScheme())
	s := NewServer(client, TestNamespace)

	response, err := s.listSubscriptions("slack:general")

	assert.NoError(t, err)

	assert.Contains(t, response, "The slack:general has no subscriptions.")
}

func TestListSubscriptions_HasAppSubscription(t *testing.T) {
	client := fake.NewSimpleDynamicClient(runtime.NewScheme(),
		NewApp("foo"),
		NewApp("bar", WithAnnotations(map[string]string{recipients.RecipientsAnnotation: "slack:general"})))
	s := NewServer(client, TestNamespace)

	response, err := s.listSubscriptions("slack:general")

	assert.NoError(t, err)

	assert.Contains(t, response, "Applications: default/bar")
}

func TestListSubscriptions_HasAppProjSubscription(t *testing.T) {
	client := fake.NewSimpleDynamicClient(runtime.NewScheme(),
		NewApp("foo"),
		NewProject("bar", WithAnnotations(map[string]string{recipients.RecipientsAnnotation: "slack:general"})))
	s := NewServer(client, TestNamespace)

	response, err := s.listSubscriptions("slack:general")

	assert.NoError(t, err)

	assert.Contains(t, response, "Projects: default/bar")
}

func TestUpdateSubscription_SubscribeToApp(t *testing.T) {
	client := fake.NewSimpleDynamicClient(runtime.NewScheme(), NewApp("foo", WithAnnotations(map[string]string{
		recipients.RecipientsAnnotation: "slack:channel1",
	})))

	var patches []map[string]interface{}
	AddPatchCollectorReactor(client, &patches)

	s := NewServer(client, TestNamespace)

	resp, err := s.updateSubscription("slack:channel2", true, UpdateSubscription{App: "foo"})
	assert.NoError(t, err)
	assert.Equal(t, "subscription updated", resp)
	assert.Len(t, patches, 1)

	val, _, _ := unstructured.NestedString(patches[0], "metadata", "annotations", recipients.RecipientsAnnotation)
	assert.Equal(t, val, "slack:channel1,slack:channel2")
}

func TestUpdateSubscription_SubscribeToAppTrigger(t *testing.T) {
	client := fake.NewSimpleDynamicClient(runtime.NewScheme(), NewApp("foo", WithAnnotations(map[string]string{
		recipients.RecipientsAnnotation: "slack:channel1",
	})))

	var patches []map[string]interface{}
	AddPatchCollectorReactor(client, &patches)

	s := NewServer(client, TestNamespace)

	resp, err := s.updateSubscription("slack:channel2", true, UpdateSubscription{App: "foo", Trigger: "on-sync-failed"})
	assert.NoError(t, err)
	assert.Equal(t, "subscription updated", resp)
	assert.Len(t, patches, 1)

	patch := patches[0]
	val, _, _ := unstructured.NestedString(patch, "metadata", "annotations", fmt.Sprintf("on-sync-failed.%s", recipients.RecipientsAnnotation))
	assert.Equal(t, val, "slack:channel2")
}

func TestUpdateSubscription_UnsubscribeAppTrigger(t *testing.T) {
	client := fake.NewSimpleDynamicClient(runtime.NewScheme(), NewApp("foo", WithAnnotations(map[string]string{
		recipients.RecipientsAnnotation:                                   "slack:channel1,slack:channel2",
		fmt.Sprintf("on-sync-failed.%s", recipients.RecipientsAnnotation): "slack:channel1,slack:channel2",
	})))

	var patches []map[string]interface{}
	AddPatchCollectorReactor(client, &patches)

	s := NewServer(client, TestNamespace)

	resp, err := s.updateSubscription("slack:channel2", false, UpdateSubscription{App: "foo", Trigger: "on-sync-failed"})
	assert.NoError(t, err)
	assert.Equal(t, "subscription updated", resp)
	assert.Len(t, patches, 1)

	patch := patches[0]
	val, _, _ := unstructured.NestedString(patch, "metadata", "annotations", fmt.Sprintf("on-sync-failed.%s", recipients.RecipientsAnnotation))
	assert.Equal(t, val, "slack:channel1")
	val, _, _ = unstructured.NestedString(patch, "metadata", "annotations", recipients.RecipientsAnnotation)
	assert.Equal(t, val, "slack:channel1")
}
