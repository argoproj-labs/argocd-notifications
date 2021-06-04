package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/argoproj-labs/argocd-notifications/controller"
	"github.com/argoproj-labs/argocd-notifications/shared/argocd"
	"github.com/argoproj-labs/argocd-notifications/shared/k8s"

	notificationscontroller "github.com/argoproj/notifications-engine/pkg/controller"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	defaultMetricsPort = 9001
)

func newControllerCommand() *cobra.Command {
	var (
		clientConfig              clientcmd.ClientConfig
		processorsCount           int
		namespace                 string
		appLabelSelector          string
		logLevel                  string
		logFormat                 string
		metricsPort               int
		argocdRepoServer          string
		argocdRepoServerPlaintext bool
		argocdRepoServerStrictTLS bool
		configMapName             string
		secretName                string
	)
	var command = cobra.Command{
		Use:   "controller",
		Short: "Starts Argo CD Notifications controller",
		RunE: func(c *cobra.Command, args []string) error {
			restConfig, err := clientConfig.ClientConfig()
			if err != nil {
				return err
			}
			dynamicClient, err := dynamic.NewForConfig(restConfig)
			if err != nil {
				return err
			}
			k8sClient, err := kubernetes.NewForConfig(restConfig)
			if err != nil {
				return err
			}
			if namespace == "" {
				namespace, _, err = clientConfig.Namespace()
				if err != nil {
					return err
				}
			}
			level, err := log.ParseLevel(logLevel)
			if err != nil {
				return err
			}
			log.SetLevel(level)

			switch strings.ToLower(logFormat) {
			case "json":
				log.SetFormatter(&log.JSONFormatter{})
			case "text":
				if os.Getenv("FORCE_LOG_COLORS") == "1" {
					log.SetFormatter(&log.TextFormatter{ForceColors: true})
				}
			default:
				return fmt.Errorf("Unknown log format '%s'", logFormat)
			}

			argocdService, err := argocd.NewArgoCDService(k8sClient, namespace, argocdRepoServer, argocdRepoServerPlaintext, argocdRepoServerStrictTLS)
			if err != nil {
				return err
			}
			defer argocdService.Close()

			k8s.ConfigMapName = configMapName
			k8s.SecretName = secretName

			registry := notificationscontroller.NewMetricsRegistry("argocd")
			http.Handle("/metrics", promhttp.HandlerFor(prometheus.Gatherers{registry, prometheus.DefaultGatherer}, promhttp.HandlerOpts{}))

			go func() {
				log.Fatal(http.ListenAndServe(fmt.Sprintf("0.0.0.0:%d", metricsPort), http.DefaultServeMux))
			}()
			log.Infof("serving metrics on port %d", metricsPort)
			log.Infof("loading configuration %d", metricsPort)

			ctrl := controller.NewController(k8sClient, dynamicClient, argocdService, namespace, appLabelSelector, registry)
			err = ctrl.Init(context.Background())
			if err != nil {
				return err
			}

			go ctrl.Run(context.Background(), processorsCount)
			<-context.Background().Done()
			return nil
		},
	}
	clientConfig = k8s.AddK8SFlagsToCmd(&command)
	command.Flags().IntVar(&processorsCount, "processors-count", 1, "Processors count.")
	command.Flags().StringVar(&appLabelSelector, "app-label-selector", "", "App label selector.")
	command.Flags().StringVar(&namespace, "namespace", "", "Namespace which controller handles. Current namespace if empty.")
	command.Flags().StringVar(&logLevel, "loglevel", "info", "Set the logging level. One of: debug|info|warn|error")
	command.Flags().StringVar(&logFormat, "logformat", "text", "Set the logging format. One of: text|json")
	command.Flags().IntVar(&metricsPort, "metrics-port", defaultMetricsPort, "Metrics port")
	command.Flags().StringVar(&argocdRepoServer, "argocd-repo-server", "argocd-repo-server:8081", "Argo CD repo server address")
	command.Flags().BoolVar(&argocdRepoServerPlaintext, "argocd-repo-server-plaintext", false, "Use a plaintext client (non-TLS) to connect to repository server")
	command.Flags().BoolVar(&argocdRepoServerStrictTLS, "argocd-repo-server-strict-tls", false, "Perform strict validation of TLS certificates when connecting to repo server")
	command.Flags().StringVar(&configMapName, "config-map-name", "argocd-notifications-cm", "Set notifications ConfigMap name")
	command.Flags().StringVar(&secretName, "secret-name", "argocd-notifications-secret", "Set notifications Secret name")
	return &command
}
