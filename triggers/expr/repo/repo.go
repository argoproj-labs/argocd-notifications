package repo

import (
	"context"
	"errors"
	"regexp"
	"strings"

	giturls "github.com/whilp/git-urls"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/argoproj-labs/argocd-notifications/shared/argocd"
	"github.com/argoproj-labs/argocd-notifications/shared/text"
	"github.com/argoproj-labs/argocd-notifications/triggers/expr/shared"
)

var (
	gitSuffix = regexp.MustCompile(`\.git$`)
)

func getCommitMetadata(commitSHA string, app *unstructured.Unstructured, argocdService argocd.Service) (*shared.CommitMetadata, error) {
	repoURL, ok, err := unstructured.NestedString(app.Object, "spec", "source", "repoURL")
	if err != nil {
		return nil, err
	}
	if !ok {
		panic(errors.New("failed to get application source repo URL"))
	}
	meta, err := argocdService.GetCommitMetadata(context.Background(), repoURL, commitSHA)
	if err != nil {
		return nil, err
	}
	return meta, nil
}

func fullNameByRepoURL(rawURL string) string {
	parsed, err := giturls.Parse(rawURL)
	if err != nil {
		panic(err)
	}

	path := gitSuffix.ReplaceAllString(parsed.Path, "")
	if pathParts := text.SplitRemoveEmpty(path, "/"); len(pathParts) >= 2 {
		return strings.Join(pathParts[:2], "/")
	}

	return path
}

func repoURLToHTTPS(rawURL string) string {
	parsed, err := giturls.Parse(rawURL)
	if err != nil {
		panic(err)
	}
	parsed.Scheme = "https"
	parsed.User = nil
	return parsed.String()
}

func NewExprs(argocdService argocd.Service, app *unstructured.Unstructured) map[string]interface{} {
	return map[string]interface{}{
		"RepoURLToHTTPS":    repoURLToHTTPS,
		"FullNameByRepoURL": fullNameByRepoURL,
		"GetCommitMetadata": func(commitSHA string) interface{} {
			meta, err := getCommitMetadata(commitSHA, app, argocdService)
			if err != nil {
				panic(err)
			}

			return *meta
		},
	}
}
