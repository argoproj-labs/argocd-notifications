package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/argoproj-labs/argocd-notifications/controller"
	"github.com/argoproj-labs/argocd-notifications/notifiers"
	"github.com/argoproj-labs/argocd-notifications/triggers"

	"github.com/ghodss/yaml"
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
		clientConfig        clientcmd.ClientConfig
		processorsCount     int
		namespace           string
		configPath          string
		notifiersConfigPath string
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

			notifiersData, err := ioutil.ReadFile(notifiersConfigPath)
			if err != nil {
				return err
			}
			notifiersConfig := notifiers.Config{}
			err = yaml.Unmarshal(notifiersData, &notifiersConfig)
			if err != nil {
				return err
			}

			config, err := getConfig(configPath)
			if err != nil {
				return err
			}

			t, err := triggers.GetTriggers(config.Templates, config.Triggers)
			if err != nil {
				return err
			}

			ctrl, err := controller.NewController(client, namespace, t, notifiers.GetAll(notifiersConfig), config.Context)
			if err != nil {
				return err
			}
			return ctrl.Run(context.Background(), processorsCount)
		},
	}
	clientConfig = addKubectlFlagsToCmd(&command)
	command.Flags().IntVar(&processorsCount, "processors-count", 3, "Processors count.")
	command.Flags().StringVar(&namespace, "namespace", "", "Namespace which controller handles. Current namespace if empty.")
	command.Flags().StringVar(&configPath, "config", "", "Configuration file location")
	command.Flags().StringVar(&notifiersConfigPath, "notifiers", "./notifiers.yaml", "Notifiers config file location")

	return &command
}

func getConfig(configPath string) (*controller.Config, error) {
	config := controller.Config{}
	defaultNotifiersData, err := ioutil.ReadFile("./assets/config.yaml")
	if err != nil {
		return nil, err
	}
	err = yaml.Unmarshal(defaultNotifiersData, &config)
	if err != nil {
		return nil, err
	}

	if configPath != "" {
		configData, err := ioutil.ReadFile(configPath)
		if err != nil {
			return nil, err
		}
		userConfig := controller.Config{}

		err = yaml.Unmarshal(configData, &userConfig)
		if err != nil {
			return nil, err
		}
		config.Triggers = append(config.Triggers, userConfig.Triggers...)
		config.Templates = append(config.Templates, userConfig.Templates...)
		for k, v := range userConfig.Context {
			config.Context[k] = v
		}
	}

	return &config, err
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
