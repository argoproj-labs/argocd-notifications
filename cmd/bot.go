package main

import (
	"fmt"

	"github.com/argoproj/notifications-engine/pkg/api"

	"github.com/spf13/cobra"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/argoproj-labs/argocd-notifications/bot"
	"github.com/argoproj-labs/argocd-notifications/bot/slack"
	"github.com/argoproj-labs/argocd-notifications/shared/k8s"
	"github.com/argoproj-labs/argocd-notifications/shared/settings"
)

func newBotCommand() *cobra.Command {
	var (
		clientConfig clientcmd.ClientConfig
		namespace    string
		port         int
		slackPath    string
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

			apiFactory := api.NewFactory(settings.GetFactorySettings(nil),
				namespace,
				k8s.NewSecretInformer(clientset, namespace), k8s.NewConfigMapInformer(clientset, namespace))

			server := bot.NewServer(dynamicClient, namespace)
			server.AddAdapter(fmt.Sprintf("/%s", slackPath), slack.NewSlackAdapter(slack.NewVerifier(apiFactory)))
			return server.Serve(port)
		},
	}
	clientConfig = k8s.AddK8SFlagsToCmd(&command)
	command.Flags().IntVar(&port, "port", 8080, "Port number.")
	command.Flags().StringVar(&namespace, "namespace", "", "Namespace which bot handles. Current namespace if empty.")
	command.Flags().StringVar(&slackPath, "slack-path", "slack", "Path to the slack bot handler")
	return &command
}
