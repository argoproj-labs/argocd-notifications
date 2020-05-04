package tools

import (
	"io"
	"io/ioutil"
	"path/filepath"

	"github.com/ghodss/yaml"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/argoproj-labs/argocd-notifications/notifiers"
	"github.com/argoproj-labs/argocd-notifications/shared/clients"
	"github.com/argoproj-labs/argocd-notifications/shared/settings"
	"github.com/argoproj-labs/argocd-notifications/triggers"
)

type commandContext struct {
	configMapPath string
	secretPath    string
	defaultCfg    settings.Config
	stdout        io.Writer
	stderr        io.Writer
	getK8SClients func() (kubernetes.Interface, dynamic.Interface, string, error)
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
	if c.secretPath == ":dummy" {
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

	return settings.ParseConfig(&configMap, &secret, c.defaultCfg)
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
