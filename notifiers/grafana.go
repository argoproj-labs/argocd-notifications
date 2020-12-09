package notifiers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"

	httputil "github.com/argoproj-labs/argocd-notifications/shared/http"

	log "github.com/sirupsen/logrus"
)

type GrafanaOptions struct {
	ApiUrl             string `json:"apiUrl"`
	ApiKey             string `json:"apiKey"`
	InsecureSkipVerify bool   `json:"insecureSkipVerify"`
}

type grafanaNotifier struct {
	opts GrafanaOptions
}

func NewGrafanaNotifier(opts GrafanaOptions) Notifier {
	return &grafanaNotifier{opts: opts}
}

type GrafanaAnnotation struct {
	Time     int64    `json:"time"` // unix ts in ms
	IsRegion bool     `json:"isRegion"`
	Tags     []string `json:"tags"`
	Text     string   `json:"text"`
}

func (n *grafanaNotifier) Send(notification Notification, tags string) error {
	ga := GrafanaAnnotation{
		Time:     time.Now().Unix() * 1000, // unix ts in ms
		IsRegion: false,
		Tags:     strings.Split(tags, "|"),
		Text:     notification.Title,
	}

	client := &http.Client{
		Transport: httputil.NewLoggingRoundTripper(
			httputil.NewTransport(n.opts.ApiUrl, n.opts.InsecureSkipVerify), log.WithField("notifier", "grafana")),
	}

	jsonValue, _ := json.Marshal(ga)
	apiUrl, err := url.Parse(n.opts.ApiUrl)

	if err != nil {
		return err
	}
	annotationApi := *apiUrl
	annotationApi.Path = path.Join(apiUrl.Path, "annotations")
	req, err := http.NewRequest("POST", annotationApi.String(), bytes.NewBuffer(jsonValue))
	if err != nil {
		log.Errorf("Failed to create grafana annotation request: %s", err)
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", n.opts.ApiKey))

	_, err = client.Do(req)
	return err
}
