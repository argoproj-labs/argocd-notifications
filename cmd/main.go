package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"

	"gopkg.in/yaml.v2"

	"github.com/argoproj-labs/argocd-notifications/controller"
	"github.com/spf13/cobra"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	if err := newCommand().Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func newCommand() *cobra.Command {
	var (
		clientConfig    clientcmd.ClientConfig
		processorsCount int
		namespace       string
		configPath      string
	)
	var command = cobra.Command{
		Use: "argocd-notifications",
		RunE: func(c *cobra.Command, args []string) error {
			restConfig, err := clientConfig.ClientConfig()
			if err != nil {
				return err
			}
			client, err := dynamic.NewForConfig(restConfig)
			if err != nil {
				return err
			}
			if namespace == "" {
				namespace, _, err = clientConfig.Namespace()
				if err != nil {
					return err
				}
			}

			configData, err := ioutil.ReadFile(configPath)
			if err != nil {
				return err
			}
			config := controller.Config{}
			err = yaml.Unmarshal(configData, &config)
			if err != nil {
				return err
			}
			ctrl := controller.NewController(client, namespace, config)
			return ctrl.Run(context.Background(), processorsCount)
		},
	}
	clientConfig = addKubectlFlagsToCmd(&command)
	command.Flags().IntVar(&processorsCount, "processors-count", 3, "Processors count.")
	command.Flags().StringVar(&namespace, "namespace", "", "Namespace which controller handles. Current namespace if empty.")
	command.Flags().StringVar(&configPath, "config", "./config.yaml", "Configuration file location")

	return &command
}

func addKubectlFlagsToCmd(cmd *cobra.Command) clientcmd.ClientConfig {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	loadingRules.DefaultClientConfig = &clientcmd.DefaultClientConfig
	overrides := clientcmd.ConfigOverrides{}
	kflags := clientcmd.RecommendedConfigOverrideFlags("")
	cmd.PersistentFlags().StringVar(&loadingRules.ExplicitPath, "kubeconfig", "", "Path to a kube config. Only required if out-of-cluster")
	clientcmd.BindOverrideFlags(&overrides, cmd.PersistentFlags(), kflags)
	return clientcmd.NewInteractiveDeferredLoadingClientConfig(loadingRules, &overrides, os.Stdin)
}
