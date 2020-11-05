package tools

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/ghodss/yaml"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"

	"github.com/argoproj-labs/argocd-notifications/shared/cmd"
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

func NewToolsCommand() *cobra.Command {
	var (
		argocdRepoServer string
		cmdContext       = commandContext{
			stdout:        os.Stdout,
			stderr:        os.Stderr,
			argocdService: &lazyArgocdServiceInitializer{argocdRepoServer: &argocdRepoServer},
		}
	)
	var command = cobra.Command{
		Use:   "tools",
		Short: "Set of CLI commands that helps to configure the controller",
		Run: func(c *cobra.Command, args []string) {
			c.HelpFunc()(c, args)
		},
	}

	command.AddCommand(newTriggerCommand(&cmdContext))
	command.AddCommand(newTemplateCommand(&cmdContext))

	command.PersistentFlags().StringVar(&cmdContext.configMapPath,
		"config-map", "", "argocd-notifications-cm.yaml file path")
	command.PersistentFlags().StringVar(&cmdContext.secretPath,
		"secret", "", "argocd-notifications-secret.yaml file path. Use empty secret if provided value is ':empty'")
	command.PersistentFlags().StringVar(&argocdRepoServer,
		"argocd-repo-server", "argocd-repo-server:8081", "Argo CD repo server address")
	clientConfig := cmd.AddK8SFlagsToCmd(&command)
	cmdContext.getK8SClients = func() (kubernetes.Interface, dynamic.Interface, string, error) {
		return getK8SClients(clientConfig)
	}
	cmdContext.argocdService.getK8SClients = cmdContext.getK8SClients
	return &command
}
