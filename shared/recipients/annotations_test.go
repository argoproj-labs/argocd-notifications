package recipients

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
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
