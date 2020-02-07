package main

import (
	"github.com/argoproj-labs/argocd-notifications/bot"
	"github.com/spf13/cobra"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/tools/clientcmd"
)

func newBotCommand() *cobra.Command {
	var (
		clientConfig clientcmd.ClientConfig
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
			server := bot.NewServer(dynamicClient)
			return server.Serve(port)
		},
	}
	clientConfig = addKubectlFlagsToCmd(&command)
	command.Flags().IntVar(&port, "port", 8080, "Port number.")
	return &command
}
