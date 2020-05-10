package builtin

import (
	"fmt"
	"testing"

	"k8s.io/utils/pointer"

	. "github.com/argoproj-labs/argocd-notifications/testing"
	"github.com/argoproj-labs/argocd-notifications/triggers"
	"github.com/stretchr/testify/assert"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type testCase struct {
	positiveInputs []*unstructured.Unstructured
	negativeInputs []*unstructured.Unstructured
}

var testCases = map[string]testCase{
	"on-sync-status-unknown": {
		positiveInputs: []*unstructured.Unstructured{NewApp("test", WithSyncStatus("Unknown"))},
		negativeInputs: []*unstructured.Unstructured{NewApp("test", WithSyncStatus("Synced"))},
	},
	"on-sync-failed": {
		positiveInputs: []*unstructured.Unstructured{
			NewApp("test", WithSyncOperationPhase("Failed")),
			NewApp("test", WithSyncOperationPhase("Error")),
		},
		negativeInputs: []*unstructured.Unstructured{NewApp("test", WithSyncOperationPhase("Running"))},
	},
	"on-sync-running": {
		positiveInputs: []*unstructured.Unstructured{NewApp("test", WithSyncOperationPhase("Running"))},
		negativeInputs: []*unstructured.Unstructured{NewApp("test", WithSyncOperationPhase("Failed"))},
	},
	"on-sync-succeeded": {
		positiveInputs: []*unstructured.Unstructured{NewApp("test", WithSyncOperationPhase("Succeeded"))},
		negativeInputs: []*unstructured.Unstructured{NewApp("test", WithSyncOperationPhase("Running"))},
	},
	"on-health-degraded": {
		positiveInputs: []*unstructured.Unstructured{NewApp("test", WithHealthStatus("Degraded"))},
		negativeInputs: []*unstructured.Unstructured{NewApp("test", WithHealthStatus("Progressing"))},
	},
}

func TestBuiltInTriggers(t *testing.T) {
	for _, trigger := range Triggers {
		trigger.Enabled = pointer.BoolPtr(true)
		t.Run(fmt.Sprintf("TestTrigger_%s", trigger.Name), func(t *testing.T) {
			if testCase, ok := testCases[trigger.Name]; !ok {
				t.Fatalf("No tests for trigger %s", trigger.Name)
			} else {
				builtInTriggers, err := triggers.GetTriggers(Templates, []triggers.NotificationTrigger{trigger}, nil)
				assert.NoError(t, err)
				item := builtInTriggers[trigger.Name]
				for i := range testCase.negativeInputs {
					res, err := item.Triggered(testCase.negativeInputs[i])
					assert.NoError(t, err)
					assert.False(t, res)
				}
				assert.True(t, len(testCase.negativeInputs) > 0, "at least one negative input is required")
				for i := range testCase.positiveInputs {
					res, err := item.Triggered(testCase.positiveInputs[i])
					assert.NoError(t, err)
					assert.True(t, res)
				}
				assert.True(t, len(testCase.positiveInputs) > 0, "at least one positive input is required")
			}
		})
	}
}
