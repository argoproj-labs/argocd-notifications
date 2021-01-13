package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/argoproj-labs/argocd-notifications/cmd/tools"
)

func main() {
	var command *cobra.Command
	if filepath.Base(os.Args[0]) == "argocd-notifications-backend" || os.Getenv("ARGOCD_NOTIFICATIONS_BACKEND") == "true" {
		command = &cobra.Command{
			Use: "argocd-notifications-backend",
			Run: func(c *cobra.Command, args []string) {
				c.HelpFunc()(c, args)
			},
		}
		command.AddCommand(newControllerCommand())
		command.AddCommand(newBotCommand())
	} else {
		command = tools.NewToolsCommand()
	}

	if err := command.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
