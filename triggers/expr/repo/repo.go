package repo

import (
	"regexp"
	"strings"

	"github.com/argoproj-labs/argocd-notifications/shared/text"

	giturls "github.com/whilp/git-urls"
)

var (
	gitSuffix = regexp.MustCompile(`\.git$`)
)

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

func NewExprs() map[string]interface{} {
	return map[string]interface{}{
		"RepoURLToHTTPS":    repoURLToHTTPS,
		"FullNameByRepoURL": fullNameByRepoURL,
	}
}
