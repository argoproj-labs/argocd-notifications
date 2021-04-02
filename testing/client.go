package testing

import (
	"encoding/json"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic/fake"
	"k8s.io/client-go/testing"
)

func NewFakeClient(objects ...runtime.Object) *fake.FakeDynamicClient {
	return fake.NewSimpleDynamicClientWithCustomListKinds(runtime.NewScheme(), map[schema.GroupVersionResource]string{
		schema.GroupVersionResource{Group: "argoproj.io", Resource: "applications", Version: "v1alpha1"}: "List",
		schema.GroupVersionResource{Group: "argoproj.io", Resource: "appprojects", Version: "v1alpha1"}:  "List",
	}, objects...)
}

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
