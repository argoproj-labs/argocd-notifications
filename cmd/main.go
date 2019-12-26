package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/argoproj-labs/argocd-notifications/assets"
	"github.com/argoproj-labs/argocd-notifications/controller"
	"github.com/argoproj-labs/argocd-notifications/notifiers"
	"github.com/argoproj-labs/argocd-notifications/triggers"

	"github.com/ghodss/yaml"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/dynamic"
	corev1 "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	configMapName            = "argocd-notifications-cm"
	secretName               = "argocd-notifications-secret"
	settingsResyncDuration   = 3 * time.Minute
	argocdURLContextVariable = "argocdUrl"
)

type config struct {
	Triggers  []triggers.NotificationTrigger  `json:"triggers"`
	Templates []triggers.NotificationTemplate `json:"templates"`
	Context   map[string]string               `json:"context"`
}

var (
	defaultCfg config
)

func init() {
	defaultCfg = config{Context: map[string]string{argocdURLContextVariable: "https://localhost:4000"}}
	err := yaml.Unmarshal([]byte(assets.DefaultConfig), &defaultCfg)
	if err != nil {
		panic(err)
	}
}

func main() {
	if err := newCommand().Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func newCommand() *cobra.Command {
	var (
		clientConfig     clientcmd.ClientConfig
		processorsCount  int
		namespace        string
		appLabelSelector string
	)
	var command = cobra.Command{
		Use: "argocd-notifications",
		RunE: func(c *cobra.Command, args []string) error {
			restConfig, err := clientConfig.ClientConfig()
			if err != nil {
				return err
			}
			dynamicClient, err := dynamic.NewForConfig(restConfig)
			if err != nil {
				return err
			}
			k8sClient, err := kubernetes.NewForConfig(restConfig)
			if err != nil {
				return err
			}
			if namespace == "" {
				namespace, _, err = clientConfig.Namespace()
				if err != nil {
					return err
				}
			}

			var cancelPrev context.CancelFunc
			watchConfig(k8sClient, namespace, func(triggers map[string]triggers.Trigger, notifiers map[string]notifiers.Notifier, contextVals map[string]string) error {
				if cancelPrev != nil {
					log.Info("Settings had been updated. Restarting controller...")
					cancelPrev()
					cancelPrev = nil
				}
				ctrl, err := controller.NewController(dynamicClient, namespace, triggers, notifiers, contextVals, appLabelSelector)
				if err != nil {
					return err
				}
				ctx, cancel := context.WithCancel(context.Background())
				cancelPrev = cancel

				err = ctrl.Init(ctx)
				if err != nil {
					return err
				}

				go ctrl.Run(ctx, processorsCount)
				return nil
			})
			<-context.Background().Done()
			return nil
		},
	}
	clientConfig = addKubectlFlagsToCmd(&command)
	command.Flags().IntVar(&processorsCount, "processors-count", 1, "Processors count.")
	command.Flags().StringVar(&appLabelSelector, "app-label-selector", "", "App label selector.")
	command.Flags().StringVar(&namespace, "namespace", "", "Namespace which controller handles. Current namespace if empty.")
	return &command
}

func parseConfig(configData map[string]string, notifiersData []byte) (map[string]triggers.Trigger, map[string]notifiers.Notifier, map[string]string, error) {
	cfg := &config{}
	if data, ok := configData["config.yaml"]; ok {
		err := yaml.Unmarshal([]byte(data), cfg)
		if err != nil {
			return nil, nil, nil, err
		}
	}

	cfg = defaultCfg.merge(cfg)
	t, err := triggers.GetTriggers(cfg.Templates, cfg.Triggers)
	if err != nil {
		return nil, nil, nil, err
	}

	notifiersConfig := notifiers.Config{}
	err = yaml.Unmarshal(notifiersData, &notifiersConfig)
	if err != nil {
		return nil, nil, nil, err
	}
	return t, notifiers.GetAll(notifiersConfig), cfg.Context, nil
}

func watchConfig(clientset kubernetes.Interface, namespace string, callback func(map[string]triggers.Trigger, map[string]notifiers.Notifier, map[string]string) error) {
	var secret *v1.Secret
	var configMap *v1.ConfigMap
	lock := &sync.Mutex{}
	onNewConfigMapAndSecret := func(newSecret *v1.Secret, newConfigMap *v1.ConfigMap) {
		lock.Lock()
		defer lock.Unlock()
		if newSecret != nil {
			secret = newSecret
		}
		if newConfigMap != nil {
			configMap = newConfigMap
		}
		if secret != nil && configMap != nil {
			if t, n, c, err := parseConfig(configMap.Data, secret.Data["notifiers.yaml"]); err == nil {
				if err = callback(t, n, c); err != nil {
					log.Fatal("Failed to start controller: %v", err)
				}
			} else {
				log.Fatal("Failed to parse new settings: %v", err)
			}
		}
	}

	onConfigMapChanged := func(newObj interface{}) {
		if cm, ok := newObj.(*v1.ConfigMap); ok {
			onNewConfigMapAndSecret(nil, cm)
		}
	}

	cmInformer := corev1.NewFilteredConfigMapInformer(clientset, namespace, settingsResyncDuration, cache.Indexers{}, func(options *metav1.ListOptions) {
		options.FieldSelector = fmt.Sprintf("metadata.name=%s", configMapName)
	})
	cmInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		UpdateFunc: func(oldObj, newObj interface{}) {
			onConfigMapChanged(newObj)
		},
		AddFunc: onConfigMapChanged,
	})

	onSecretChanged := func(newObj interface{}) {
		if s, ok := newObj.(*v1.Secret); ok {
			onNewConfigMapAndSecret(s, nil)
		}
	}

	secretInformer := corev1.NewFilteredSecretInformer(clientset, namespace, settingsResyncDuration, cache.Indexers{}, func(options *metav1.ListOptions) {
		options.FieldSelector = fmt.Sprintf("metadata.name=%s", secretName)
	})
	secretInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: onSecretChanged,
		UpdateFunc: func(oldObj, newObj interface{}) {
			onSecretChanged(newObj)
		},
	})
	go secretInformer.Run(context.Background().Done())
	go cmInformer.Run(context.Background().Done())

	if !cache.WaitForCacheSync(context.Background().Done(), cmInformer.HasSynced, secretInformer.HasSynced) {
		log.Fatal(errors.New("timed out waiting for caches to sync"))
	}
}

func coalesce(first string, other ...string) string {
	res := first
	for i := range other {
		if res != "" {
			break
		}
		res = other[i]
	}
	return res
}

func (cfg *config) merge(other *config) *config {
	triggersMap := map[string]triggers.NotificationTrigger{}
	for i := range cfg.Triggers {
		triggersMap[cfg.Triggers[i].Name] = cfg.Triggers[i]
	}
	for _, item := range other.Triggers {
		if existing, ok := triggersMap[item.Name]; ok {
			existing.Condition = coalesce(item.Condition)
			existing.Template = coalesce(item.Template)
			if item.Enabled != nil {
				existing.Enabled = item.Enabled
			}
			triggersMap[item.Name] = existing
		} else {
			triggersMap[item.Name] = item
		}
	}

	templatesMap := map[string]triggers.NotificationTemplate{}
	for i := range cfg.Templates {
		templatesMap[cfg.Templates[i].Name] = cfg.Templates[i]
	}
	for _, item := range other.Templates {
		if existing, ok := templatesMap[item.Name]; ok {
			existing.Body = coalesce(item.Body)
			existing.Title = coalesce(item.Title)
			templatesMap[item.Name] = existing
		} else {
			templatesMap[item.Name] = item
		}
	}

	contextValues := map[string]string{}
	for k, v := range cfg.Context {
		contextValues[k] = v
	}
	for k, v := range other.Context {
		contextValues[k] = v
	}
	res := &config{}
	for k := range triggersMap {
		res.Triggers = append(res.Triggers, triggersMap[k])
	}
	for k := range templatesMap {
		res.Templates = append(res.Templates, templatesMap[k])
	}
	res.Context = contextValues
	return res
}

func addKubectlFlagsToCmd(cmd *cobra.Command) clientcmd.ClientConfig {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	loadingRules.DefaultClientConfig = &clientcmd.DefaultClientConfig
	overrides := clientcmd.ConfigOverrides{}
	kflags := clientcmd.RecommendedConfigOverrideFlags("")
	cmd.PersistentFlags().StringVar(&loadingRules.ExplicitPath, "kubeconfig", "", "Path to a kube config. Only required if out-of-cluster")
	clientcmd.BindOverrideFlags(&overrides, cmd.PersistentFlags(), kflags)
	return clientcmd.NewInteractiveDeferredLoadingClientConfig(loadingRules, &overrides, os.Stdin)
}
