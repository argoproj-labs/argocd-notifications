package main

import (
	"fmt"
	"os"
	"path/filepath"

	argocert "github.com/argoproj/argo-cd/util/cert"
	"github.com/spf13/cobra"

	"github.com/argoproj-labs/argocd-notifications/cmd/tools"
	"github.com/argoproj-labs/argocd-notifications/pkg/util/http"
)

func init() {
	// resolve certificates using injected "argocd-tls-certs-cm" ConfigMap
	http.SetCertResolver(argocert.GetCertificateForConnect)
}

func main() {
	binaryName := filepath.Base(os.Args[0])
	if val := os.Getenv("ARGOCD_NOTIFICATIONS_BINARY"); val != "" {
		binaryName = val
	}
	var command *cobra.Command
	switch binaryName {
	case "argocd-notifications-backend":
		command = &cobra.Command{
			Use: "argocd-notifications-backend",
			Run: func(c *cobra.Command, args []string) {
				c.HelpFunc()(c, args)
			},
		}
		command.AddCommand(newControllerCommand())
		command.AddCommand(newBotCommand())
	default:
		command = tools.NewToolsCommand()
	}

	if err := command.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
