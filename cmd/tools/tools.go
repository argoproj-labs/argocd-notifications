package tools

import (
	"log"

	"github.com/argoproj-labs/argocd-notifications/shared/argocd"
	"github.com/argoproj-labs/argocd-notifications/shared/k8s"
	"github.com/argoproj-labs/argocd-notifications/shared/settings"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/argoproj/notifications-engine/pkg/cmd"
	"github.com/spf13/cobra"
)

func NewToolsCommand() *cobra.Command {
	var (
		argocdRepoServer string
	)

	var argocdService argocd.Service
	toolsCommand := cmd.NewToolsCommand(
		"argocd-notifications",
		"argocd-notifications",
		k8s.Applications,
		settings.GetFactorySettings(argocdService), func(clientConfig clientcmd.ClientConfig) {
			k8sCfg, err := clientConfig.ClientConfig()
			if err != nil {
				log.Fatalf("Failed to parse k8s config: %v", err)
			}
			ns, _, err := clientConfig.Namespace()
			if err != nil {
				log.Fatalf("Failed to parse k8s config: %v", err)
			}
			argocdService, err = argocd.NewArgoCDService(kubernetes.NewForConfigOrDie(k8sCfg), ns, argocdRepoServer)
			if err != nil {
				log.Fatalf("Failed to initalize Argo CD service: %v", err)
			}
		})
	toolsCommand.PersistentFlags().StringVar(&argocdRepoServer, "argocd-repo-server", "argocd-repo-server:8081", "Argo CD repo server address")
	return toolsCommand
}
