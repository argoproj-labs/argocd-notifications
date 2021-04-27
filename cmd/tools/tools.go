package tools

import (
	"github.com/argoproj-labs/argocd-notifications/expr"
	"github.com/argoproj-labs/argocd-notifications/shared/argocd"
	"github.com/argoproj-labs/argocd-notifications/shared/k8s"
	"github.com/argoproj-labs/argocd-notifications/shared/legacy"

	"github.com/argoproj/notifications-engine/pkg/cmd"
	"github.com/argoproj/notifications-engine/pkg/services"
	"github.com/ghodss/yaml"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func NewToolsCommand() *cobra.Command {
	var (
		argocdRepoServer string
	)

	toolsCommand := cmd.NewToolsCommand("argocd-notifications", cmd.Config{
		Resource:      k8s.Applications,
		SecretName:    k8s.SecretName,
		ConfigMapName: k8s.ConfigMapName,
		CLIName:       "argocd-notifications",
		CreateVars: func(obj map[string]interface{}, dest services.Destination, cmdContext cmd.CommandContext) (map[string]interface{}, error) {
			k8sClient, _, ns, err := cmdContext.GetK8SClients()
			if err != nil {
				return nil, err
			}
			argocdService, err := argocd.NewArgoCDService(k8sClient, ns, argocdRepoServer)
			if err != nil {
				return nil, err
			}
			configMap, err := cmdContext.GetConfigMap()
			if err != nil {
				return nil, err
			}
			context := map[string]string{}
			if contextYaml, ok := configMap.Data["context"]; ok {
				if err := yaml.Unmarshal([]byte(contextYaml), &context); err != nil {
					return nil, err
				}
			}

			return expr.Spawn(&unstructured.Unstructured{Object: obj}, argocdService, map[string]interface{}{
				"app":     obj,
				"context": legacy.InjectLegacyVar(context, dest.Service),
			}), nil
		},
	})
	toolsCommand.PersistentFlags().StringVar(&argocdRepoServer, "argocd-repo-server", "argocd-repo-server:8081", "Argo CD repo server address")
	return toolsCommand

}
