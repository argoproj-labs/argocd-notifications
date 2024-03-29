package k8s

import (
	"fmt"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	corev1 "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

var (
	ConfigMapName = "argocd-notifications-cm"
	SecretName    = "argocd-notifications-secret"
)

const (
	settingsResyncDuration = 3 * time.Minute
)

func NewSecretInformer(clientset kubernetes.Interface, namespace string) cache.SharedIndexInformer {
	return corev1.NewFilteredSecretInformer(clientset, namespace, settingsResyncDuration, cache.Indexers{}, func(options *metav1.ListOptions) {
		options.FieldSelector = fmt.Sprintf("metadata.name=%s", SecretName)
	})
}

func NewConfigMapInformer(clientset kubernetes.Interface, namespace string) cache.SharedIndexInformer {
	return corev1.NewFilteredConfigMapInformer(clientset, namespace, settingsResyncDuration, cache.Indexers{}, func(options *metav1.ListOptions) {
		options.FieldSelector = fmt.Sprintf("metadata.name=%s", ConfigMapName)
	})
}
