package bot

import (
	"testing"

	"github.com/argoproj/notifications-engine/pkg/subscriptions"

	. "github.com/argoproj-labs/argocd-notifications/testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/utils/pointer"
)

func TestListRecipients_NoSubscriptions(t *testing.T) {
	client := NewFakeClient()
	s := NewServer(client, TestNamespace)

	response, err := s.listSubscriptions("slack", "general")

	assert.NoError(t, err)

	assert.Contains(t, "The general has no subscriptions.", response)
}

func TestListSubscriptions_HasAppSubscription(t *testing.T) {
	client := NewFakeClient(
		NewApp("foo"),
		NewApp("bar", WithAnnotations(map[string]string{subscriptions.SubscribeAnnotationKey("my-trigger", "slack"): "general"})))
	s := NewServer(client, TestNamespace)

	response, err := s.listSubscriptions("slack", "general")

	assert.NoError(t, err)

	assert.Contains(t, response, "Applications: default/bar")
}

func TestListSubscriptions_HasAppProjSubscription(t *testing.T) {
	client := NewFakeClient(
		NewApp("foo"),
		NewProject("bar", WithAnnotations(map[string]string{subscriptions.SubscribeAnnotationKey("my-trigger", "slack"): "general"})))
	s := NewServer(client, TestNamespace)

	response, err := s.listSubscriptions("slack", "general")

	assert.NoError(t, err)

	assert.Contains(t, response, "Projects: default/bar")
}

func TestUpdateSubscription_SubscribeToApp(t *testing.T) {
	client := NewFakeClient(NewApp("foo", WithAnnotations(map[string]string{
		subscriptions.SubscribeAnnotationKey("my-trigger", "slack"): "channel1",
	})))

	var patches []map[string]interface{}
	AddPatchCollectorReactor(client, &patches)

	s := NewServer(client, TestNamespace)

	resp, err := s.updateSubscription("slack", "channel2", true, UpdateSubscription{App: "foo", Trigger: "my-trigger"})
	assert.NoError(t, err)
	assert.Equal(t, "subscription updated", resp)
	assert.Len(t, patches, 1)

	val, _, _ := unstructured.NestedString(patches[0], "metadata", "annotations", subscriptions.SubscribeAnnotationKey("my-trigger", "slack"))
	assert.Equal(t, val, "channel1;channel2")
}

func TestUpdateSubscription_SubscribeToAppTrigger(t *testing.T) {
	client := NewFakeClient(NewApp("foo", WithAnnotations(map[string]string{
		subscriptions.SubscribeAnnotationKey("my-trigger", "slack"): "channel1",
	})))

	var patches []map[string]interface{}
	AddPatchCollectorReactor(client, &patches)

	s := NewServer(client, TestNamespace)

	resp, err := s.updateSubscription("slack", "channel2", true, UpdateSubscription{App: "foo", Trigger: "on-sync-failed"})
	assert.NoError(t, err)
	assert.Equal(t, "subscription updated", resp)
	assert.Len(t, patches, 1)

	patch := patches[0]
	val, _, _ := unstructured.NestedString(patch, "metadata", "annotations", subscriptions.SubscribeAnnotationKey("on-sync-failed", "slack"))
	assert.Equal(t, "channel2", val)
}

func TestCopyStringMap(t *testing.T) {
	in := map[string]string{"key": "val"}
	out := copyStringMap(in)
	assert.Equal(t, in, out)
	assert.False(t, &in == &out)
}

func TestAnnotationsPatch(t *testing.T) {
	oldAnnotations := map[string]string{"key1": "val1", "key2": "val2", "key3": "val3"}
	newAnnotations := map[string]string{"key2": "val2-updated", "key3": "val3", "key4": "val4"}
	patch := annotationsPatch(oldAnnotations, newAnnotations)
	assert.Equal(t, map[string]*string{
		"key1": nil,
		"key2": pointer.StringPtr("val2-updated"),
		"key4": pointer.StringPtr("val4"),
	}, patch)
}
