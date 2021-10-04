package cmd

import (
	"fmt"
	"os"

	"github.com/argoproj/notifications-engine/pkg/api"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func withDebugLogs() func() {
	level := log.GetLevel()
	log.SetLevel(log.DebugLevel)
	return func() {
		log.SetLevel(level)
	}
}

func addOutputFlags(cmd *cobra.Command, output *string) {
	cmd.Flags().StringVarP(output, "output", "o", "wide", "Output format. One of:json|yaml|wide|name")
}

func NewToolsCommand(name string, cliName string, resource schema.GroupVersionResource, settings api.Settings, opts ...func(cfg clientcmd.ClientConfig)) *cobra.Command {
	var (
		cmdContext = commandContext{
			Settings: settings,
			resource: resource,
			stdout:   os.Stdout,
			stderr:   os.Stderr,
			stdin:    os.Stdin,
			cliName:  cliName,
		}
	)
	var command = cobra.Command{
		Use:   name,
		Short: "Set of CLI commands that helps manage notifications settings",
		Run: func(c *cobra.Command, args []string) {
			c.HelpFunc()(c, args)
		},
	}

	command.AddCommand(newTriggerCommand(&cmdContext))
	command.AddCommand(newTemplateCommand(&cmdContext))

	command.PersistentFlags().StringVar(&cmdContext.configMapPath,
		"config-map", "", fmt.Sprintf("%s.yaml file path", settings.ConfigMapName))
	command.PersistentFlags().StringVar(&cmdContext.secretPath,
		"secret", "", fmt.Sprintf("%s.yaml file path. Use empty secret if provided value is ':empty'", settings.SecretName))
	clientConfig := addK8SFlagsToCmd(&command)
	command.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		for i := range opts {
			opts[i](clientConfig)
		}
		ns, _, err := clientConfig.Namespace()
		if err != nil {
			log.Fatalf("Failed to extract namespace: %v", err)
		}
		cmdContext.namespace = ns
		config, err := clientConfig.ClientConfig()
		if err != nil {
			log.Fatalf("Failed to create k8s client: %v", err)
		}

		cmdContext.dynamicClient = dynamic.NewForConfigOrDie(config)
		cmdContext.k8sClient = kubernetes.NewForConfigOrDie(config)
	}
	return &command
}

func addK8SFlagsToCmd(cmd *cobra.Command) clientcmd.ClientConfig {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	loadingRules.DefaultClientConfig = &clientcmd.DefaultClientConfig
	overrides := clientcmd.ConfigOverrides{}
	kflags := clientcmd.RecommendedConfigOverrideFlags("")
	cmd.PersistentFlags().StringVar(&loadingRules.ExplicitPath, "kubeconfig", "", "Path to a kube config. Only required if out-of-cluster")
	clientcmd.BindOverrideFlags(&overrides, cmd.PersistentFlags(), kflags)
	return clientcmd.NewInteractiveDeferredLoadingClientConfig(loadingRules, &overrides, os.Stdin)
}
