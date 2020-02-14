package testing

import (
	"encoding/json"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic/fake"
	"k8s.io/client-go/testing"
)

func AddPatchCollectorReactor(client *fake.FakeDynamicClient, patches *[]map[string]interface{}) {
	client.PrependReactor("patch", "*", func(action testing.Action) (handled bool, ret runtime.Object, err error) {
		if patchAction, ok := action.(testing.PatchAction); ok {
			patch := make(map[string]interface{})
			if err := json.Unmarshal(patchAction.GetPatch(), &patch); err != nil {
				return false, nil, err
			} else {
				*patches = append(*patches, patch)
			}
		}
		return true, nil, nil
	})
}
