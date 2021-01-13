package tools

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"path/filepath"
	"sync"

	"github.com/argoproj/gitops-engine/pkg/utils/kube"
	"github.com/ghodss/yaml"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/argoproj-labs/argocd-notifications/expr/shared"
	"github.com/argoproj-labs/argocd-notifications/shared/argocd"
	"github.com/argoproj-labs/argocd-notifications/shared/k8s"
	"github.com/argoproj-labs/argocd-notifications/shared/legacy"
	"github.com/argoproj-labs/argocd-notifications/shared/settings"
)

type (
	clientsSource = func() (kubernetes.Interface, dynamic.Interface, string, error)
)

type commandContext struct {
	configMapPath string
	secretPath    string
	stdout        io.Writer
	stdin         io.Reader
	stderr        io.Writer
	getK8SClients clientsSource
	argocdService *lazyArgocdServiceInitializer
}

type lazyArgocdServiceInitializer struct {
	argocdRepoServer *string
	argocdService    argocd.Service
	init             sync.Once
	getK8SClients    clientsSource
}

func (svc *lazyArgocdServiceInitializer) initArgoCDService() error {
	k8sClient, _, ns, err := svc.getK8SClients()
	if err != nil {
		return err
	}
	argocdService, err := argocd.NewArgoCDService(k8sClient, ns, *svc.argocdRepoServer)
	if err != nil {
		return err
	}
	svc.argocdService = argocdService
	return nil
}

func (svc *lazyArgocdServiceInitializer) GetCommitMetadata(ctx context.Context, repoURL string, commitSHA string) (*shared.CommitMetadata, error) {
	var err error
	svc.init.Do(func() {
		err = svc.initArgoCDService()
	})
	if err != nil {
		return nil, err
	}
	return svc.argocdService.GetCommitMetadata(ctx, repoURL, commitSHA)
}

func getK8SClients(clientConfig clientcmd.ClientConfig) (kubernetes.Interface, dynamic.Interface, string, error) {
	ns, _, err := clientConfig.Namespace()
	if err != nil {
		return nil, nil, "", err
	}
	config, err := clientConfig.ClientConfig()
	if err != nil {
		return nil, nil, "", err
	}
	k8sClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, nil, "", err
	}
	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, nil, "", err
	}
	return k8sClient, dynamicClient, ns, nil
}

func (c *commandContext) unmarshalFromFile(filePath string, name string, gk schema.GroupKind, result interface{}) error {
	var err error
	var data []byte
	if filePath == "-" {
		data, err = ioutil.ReadAll(c.stdin)
	} else {
		data, err = ioutil.ReadFile(c.configMapPath)
	}
	if err != nil {
		return err
	}
	objs, err := kube.SplitYAML(data)
	if err != nil {
		return err
	}

	for _, obj := range objs {
		if obj.GetName() == name && obj.GroupVersionKind().GroupKind() == gk {
			return runtime.DefaultUnstructuredConverter.FromUnstructured(obj.Object, result)
		}
	}
	return fmt.Errorf("file '%s' does not have '%s/%s/%s'", filePath, gk.Group, gk.Kind, name)
}

func (c *commandContext) getConfig() (*settings.Config, error) {
	var configMap v1.ConfigMap
	if c.configMapPath == "" {
		k8sClient, _, ns, err := c.getK8SClients()
		if err != nil {
			return nil, err
		}
		cm, err := k8sClient.CoreV1().ConfigMaps(ns).Get(context.Background(), k8s.ConfigMapName, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}
		configMap = *cm
	} else {
		if err := c.unmarshalFromFile(c.configMapPath, k8s.ConfigMapName, schema.GroupKind{Kind: "ConfigMap"}, &configMap); err != nil {
			return nil, err
		}
	}

	var secret v1.Secret
	if c.secretPath == ":empty" {
		secret = v1.Secret{}
	} else if c.secretPath == "" {
		k8sClient, _, ns, err := c.getK8SClients()
		if err != nil {
			return nil, err
		}
		s, err := k8sClient.CoreV1().Secrets(ns).Get(context.Background(), k8s.SecretName, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}
		secret = *s
	} else {
		if err := c.unmarshalFromFile(c.secretPath, k8s.SecretName, schema.GroupKind{Kind: "Secret"}, &secret); err != nil {
			return nil, err
		}
	}
	return settings.NewConfig(&configMap, &secret, &lazyArgocdServiceInitializer{}, legacy.ApplyLegacyConfig)
}

func (c *commandContext) loadApplication(application string) (*unstructured.Unstructured, error) {
	if ext := filepath.Ext(application); ext != "" {
		data, err := ioutil.ReadFile(application)
		if err != nil {
			return nil, err
		}
		var app unstructured.Unstructured
		err = yaml.Unmarshal(data, &app)
		return &app, err
	}
	_, client, ns, err := c.getK8SClients()
	if err != nil {
		return nil, err
	}
	app, err := k8s.NewAppClient(client, ns).Get(context.Background(), application, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	return app, nil
}
