package tools

import (
	"context"
	"io"
	"io/ioutil"
	"path/filepath"
	"sync"

	"github.com/ghodss/yaml"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/argoproj-labs/argocd-notifications/notifiers"
	"github.com/argoproj-labs/argocd-notifications/shared/argocd"
	"github.com/argoproj-labs/argocd-notifications/shared/clients"
	"github.com/argoproj-labs/argocd-notifications/shared/settings"
	"github.com/argoproj-labs/argocd-notifications/triggers"
	"github.com/argoproj-labs/argocd-notifications/triggers/expr/shared"
)

type (
	clientsSource = func() (kubernetes.Interface, dynamic.Interface, string, error)
)

type commandContext struct {
	configMapPath string
	secretPath    string
	stdout        io.Writer
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

func (c *commandContext) getConfig() (map[string]triggers.Trigger, map[string]notifiers.Notifier, *settings.Config, error) {
	var configMap v1.ConfigMap
	if c.configMapPath == "" {
		k8sClient, _, ns, err := c.getK8SClients()
		if err != nil {
			return nil, nil, nil, err
		}
		cm, err := k8sClient.CoreV1().ConfigMaps(ns).Get(settings.ConfigMapName, metav1.GetOptions{})
		if err != nil {
			return nil, nil, nil, err
		}
		configMap = *cm
	} else {
		data, err := ioutil.ReadFile(c.configMapPath)
		if err != nil {
			return nil, nil, nil, err
		}
		if err = yaml.Unmarshal(data, &configMap); err != nil {
			return nil, nil, nil, err
		}
	}

	var secret v1.Secret
	if c.secretPath == ":empty" {
		secret = v1.Secret{}
	} else if c.secretPath == "" {
		k8sClient, _, ns, err := c.getK8SClients()
		if err != nil {
			return nil, nil, nil, err
		}
		s, err := k8sClient.CoreV1().Secrets(ns).Get(settings.SecretName, metav1.GetOptions{})
		if err != nil {
			return nil, nil, nil, err
		}
		secret = *s
	} else {
		data, err := ioutil.ReadFile(c.secretPath)
		if err != nil {
			return nil, nil, nil, err
		}
		if err = yaml.Unmarshal(data, &secret); err != nil {
			return nil, nil, nil, err
		}
	}
	return settings.ParseConfig(&configMap, &secret, settings.Config{}, &lazyArgocdServiceInitializer{})
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
	app, err := clients.NewAppClient(client, ns).Get(application, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	return app, nil
}
