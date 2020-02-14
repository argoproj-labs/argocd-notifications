package main

import (
	"fmt"
	"os"

	"k8s.io/client-go/tools/clientcmd"

	"github.com/spf13/cobra"
)

func main() {
	var command = &cobra.Command{
		Use:   "argocd-notifications",
		Short: "argocd controls a Argo CD server",
		RunE: func(c *cobra.Command, args []string) error {
			// run controller command by default
			os.Args = append([]string{os.Args[0], "controller"}, os.Args[1:]...)
			return c.Execute()
		},
	}
	command.AddCommand(newControllerCommand())
	command.AddCommand(newBotCommand())
	if err := command.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
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
