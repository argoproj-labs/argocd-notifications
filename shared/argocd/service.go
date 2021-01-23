package argocd

import (
	"context"

	"github.com/argoproj-labs/argocd-notifications/expr/shared"
	"github.com/argoproj/argo-cd/pkg/apis/application/v1alpha1"
	"github.com/argoproj/argo-cd/reposerver/apiclient"
	"github.com/argoproj/argo-cd/util/db"
	"github.com/argoproj/argo-cd/util/settings"
	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
)

//go:generate mockgen -destination=./mocks/service.go -package=mocks github.com/argoproj-labs/argocd-notifications/shared/argocd Service

type Service interface {
	GetCommitMetadata(ctx context.Context, repoURL string, commitSHA string) (*shared.CommitMetadata, error)
	GetAppDetails(ctx context.Context, appSource *v1alpha1.ApplicationSource) (*shared.AppDetail, error)
}

func NewArgoCDService(clientset kubernetes.Interface, namespace string, repoServerAddress string) (*argoCDService, error) {
	ctx, cancel := context.WithCancel(context.Background())
	settingsMgr := settings.NewSettingsManager(ctx, clientset, namespace)
	repoClientset := apiclient.NewRepoServerClientset(repoServerAddress, 5)
	closer, repoClient, err := repoClientset.NewRepoServerClient()
	if err != nil {
		cancel()
		return nil, err
	}

	dispose := func() {
		cancel()
		if err := closer.Close(); err != nil {
			log.Warnf("Failed to close repo server connection: %v", err)
		}
	}
	return &argoCDService{settingsMgr: settingsMgr, namespace: namespace, repoServerClient: repoClient, dispose: dispose}, nil
}

type argoCDService struct {
	clientset        kubernetes.Interface
	namespace        string
	settingsMgr      *settings.SettingsManager
	repoServerClient apiclient.RepoServerServiceClient
	dispose          func()
}

func (svc *argoCDService) GetCommitMetadata(ctx context.Context, repoURL string, commitSHA string) (*shared.CommitMetadata, error) {
	argocdDB := db.NewDB(svc.namespace, svc.settingsMgr, svc.clientset)
	repo, err := argocdDB.GetRepository(ctx, repoURL)
	if err != nil {
		return nil, err
	}
	metadata, err := svc.repoServerClient.GetRevisionMetadata(ctx, &apiclient.RepoServerRevisionMetadataRequest{
		Repo:     repo,
		Revision: commitSHA,
	})
	if err != nil {
		return nil, err
	}
	return &shared.CommitMetadata{
		Message: metadata.Message,
		Author:  metadata.Author,
		Date:    metadata.Date.Time,
		Tags:    metadata.Tags,
	}, nil
}

func (svc *argoCDService) getKustomizeOptions(source *v1alpha1.ApplicationSource) (*v1alpha1.KustomizeOptions, error) {
	kustomizeSettings, err := svc.settingsMgr.GetKustomizeSettings()
	if err != nil {
		return nil, err
	}
	return kustomizeSettings.GetOptions(*source)
}

func (svc *argoCDService) GetAppDetails(ctx context.Context, appSource *v1alpha1.ApplicationSource) (*shared.AppDetail, error) {
	argocdDB := db.NewDB(svc.namespace, svc.settingsMgr, svc.clientset)
	repo, err := argocdDB.GetRepository(ctx, appSource.RepoURL)
	if err != nil {
		return nil, err
	}
	helmRepos, err := argocdDB.ListHelmRepositories(ctx)
	if err != nil {
		return nil, err
	}
	kustomizeOptions, err := svc.getKustomizeOptions(appSource)
	if err != nil {
		return nil, err
	}
	appDetail, err := svc.repoServerClient.GetAppDetails(ctx, &apiclient.RepoServerAppDetailsQuery{
		Repo:             repo,
		Source:           appSource,
		Repos:            helmRepos,
		KustomizeOptions: kustomizeOptions,
	})
	if err != nil {
		return nil, err
	}
	has := &shared.HelmAppSpec{
		Name:           appDetail.Helm.Name,
		ValueFiles:     appDetail.Helm.ValueFiles,
		Parameters:     appDetail.Helm.Parameters,
		Values:         appDetail.Helm.Values,
		FileParameters: appDetail.Helm.FileParameters,
	}
	return &shared.AppDetail{
		Type:      appDetail.Type,
		Helm:      has,
		Ksonnet:   appDetail.Ksonnet,
		Kustomize: appDetail.Kustomize,
		Directory: appDetail.Directory,
	}, nil
}

func (svc *argoCDService) Close() {
	svc.dispose()
}
