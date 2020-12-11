package slack

import (
	"errors"
	"fmt"
	"net/http"

	slackclient "github.com/slack-go/slack"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/cache"

	"github.com/argoproj-labs/argocd-notifications/shared/settings"
)

type HasSigningSecret interface {
	GetSigningSecret() string
}

type RequestVerifier func(data []byte, header http.Header) error

func NewVerifier(secretInformer cache.SharedIndexInformer) RequestVerifier {
	return func(data []byte, header http.Header) error {
		secrets := secretInformer.GetStore().List()
		if len(secrets) == 0 {
			return fmt.Errorf("cannot find secret %s the slack app secret", settings.SecretName)
		}
		secret, ok := secrets[0].(*v1.Secret)
		if !ok {
			return errors.New("unexpected object in the secret informer storage")
		}
		notifiers, err := settings.ParseSecret(secret)
		if err != nil {
			return errors.New("unable to parse slack configuration")
		}
		signingSecret := ""
		for _, notifier := range notifiers {
			if hasSecret, ok := notifier.(HasSigningSecret); ok {
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
