package slack

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/argoproj-labs/argocd-notifications/pkg"

	slackclient "github.com/slack-go/slack"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/cache"

	"github.com/argoproj-labs/argocd-notifications/shared/k8s"
)

type HasSigningSecret interface {
	GetSigningSecret() string
}

type RequestVerifier func(data []byte, header http.Header) error

func NewVerifier(cmInformer cache.SharedIndexInformer, secretInformer cache.SharedIndexInformer) RequestVerifier {
	return func(data []byte, header http.Header) error {
		secrets := secretInformer.GetStore().List()
		if len(secrets) == 0 {
			return fmt.Errorf("cannot find secret %s the slack app secret", k8s.SecretName)
		}
		secret, ok := secrets[0].(*v1.Secret)
		if !ok {
			return errors.New("unexpected object in the secret informer storage")
		}
		configMaps := cmInformer.GetStore().List()
		if len(configMaps) == 0 {
			return fmt.Errorf("cannot find config map %s the slack app secret", k8s.ConfigMapName)
		}
		cm, ok := configMaps[0].(*v1.ConfigMap)
		if !ok {
			return errors.New("unexpected object in the configmap informer storage")
		}
		cfg, err := pkg.ParseConfig(cm, secret)
		if err != nil {
			return fmt.Errorf("unable to parse slack configuration: %v", err)
		}
		notifier, err := pkg.NewNotifier(*cfg)
		if err != nil {
			return fmt.Errorf("unable to parse slack configuration: %v", err)
		}
		signingSecret := ""
		for _, service := range notifier.GetServices() {
			if hasSecret, ok := service.(HasSigningSecret); ok {
				signingSecret = hasSecret.GetSigningSecret()
				if signingSecret == "" {
					return errors.New("slack signing secret is not configured")
				}
			}
		}

		if signingSecret == "" {
			return errors.New("slack is not configured")
		}

		verifier, err := slackclient.NewSecretsVerifier(header, signingSecret)
		if err != nil {
			return err
		}
		_, err = verifier.Write(data)
		if err != nil {
			return err
		}
		return verifier.Ensure()
	}
}
