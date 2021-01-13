package main

import (
	"context"
	"log"
	"net/http"
	"sync"

	"github.com/spf13/cobra"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/argoproj-labs/argocd-notifications/bot"
	"github.com/argoproj-labs/argocd-notifications/bot/slack"
	"github.com/argoproj-labs/argocd-notifications/shared/k8s"
	"github.com/argoproj-labs/argocd-notifications/shared/legacy"
	"github.com/argoproj-labs/argocd-notifications/shared/settings"
)

func newBotCommand() *cobra.Command {
	var (
		clientConfig clientcmd.ClientConfig
		namespace    string
		port         int
	)
	var command = cobra.Command{
		Use:   "bot",
		Short: "Starts Argo CD Notifications bot",
		RunE: func(c *cobra.Command, args []string) error {
			restConfig, err := clientConfig.ClientConfig()
			if err != nil {
				return err
			}
			dynamicClient, err := dynamic.NewForConfig(restConfig)
			if err != nil {
				return err
			}
			clientset, err := kubernetes.NewForConfig(restConfig)
			if err != nil {
				return err
			}
			if namespace == "" {
				namespace, _, err = clientConfig.Namespace()
				if err != nil {
					return err
				}
			}
			cfgSrc := make(chan settings.Config)
			if err = settings.WatchConfig(context.Background(), nil, clientset, namespace, func(config settings.Config) error {
				cfgSrc <- config
				return nil
			}, legacy.ApplyLegacyConfig); err != nil {
				log.Fatal(err)
			}
			server := bot.NewServer(dynamicClient, namespace)
			server.AddAdapter("/slack", slack.NewSlackAdapter(getVerifier(cfgSrc)))
			return server.Serve(port)
		},
	}
	clientConfig = k8s.AddK8SFlagsToCmd(&command)
	command.Flags().IntVar(&port, "port", 8080, "Port number.")
	command.Flags().StringVar(&namespace, "namespace", "", "Namespace which bot handles. Current namespace if empty.")
	return &command
}

func getVerifier(cfgSrc chan settings.Config) slack.RequestVerifier {
	cfg := <-cfgSrc
	verifier := slack.NewVerifier(cfg)

	var lock sync.Mutex

	go func() {
		for next := range cfgSrc {
			lock.Lock()
			verifier = slack.NewVerifier(next)
			lock.Unlock()
		}
	}()

	return func(data []byte, header http.Header) (string, error) {
		var currentVerifier slack.RequestVerifier
		lock.Lock()
		currentVerifier = verifier
		lock.Unlock()
		return currentVerifier(data, header)
	}
}
