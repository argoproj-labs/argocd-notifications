package k8s

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

var (
	Applications = schema.GroupVersionResource{Group: "argoproj.io", Version: "v1alpha1", Resource: "applications"}
	AppProjects  = schema.GroupVersionResource{Group: "argoproj.io", Version: "v1alpha1", Resource: "appprojects"}
)

func NewAppClient(client dynamic.Interface, namespace string) dynamic.ResourceInterface {
	resClient := client.Resource(Applications).Namespace(namespace)
	return resClient
}

func NewAppProjClient(client dynamic.Interface, namespace string) dynamic.ResourceInterface {
	resClient := client.Resource(AppProjects).Namespace(namespace)
	return resClient
}
