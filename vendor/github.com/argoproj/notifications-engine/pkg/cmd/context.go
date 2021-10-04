package cmd

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"path/filepath"
	"time"

	"github.com/ghodss/yaml"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	kubeyaml "k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/dynamic"
	informersv1 "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"

	"github.com/argoproj/notifications-engine/pkg/api"
)

type commandContext struct {
	api.Settings
	resource      schema.GroupVersionResource
	dynamicClient dynamic.Interface
	k8sClient     kubernetes.Interface
	cliName       string
	configMapPath string
	secretPath    string
	stdout        io.Writer
	stdin         io.Reader
	stderr        io.Writer
	namespace     string
}

func splitYAML(yamlData []byte) ([]*unstructured.Unstructured, error) {
	d := kubeyaml.NewYAMLOrJSONDecoder(bytes.NewReader(yamlData), 4096)
	var objs []*unstructured.Unstructured
	for {
		ext := runtime.RawExtension{}
		if err := d.Decode(&ext); err != nil {
			if err == io.EOF {
				break
			}
			return objs, fmt.Errorf("failed to unmarshal manifest: %v", err)
		}
		ext.Raw = bytes.TrimSpace(ext.Raw)
		if len(ext.Raw) == 0 || bytes.Equal(ext.Raw, []byte("null")) {
			continue
		}
		u := &unstructured.Unstructured{}
		if err := yaml.Unmarshal(ext.Raw, u); err != nil {
			return objs, fmt.Errorf("failed to unmarshal manifest: %v", err)
		}
		objs = append(objs, u)
	}
	return objs, nil
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
	objs, err := splitYAML(data)
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

func (c *commandContext) loadResource(name string) (*unstructured.Unstructured, error) {
	if ext := filepath.Ext(name); ext != "" {
		data, err := ioutil.ReadFile(name)
		if err != nil {
			return nil, err
		}
		var res unstructured.Unstructured
		err = yaml.Unmarshal(data, &res)
		return &res, err
	}
	res, err := c.dynamicClient.Resource(c.resource).Namespace(c.namespace).Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (c *commandContext) getSecret() (*v1.Secret, error) {
	var secret v1.Secret
	if c.secretPath == ":empty" {
		secret = v1.Secret{}
	} else if c.secretPath == "" {
		s, err := c.k8sClient.CoreV1().Secrets(c.namespace).Get(context.Background(), c.SecretName, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}
		secret = *s
	} else {
		if err := c.unmarshalFromFile(c.secretPath, c.SecretName, schema.GroupKind{Kind: "Secret"}, &secret); err != nil {
			return nil, err
		}
	}
	secret.Name = c.SecretName
	secret.Namespace = c.namespace
	return &secret, nil
}

func (c *commandContext) getConfigMap() (*v1.ConfigMap, error) {
	var configMap v1.ConfigMap
	if c.configMapPath == "" {
		cm, err := c.k8sClient.CoreV1().ConfigMaps(c.namespace).Get(context.Background(), c.ConfigMapName, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}
		configMap = *cm
	} else {
		if err := c.unmarshalFromFile(c.configMapPath, c.ConfigMapName, schema.GroupKind{Kind: "ConfigMap"}, &configMap); err != nil {
			return nil, err
		}
	}
	configMap.Name = c.ConfigMapName
	configMap.Namespace = c.namespace
	return &configMap, nil
}

func (c *commandContext) getAPI() (api.API, error) {
	secretInformer := informersv1.NewSecretInformer(c.k8sClient, c.namespace, time.Minute*3, cache.Indexers{})
	s, err := c.getSecret()
	if err != nil {
		return nil, err
	}
	if err := secretInformer.GetStore().Add(s); err != nil {
		return nil, err
	}
	cm, err := c.getConfigMap()
	if err != nil {
		return nil, err
	}
	cmInformer := informersv1.NewConfigMapInformer(c.k8sClient, c.namespace, time.Minute*3, cache.Indexers{})
	if err := cmInformer.GetStore().Add(cm); err != nil {
		return nil, err
	}

	return api.NewFactory(c.Settings, c.namespace, secretInformer, cmInformer).GetAPI()
}
