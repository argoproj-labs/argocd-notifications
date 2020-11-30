package recipients

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/utils/pointer"
)

func TestGetRecipientsFromAnnotations_NoTriggerNameInAnnotation(t *testing.T) {
	recipients := GetRecipientsFromAnnotations(
		map[string]string{RecipientsAnnotation: "slack:test"}, "on-app-sync-unknown")
	assert.ElementsMatch(t, recipients, []string{"slack:test"})
}

func TestGetRecipientsFromAnnotations_HasTriggerNameInAnnotation(t *testing.T) {
	recipients := GetRecipientsFromAnnotations(map[string]string{
		RecipientsAnnotation: "slack:test",
		fmt.Sprintf("on-app-sync-unknown.%s", RecipientsAnnotation):    "slack:test1",
		fmt.Sprintf("on-app-health-degraded.%s", RecipientsAnnotation): "slack:test2",
	}, "on-app-sync-unknown")
	assert.ElementsMatch(t, recipients, []string{"slack:test", "slack:test1"})
}

func TestCopyStringMap(t *testing.T) {
	in := map[string]string{"key": "val"}
	out := CopyStringMap(in)
	assert.Equal(t, in, out)
	assert.False(t, &in == &out)
}

func TestAnnotationsPatch(t *testing.T) {
	oldAnnotations := map[string]string{"key1": "val1", "key2": "val2", "key3": "val3"}
	newAnnotations := map[string]string{"key2": "val2-updated", "key3": "val3", "key4": "val4"}
	patch := AnnotationsPatch(oldAnnotations, newAnnotations)
	assert.Equal(t, map[string]*string{
		"key1": nil,
		"key2": pointer.StringPtr("val2-updated"),
		"key4": pointer.StringPtr("val4"),
	}, patch)
}
