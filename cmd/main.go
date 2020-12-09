package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/argoproj-labs/argocd-notifications/cmd/tools"
)

func main() {
	var command = &cobra.Command{
		Use:   "argocd-notifications",
		Short: "argocd-notifications notifies about Argo CD application changes",
		Run: func(c *cobra.Command, args []string) {
			c.HelpFunc()(c, args)
		},
	}
	command.AddCommand(newControllerCommand())
	command.AddCommand(newBotCommand())
	command.AddCommand(tools.NewToolsCommand())
	if err := command.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
