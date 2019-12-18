package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"
	"time"

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
	argocdURLContextVariable = "argocdURL"
	triggersConfigMapKey     = "triggers"
	templatesConfigMapKey    = "templates"
	contextConfigMapKey      = "context"
)

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
			watchSettings(k8sClient, namespace, func(triggers map[string]triggers.Trigger, notifiers map[string]notifiers.Notifier, contextVals map[string]string) error {
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
	notificationTemplates := make([]triggers.NotificationTemplate, 0)
	if data, ok := configData[templatesConfigMapKey]; ok {
		err := yaml.Unmarshal([]byte(data), &notificationTemplates)
		if err != nil {
			return nil, nil, nil, err
		}
	}
	notificationTriggers := make([]triggers.NotificationTrigger, 0)
	if data, ok := configData[triggersConfigMapKey]; ok {
		err := yaml.Unmarshal([]byte(data), &notificationTriggers)
		if err != nil {
			return nil, nil, nil, err
		}
	}
	notifiersConfig := notifiers.Config{}
	err := yaml.Unmarshal(notifiersData, &notifiersConfig)
	if err != nil {
		return nil, nil, nil, err
	}
	contextValues := make(map[string]string)
	if data, ok := configData[contextConfigMapKey]; ok {
		err := yaml.Unmarshal([]byte(data), &contextValues)
		if err != nil {
			return nil, nil, nil, err
		}
	}

	if _, ok := contextValues[argocdURLContextVariable]; !ok {
		contextValues[argocdURLContextVariable] = "https://localhost:4000"
	}
	t, err := triggers.GetTriggers(notificationTemplates, notificationTriggers)
	if err != nil {
		return nil, nil, nil, err
	}
	return t, notifiers.GetAll(notifiersConfig), contextValues, nil
}

func watchSettings(clientset kubernetes.Interface, namespace string, callback func(map[string]triggers.Trigger, map[string]notifiers.Notifier, map[string]string) error) {
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

func addKubectlFlagsToCmd(cmd *cobra.Command) clientcmd.ClientConfig {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	loadingRules.DefaultClientConfig = &clientcmd.DefaultClientConfig
	overrides := clientcmd.ConfigOverrides{}
	kflags := clientcmd.RecommendedConfigOverrideFlags("")
	cmd.PersistentFlags().StringVar(&loadingRules.ExplicitPath, "kubeconfig", "", "Path to a kube config. Only required if out-of-cluster")
	clientcmd.BindOverrideFlags(&overrides, cmd.PersistentFlags(), kflags)
	return clientcmd.NewInteractiveDeferredLoadingClientConfig(loadingRules, &overrides, os.Stdin)
}
