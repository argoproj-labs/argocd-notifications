package tools

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/ghodss/yaml"
	"github.com/spf13/cobra"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"

	"github.com/argoproj-labs/argocd-notifications/shared/cmd"
	"github.com/argoproj-labs/argocd-notifications/shared/settings"
)

func addOutputFlags(cmd *cobra.Command, output *string) {
	cmd.Flags().StringVarP(output, "output", "o", "wide", "Output format. One of:json|yaml|wide|name")
}

func printFormatted(input interface{}, output string, out io.Writer) error {
	switch output {
	case "json":
		data, err := json.MarshalIndent(input, "", "  ")
		if err != nil {
			return err
		}
		_, err = out.Write([]byte(string(data) + "\n"))
		return err
	case "yaml":
		data, err := yaml.Marshal(input)
		if err != nil {
			return err
		}
		_, err = out.Write(data)
		return err
	default:
		return fmt.Errorf("output '%s' is not supported", output)
	}
}

func NewToolsCommand(defaultCfg settings.Config) *cobra.Command {
	var (
		cmdContext = commandContext{
			defaultCfg: defaultCfg,
			stdout:     os.Stdout,
			stderr:     os.Stderr,
		}
	)
	var command = cobra.Command{
		Use: "tools",
		Run: func(c *cobra.Command, args []string) {
			c.HelpFunc()(c, args)
		},
	}

	command.AddCommand(newTriggerCommand(&cmdContext))
	command.AddCommand(newTemplateCommand(&cmdContext))

	command.PersistentFlags().StringVar(&cmdContext.configMapPath,
		"argocd-notification-cm-path", "", "argocd-notification-cm.yaml file path")
	command.PersistentFlags().StringVar(&cmdContext.secretPath, "argocd-notification-secret-path", ":dummy",
		"argocd-notification-secret.yaml file path. Use empty secret if provided value is ':dummy'")
	clientConfig := cmd.AddK8SFlagsToCmd(&command)
	cmdContext.getK8SClients = func() (kubernetes.Interface, dynamic.Interface, string, error) {
		return getK8SClients(clientConfig)
	}
	return &command
}
