package repo

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestToHttps(t *testing.T) {
	for in, expected := range map[string]string{
		"git@github.com:argoproj/argo-cd.git":    "https://github.com/argoproj/argo-cd.git",
		"http://github.com/argoproj/argo-cd.git": "https://github.com/argoproj/argo-cd.git",
	} {
		actual := repoURLToHTTPS(in)
		assert.Equal(t, actual, expected)
	}
}

func TestParseFullName(t *testing.T) {
	for in, expected := range map[string]string{
		"git@github.com:argoproj/argo-cd.git":             "argoproj/argo-cd",
		"http://github.com/argoproj/argo-cd.git":          "argoproj/argo-cd",
		"http://github.com/argoproj/argo-cd":              "argoproj/argo-cd",
		"https://user@bitbucket.org/argoproj/argo-cd.git": "argoproj/argo-cd",
		"git@gitlab.com:argoproj/argo-cd.git":             "argoproj/argo-cd",
		"https://gitlab.com/argoproj/argo-cd.git":         "argoproj/argo-cd",
	} {
		actual := fullNameByRepoURL(in)
		assert.Equal(t, actual, expected)
	}
}
