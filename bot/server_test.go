package bot

import (
	"fmt"
	"testing"

	"k8s.io/utils/pointer"

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
		NewApp("bar", WithAnnotations(map[string]string{recipients.AnnotationKey: "slack:general"})))
	s := NewServer(client, TestNamespace)

	response, err := s.listSubscriptions("slack:general")

	assert.NoError(t, err)

	assert.Contains(t, response, "Applications: default/bar")
}

func TestListSubscriptions_HasAppProjSubscription(t *testing.T) {
	client := fake.NewSimpleDynamicClient(runtime.NewScheme(),
		NewApp("foo"),
		NewProject("bar", WithAnnotations(map[string]string{recipients.AnnotationKey: "slack:general"})))
	s := NewServer(client, TestNamespace)

	response, err := s.listSubscriptions("slack:general")

	assert.NoError(t, err)

	assert.Contains(t, response, "Projects: default/bar")
}

func TestUpdateSubscription_SubscribeToApp(t *testing.T) {
	client := fake.NewSimpleDynamicClient(runtime.NewScheme(), NewApp("foo", WithAnnotations(map[string]string{
		recipients.AnnotationKey: "slack:channel1",
	})))

	var patches []map[string]interface{}
	AddPatchCollectorReactor(client, &patches)

	s := NewServer(client, TestNamespace)

	resp, err := s.updateSubscription("slack:channel2", true, UpdateSubscription{App: "foo"})
	assert.NoError(t, err)
	assert.Equal(t, "subscription updated", resp)
	assert.Len(t, patches, 1)

	val, _, _ := unstructured.NestedString(patches[0], "metadata", "annotations", recipients.AnnotationKey)
	assert.Equal(t, val, "slack:channel1,slack:channel2")
}

func TestUpdateSubscription_SubscribeToAppTrigger(t *testing.T) {
	client := fake.NewSimpleDynamicClient(runtime.NewScheme(), NewApp("foo", WithAnnotations(map[string]string{
		recipients.AnnotationKey: "slack:channel1",
	})))

	var patches []map[string]interface{}
	AddPatchCollectorReactor(client, &patches)

	s := NewServer(client, TestNamespace)

	resp, err := s.updateSubscription("slack:channel2", true, UpdateSubscription{App: "foo", Trigger: "on-sync-failed"})
	assert.NoError(t, err)
	assert.Equal(t, "subscription updated", resp)
	assert.Len(t, patches, 1)

	patch := patches[0]
	val, _, _ := unstructured.NestedString(patch, "metadata", "annotations", fmt.Sprintf("on-sync-failed.%s", recipients.AnnotationKey))
	assert.Equal(t, val, "slack:channel2")
}

func TestUpdateSubscription_UnsubscribeAppTrigger(t *testing.T) {
	client := fake.NewSimpleDynamicClient(runtime.NewScheme(), NewApp("foo", WithAnnotations(map[string]string{
		recipients.AnnotationKey:                                   "slack:channel1,slack:channel2",
		fmt.Sprintf("on-sync-failed.%s", recipients.AnnotationKey): "slack:channel1,slack:channel2",
	})))

	var patches []map[string]interface{}
	AddPatchCollectorReactor(client, &patches)

	s := NewServer(client, TestNamespace)

	resp, err := s.updateSubscription("slack:channel2", false, UpdateSubscription{App: "foo", Trigger: "on-sync-failed"})
	assert.NoError(t, err)
	assert.Equal(t, "subscription updated", resp)
	assert.Len(t, patches, 1)

	patch := patches[0]
	val, _, _ := unstructured.NestedString(patch, "metadata", "annotations", fmt.Sprintf("on-sync-failed.%s", recipients.AnnotationKey))
	assert.Equal(t, val, "slack:channel1")
	val, _, _ = unstructured.NestedString(patch, "metadata", "annotations", recipients.AnnotationKey)
	assert.Equal(t, val, "slack:channel1")
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
