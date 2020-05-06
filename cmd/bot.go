package main

import (
	"context"
	"log"

	"github.com/spf13/cobra"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/argoproj-labs/argocd-notifications/bot"
	"github.com/argoproj-labs/argocd-notifications/bot/slack"
	"github.com/argoproj-labs/argocd-notifications/shared/cmd"
	"github.com/argoproj-labs/argocd-notifications/shared/settings"
)

func newBotCommand() *cobra.Command {
	var (
		clientConfig clientcmd.ClientConfig
		namespace    string
		port         int
	)
	var command = cobra.Command{
		Use: "bot",
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
			secretInformer := settings.NewSecretInformer(clientset, namespace)
			go secretInformer.Run(context.Background().Done())
			if !cache.WaitForCacheSync(context.Background().Done(), secretInformer.HasSynced) {
				log.Fatal("Timed out waiting for caches to sync")
			}
			server := bot.NewServer(dynamicClient, namespace)
			server.AddAdapter("/slack", slack.NewSlackAdapter(slack.NewVerifier(secretInformer)))
			return server.Serve(port)
		},
	}
	clientConfig = cmd.AddK8SFlagsToCmd(&command)
	command.Flags().IntVar(&port, "port", 8080, "Port number.")
	command.Flags().StringVar(&namespace, "namespace", "", "Namespace which bot handles. Current namespace if empty.")
	return &command
}
