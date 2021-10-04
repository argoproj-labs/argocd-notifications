package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"

	httputil "github.com/argoproj/notifications-engine/pkg/util/http"

	log "github.com/sirupsen/logrus"
)

type GrafanaOptions struct {
	ApiUrl             string `json:"apiUrl"`
	ApiKey             string `json:"apiKey"`
	InsecureSkipVerify bool   `json:"insecureSkipVerify"`
}

type grafanaService struct {
	opts GrafanaOptions
}

func NewGrafanaService(opts GrafanaOptions) NotificationService {
	return &grafanaService{opts: opts}
}

type GrafanaAnnotation struct {
	Time     int64    `json:"time"` // unix ts in ms
	IsRegion bool     `json:"isRegion"`
	Tags     []string `json:"tags"`
	Text     string   `json:"text"`
}

func (s *grafanaService) Send(notification Notification, dest Destination) error {
	ga := GrafanaAnnotation{
		Time:     time.Now().Unix() * 1000, // unix ts in ms
		IsRegion: false,
		Tags:     strings.Split(dest.Recipient, "|"),
		Text:     notification.Message,
	}

	client := &http.Client{
		Transport: httputil.NewLoggingRoundTripper(
			httputil.NewTransport(s.opts.ApiUrl, s.opts.InsecureSkipVerify), log.WithField("service", "grafana")),
	}

	jsonValue, _ := json.Marshal(ga)
	apiUrl, err := url.Parse(s.opts.ApiUrl)

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
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", s.opts.ApiKey))

	_, err = client.Do(req)
	return err
}
