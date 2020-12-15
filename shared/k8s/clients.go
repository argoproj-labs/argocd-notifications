package k8s

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

func NewAppClient(client dynamic.Interface, namespace string) dynamic.ResourceInterface {
	appResource := schema.GroupVersionResource{Group: "argoproj.io", Version: "v1alpha1", Resource: "applications"}
	resClient := client.Resource(appResource).Namespace(namespace)
	return resClient
}

func NewAppProjClient(client dynamic.Interface, namespace string) dynamic.ResourceInterface {
	appResource := schema.GroupVersionResource{Group: "argoproj.io", Version: "v1alpha1", Resource: "appprojects"}
	resClient := client.Resource(appResource).Namespace(namespace)
	return resClient
}
