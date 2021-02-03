package settings

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/ghodss/yaml"
	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"

	"github.com/argoproj-labs/argocd-notifications/pkg"
	"github.com/argoproj-labs/argocd-notifications/pkg/services"
	"github.com/argoproj-labs/argocd-notifications/shared/argocd"
	"github.com/argoproj-labs/argocd-notifications/shared/k8s"
)

type Config struct {
	pkg.Config

	// Context holds list of configured key value pairs available in notification templates
	Context map[string]string
	// Subscriptions holds list of default application subscriptions
	Subscriptions DefaultSubscriptions
	// DefaultTriggers holds list of triggers that is used by default if subscriber don't specify trigger
	DefaultTriggers []string
	// ArgoCDService encapsulates methods provided by Argo CD
	ArgoCDService argocd.Service
	// API allows sending notifications
	API pkg.API
}

// Returns list of recipients for the specified trigger
func (cfg Config) GetGlobalSubscriptions(labels map[string]string) pkg.Subscriptions {
	subscriptions := pkg.Subscriptions{}
	for _, s := range cfg.Subscriptions {
		triggers := s.Triggers
		if len(triggers) == 0 {
			triggers = cfg.DefaultTriggers
		}
		for _, trigger := range triggers {
			if s.MatchesTrigger(trigger) && s.Selector.Matches(fields.Set(labels)) {
				for _, recipient := range s.Recipients {
					parts := strings.Split(recipient, ":")
					dest := services.Destination{Service: parts[0]}
					if len(parts) > 1 {
						dest.Recipient = parts[1]
					}
					subscriptions[trigger] = append(subscriptions[trigger], dest)
				}
			}
		}
	}
	return subscriptions
}

type CfgOpts = func(*Config, *v1.ConfigMap, *v1.Secret) error

// NewConfig retrieves configured templates and triggers from the provided config map
func NewConfig(configMap *v1.ConfigMap, secret *v1.Secret, argocdService argocd.Service, opts ...CfgOpts) (*Config, error) {
	// read all the keys in format of templates.%s and triggers.%s
	// to create config
	c, err := pkg.ParseConfig(configMap, secret)
	if err != nil {
		return nil, err
	}
	cfg := Config{
		Config: *c,
		Context: map[string]string{
			"argocdUrl": "https://localhost:4000",
		},
		ArgoCDService: argocdService,
	}

	if subscriptionYaml, ok := configMap.Data["subscriptions"]; ok {
		if err := yaml.Unmarshal([]byte(subscriptionYaml), &cfg.Subscriptions); err != nil {
			return nil, err
		}
	}

	if contextYaml, ok := configMap.Data["context"]; ok {
		if err := yaml.Unmarshal([]byte(contextYaml), &cfg.Context); err != nil {
			return nil, err
		}
	}

	if defaultTriggersYaml, ok := configMap.Data["defaultTriggers"]; ok {
		if err := yaml.Unmarshal([]byte(defaultTriggersYaml), &cfg.DefaultTriggers); err != nil {
			return nil, err
		}
	}

	for _, fn := range opts {
		if err := fn(&cfg, configMap, secret); err != nil {
			return nil, err
		}
	}

	if cfg.API, err = pkg.NewAPI(*c); err != nil {
		return nil, err
	} else {
		return &cfg, nil
	}
}

func WatchConfig(
	ctx context.Context,
	argocdService argocd.Service,
	clientset kubernetes.Interface,
	namespace string,
	callback func(Config) error, opts ...CfgOpts,
) error {
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
			if cfg, err := NewConfig(configMap, secret, argocdService, opts...); err == nil {
				if err = callback(*cfg); err != nil {
					log.Warnf("Failed to apply new settings: %v", err)
				}
			} else {
				log.Warnf("Failed to parse new settings: %v", err)
			}
		}
	}

	onConfigMapChanged := func(newObj interface{}) {
		if cm, ok := newObj.(*v1.ConfigMap); ok {
			onNewConfigMapAndSecret(nil, cm)
		}
	}

	onSecretChanged := func(newObj interface{}) {
		if s, ok := newObj.(*v1.Secret); ok {
			onNewConfigMapAndSecret(s, nil)
		}
	}

	cmInformer := k8s.NewConfigMapInformer(clientset, namespace)
	cmInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		UpdateFunc: func(oldObj, newObj interface{}) {
			onConfigMapChanged(newObj)
		},
		AddFunc: func(obj interface{}) {
			log.Info("config map found")
			onConfigMapChanged(obj)
		},
	})

	secretInformer := k8s.NewSecretInformer(clientset, namespace)
	secretInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		UpdateFunc: func(oldObj, newObj interface{}) {
			onSecretChanged(newObj)
		},
		AddFunc: func(obj interface{}) {
			log.Info("secret found")
			onSecretChanged(obj)
		},
	})
	go secretInformer.Run(ctx.Done())
	go cmInformer.Run(ctx.Done())

	if !cache.WaitForCacheSync(ctx.Done(), cmInformer.HasSynced, secretInformer.HasSynced) {
		return errors.New("timed out waiting for caches to sync")
	}
	var missingWarn []string
	if len(cmInformer.GetStore().List()) == 0 {
		missingWarn = append(missingWarn, fmt.Sprintf("config map %s", k8s.ConfigMapName))
	}
	if len(secretInformer.GetStore().List()) == 0 {
		missingWarn = append(missingWarn, fmt.Sprintf("secret %s", k8s.SecretName))
	}
	if len(missingWarn) > 0 {
		log.Warnf("Cannot find %s. Waiting when both config map and secret are created.", strings.Join(missingWarn, " and "))
	}
	return nil
}
