package services

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"strings"
	texttemplate "text/template"

	"github.com/argoproj/argo-cd/util/text"
	"github.com/bradleyfalzon/ghinstallation"
	"github.com/google/go-github/v33/github"
	log "github.com/sirupsen/logrus"

	"github.com/argoproj-labs/argocd-notifications/expr/repo"
	httputil "github.com/argoproj-labs/argocd-notifications/pkg/util/http"
)

type GitHubOptions struct {
	AppID             int64  `json:"appID"`
	InstallationID    int64  `json:"installationID"`
	PrivateKey        string `json:"privateKey"`
	EnterpriseBaseURL string `json:"enterpriseBaseURL"`
}

type GitHubNotification struct {
	repoURL   string
	revision  string
	State     string `json:"state,omitempty"`
	Label     string `json:"label,omitempty"`
	TargetURL string `json:"targetURL,omitempty"`
}

const (
	repoURLtemplate  = "{{.app.spec.source.repoURL}}"
	revisionTemplate = "{{.app.status.sync.revision}}"
)

func (g *GitHubNotification) GetTemplater(name string, f texttemplate.FuncMap) (Templater, error) {
	repoURL, err := texttemplate.New(name).Funcs(f).Parse(repoURLtemplate)
	if err != nil {
		return nil, err
	}

	revision, err := texttemplate.New(name).Funcs(f).Parse(revisionTemplate)
	if err != nil {
		return nil, err
	}

	state, err := texttemplate.New(name).Funcs(f).Parse(g.State)
	if err != nil {
		return nil, err
	}

	label, err := texttemplate.New(name).Funcs(f).Parse(g.Label)
	if err != nil {
		return nil, err
	}

	targetURL, err := texttemplate.New(name).Funcs(f).Parse(g.TargetURL)
	if err != nil {
		return nil, err
	}

	return func(notification *Notification, vars map[string]interface{}) error {
		if notification.GitHub == nil {
			notification.GitHub = &GitHubNotification{}
		}

		var repoData bytes.Buffer
		if err := repoURL.Execute(&repoData, vars); err != nil {
			return err
		}
		notification.GitHub.repoURL = repoData.String()

		var revisionData bytes.Buffer
		if err := revision.Execute(&revisionData, vars); err != nil {
			return err
		}
		notification.GitHub.revision = revisionData.String()

		var stateData bytes.Buffer
		if err := state.Execute(&stateData, vars); err != nil {
			return err
		}
		notification.GitHub.State = stateData.String()

		var labelData bytes.Buffer
		if err := label.Execute(&labelData, vars); err != nil {
			return err
		}
		notification.GitHub.Label = labelData.String()

		var targetData bytes.Buffer
		if err := targetURL.Execute(&targetData, vars); err != nil {
			return err
		}
		notification.GitHub.TargetURL = targetData.String()

		return nil
	}, nil
}

func NewGitHubService(opts GitHubOptions) (NotificationService, error) {
	url := "https://api.github.com"
	if opts.EnterpriseBaseURL != "" {
		url = opts.EnterpriseBaseURL
	}

	tr := httputil.NewLoggingRoundTripper(
		httputil.NewTransport(url, false), log.WithField("service", "github"))
	itr, err := ghinstallation.New(tr, opts.AppID, opts.InstallationID, []byte(opts.PrivateKey))
	if err != nil {
		return nil, err
	}

	var client *github.Client
	if opts.EnterpriseBaseURL == "" {
		client = github.NewClient(&http.Client{Transport: itr})
	} else {
		itr.BaseURL = opts.EnterpriseBaseURL
		client, err = github.NewEnterpriseClient(opts.EnterpriseBaseURL, "", &http.Client{Transport: itr})
		if err != nil {
			return nil, err
		}
	}

	return &gitHubService{
		opts:   opts,
		client: client,
	}, nil
}

type gitHubService struct {
	opts GitHubOptions

	client *github.Client
}

func (g gitHubService) Send(notification Notification, _ Destination) error {
	if notification.GitHub == nil {
		return fmt.Errorf("config is empty")
	}

	u := strings.Split(repo.FullNameByRepoURL(notification.GitHub.repoURL), "/")
	// maximum is 140 characters
	description := text.Trunc(notification.Message, 140)
	_, _, err := g.client.Repositories.CreateStatus(
		context.Background(),
		u[0],
		u[1],
		notification.GitHub.revision,
		&github.RepoStatus{
			State:       &notification.GitHub.State,
			Description: &description,
			Context:     &notification.GitHub.Label,
			TargetURL:   &notification.GitHub.TargetURL,
		},
	)

	return err
}
