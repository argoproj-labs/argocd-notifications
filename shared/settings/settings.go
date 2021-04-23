package settings

import (
	"context"
	"errors"
	"strings"
	"sync"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/argoproj-labs/argocd-notifications/shared/argocd"
	"github.com/argoproj-labs/argocd-notifications/shared/k8s"

	"github.com/argoproj/notifications-engine/pkg"
	"github.com/argoproj/notifications-engine/pkg/services"
	"github.com/ghodss/yaml"
	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

const (
	partOfLabel = "app.kubernetes.io/part-of"
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

func isConfig(obj metav1.Object, name string) bool {
	if obj.GetName() == name {
		return true
	}
	if obj.GetLabels() == nil {
		return false
	}
	return obj.GetLabels()[partOfLabel] == "argocd-notifications"
}

func WatchConfig(
	ctx context.Context,
	argocdService argocd.Service,
	clientset kubernetes.Interface,
	namespace string,
	callback func(Config) error, opts ...CfgOpts,
) error {
	cmInformer := k8s.NewConfigMapInformer(clientset, namespace)
	secretInformer := k8s.NewSecretInformer(clientset, namespace)

	lock := &sync.Mutex{}
	onConfigChanged := func() {
		lock.Lock()
		defer lock.Unlock()
		configMap := v1.ConfigMap{Data: map[string]string{}}

		for _, obj := range cmInformer.GetStore().List() {
			cm, ok := obj.(*v1.ConfigMap)
			if !ok || !isConfig(cm, "argocd-notifications-cm") {
				continue
			}
			for k, v := range cm.Data {
				if _, has := configMap.Data[k]; has {
					log.Warnf("Key '%s' of ConfigMap '%s' ignored because it is a duplicate", k, cm.Name)
				} else {
					configMap.Data[k] = v
				}
			}
		}
		secret := v1.Secret{Data: map[string][]byte{}}
		for _, obj := range secretInformer.GetStore().List() {
			s, ok := obj.(*v1.Secret)
			if !ok || !isConfig(s, "argocd-notifications-secret") {
				continue
			}
			for k, v := range s.Data {
				if _, has := secret.Data[k]; has {
					log.Warnf("Key '%s' of Secret '%s' ignored because it is a duplicate", k, s.Name)
				} else {
					secret.Data[k] = v
				}
			}
		}

		if cfg, err := NewConfig(&configMap, &secret, argocdService, opts...); err == nil {
			if err = callback(*cfg); err != nil {
				log.Warnf("Failed to apply new settings: %v", err)
			}
		} else {
			log.Warnf("Failed to parse new settings: %v", err)
		}
	}

	cmInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		UpdateFunc: func(oldObj, newObj interface{}) {
			onConfigChanged()
		},
	})

	secretInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		UpdateFunc: func(oldObj, newObj interface{}) {
			onConfigChanged()
		},
	})
	go secretInformer.Run(ctx.Done())
	go cmInformer.Run(ctx.Done())

	if !cache.WaitForCacheSync(ctx.Done(), cmInformer.HasSynced, secretInformer.HasSynced) {
		return errors.New("timed out waiting for caches to sync")
	}
	go onConfigChanged()
	return nil
}
