package slack

import (
	"errors"
	"net/http"

	"github.com/argoproj-labs/argocd-notifications/shared/settings"

	slackclient "github.com/slack-go/slack"
)

type HasSigningSecret interface {
	GetSigningSecret() string
}

type RequestVerifier func(data []byte, header http.Header) error

func NewVerifier(cfg settings.Config) RequestVerifier {
	return func(data []byte, header http.Header) error {
		signingSecret := ""
		for _, service := range cfg.Notifier.GetServices() {
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
