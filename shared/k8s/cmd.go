package k8s

import (
	"context"
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/leaderelection"
	"k8s.io/client-go/tools/leaderelection/resourcelock"
	componentbaseconfig "k8s.io/component-base/config"
)

func AddK8SFlagsToCmd(cmd *cobra.Command) clientcmd.ClientConfig {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	loadingRules.DefaultClientConfig = &clientcmd.DefaultClientConfig
	overrides := clientcmd.ConfigOverrides{}
	kflags := clientcmd.RecommendedConfigOverrideFlags("")
	cmd.PersistentFlags().StringVar(&loadingRules.ExplicitPath, "kubeconfig", "", "Path to a kube config. Only required if out-of-cluster")
	clientcmd.BindOverrideFlags(&overrides, cmd.PersistentFlags(), kflags)
	return clientcmd.NewInteractiveDeferredLoadingClientConfig(loadingRules, &overrides, os.Stdin)
}

func CreateLeaderElector(
	lockName string,
	namespace string,
	id string,
	k8sClient *kubernetes.Clientset,
	configuration *componentbaseconfig.LeaderElectionConfiguration,
	startLeadingHandler func(ctx context.Context)) (*leaderelection.LeaderElector, error) {
	return leaderelection.NewLeaderElector(leaderelection.LeaderElectionConfig{
		Lock: &resourcelock.LeaseLock{
			LeaseMeta: metav1.ObjectMeta{
				Name:      lockName,
				Namespace: namespace,
			},
			Client: k8sClient.CoordinationV1(),
			LockConfig: resourcelock.ResourceLockConfig{
				Identity: id,
			},
		},
		ReleaseOnCancel: true,
		LeaseDuration:   configuration.LeaseDuration.Duration,
		RenewDeadline:   configuration.RenewDeadline.Duration,
		RetryPeriod:     configuration.RetryPeriod.Duration,
		Callbacks: leaderelection.LeaderCallbacks{
			OnStartedLeading: startLeadingHandler,
			OnStoppedLeading: func() {
				log.Infof("leader lost: %s", id)
				os.Exit(0)
			},
			OnNewLeader: func(identity string) {
				if identity == id {
					// I'm new leader
					return
				}
				log.Infof("new leader is elected: %s", identity)
			},
		},
	})
}
