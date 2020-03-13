package main

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/argoproj-labs/argocd-notifications/builtin"
	"github.com/argoproj-labs/argocd-notifications/controller"
	"github.com/argoproj-labs/argocd-notifications/notifiers"
	"github.com/argoproj-labs/argocd-notifications/shared/settings"
	"github.com/argoproj-labs/argocd-notifications/triggers"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	argocdURLContextVariable = "argocdUrl"
)

var (
	defaultCfg settings.Config
)

func init() {
	defaultCfg = settings.Config{
		Templates: builtin.Templates,
		Triggers:  builtin.Triggers,
		Context:   map[string]string{argocdURLContextVariable: "https://localhost:4000"},
	}
}

func newControllerCommand() *cobra.Command {
	var (
		clientConfig     clientcmd.ClientConfig
		processorsCount  int
		namespace        string
		appLabelSelector string
		logLevel         string
	)
	var command = cobra.Command{
		Use: "controller",
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
			level, err := log.ParseLevel(logLevel)
			if err != nil {
				return err
			}
			log.SetLevel(level)

			var cancelPrev context.CancelFunc
			watchConfig(context.Background(), k8sClient, namespace, func(triggers map[string]triggers.Trigger, notifiers map[string]notifiers.Notifier, contextVals map[string]string) error {
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
	command.Flags().StringVar(&logLevel, "loglevel", "info", "Set the logging level. One of: debug|info|warn|error")
	return &command
}

func parseConfig(configMap *v1.ConfigMap, secret *v1.Secret) (map[string]triggers.Trigger, map[string]notifiers.Notifier, map[string]string, error) {
	cfg, err := settings.ParseConfigMap(configMap)
	if err != nil {
		return nil, nil, nil, err
	}
	cfg = defaultCfg.Merge(cfg)
	t, err := triggers.GetTriggers(cfg.Templates, cfg.Triggers)
	if err != nil {
		return nil, nil, nil, err
	}
	notifiersConfig, err := settings.ParseSecret(secret)
	if err != nil {
		return nil, nil, nil, err
	}
	return t, notifiers.GetAll(notifiersConfig), cfg.Context, nil
}

func watchConfig(ctx context.Context, clientset kubernetes.Interface, namespace string, callback func(map[string]triggers.Trigger, map[string]notifiers.Notifier, map[string]string) error) {
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
			if t, n, c, err := parseConfig(configMap, secret); err == nil {
				if err = callback(t, n, c); err != nil {
					log.Fatalf("Failed to start controller: %v", err)
				}
			} else {
				log.Fatalf("Failed to parse new settings: %v", err)
			}
		}
	}

	onConfigMapChanged := func(newObj interface{}) {
		if cm, ok := newObj.(*v1.ConfigMap); ok {
			onNewConfigMapAndSecret(nil, cm)
		}
	}

	cmInformer := settings.NewConfigMapInformer(clientset, namespace)
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

	secretInformer := settings.NewSecretInformer(clientset, namespace)
	secretInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: onSecretChanged,
		UpdateFunc: func(oldObj, newObj interface{}) {
			onSecretChanged(newObj)
		},
	})
	go secretInformer.Run(ctx.Done())
	go cmInformer.Run(ctx.Done())

	if !cache.WaitForCacheSync(ctx.Done(), cmInformer.HasSynced, secretInformer.HasSynced) {
		log.Fatal(errors.New("timed out waiting for caches to sync"))
	}
	var missingWarn []string
	if len(cmInformer.GetStore().List()) == 0 {
		missingWarn = append(missingWarn, fmt.Sprintf("config map %s", settings.ConfigMapName))
	}
	if len(secretInformer.GetStore().List()) == 0 {
		missingWarn = append(missingWarn, fmt.Sprintf("secret %s", settings.SecretName))
	}
	if len(missingWarn) > 0 {
		log.Warnf("Cannot find %s. Waiting when both config map and secret are created.", strings.Join(missingWarn, " and "))
	}
}
