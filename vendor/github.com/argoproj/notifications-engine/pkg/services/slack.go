package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	texttemplate "text/template"

	httputil "github.com/argoproj/notifications-engine/pkg/util/http"

	log "github.com/sirupsen/logrus"
	"github.com/slack-go/slack"
	"golang.org/x/time/rate"
)

// No rate limit unless Slack requests it (allows for Slack to control bursting)
var rateLimiter = rate.NewLimiter(rate.Inf, 1)
var threadTSs = map[string]map[string]string{}

type SlackNotification struct {
	Attachments     string `json:"attachments,omitempty"`
	Blocks          string `json:"blocks,omitempty"`
	GroupingKey     string `json:"groupingKey"`
	NotifyBroadcast bool   `json:"notifyBroadcast"`
}

func (n *SlackNotification) GetTemplater(name string, f texttemplate.FuncMap) (Templater, error) {
	slackAttachments, err := texttemplate.New(name).Funcs(f).Parse(n.Attachments)
	if err != nil {
		return nil, err
	}
	slackBlocks, err := texttemplate.New(name).Funcs(f).Parse(n.Blocks)
	if err != nil {
		return nil, err
	}
	groupingKey, err := texttemplate.New(name).Funcs(f).Parse(n.GroupingKey)
	if err != nil {
		return nil, err
	}

	return func(notification *Notification, vars map[string]interface{}) error {
		if notification.Slack == nil {
			notification.Slack = &SlackNotification{}
		}
		var slackAttachmentsData bytes.Buffer
		if err := slackAttachments.Execute(&slackAttachmentsData, vars); err != nil {
			return err
		}
		notification.Slack.Attachments = slackAttachmentsData.String()

		var slackBlocksData bytes.Buffer
		if err := slackBlocks.Execute(&slackBlocksData, vars); err != nil {
			return err
		}
		notification.Slack.Blocks = slackBlocksData.String()

		var groupingKeyData bytes.Buffer
		if err := groupingKey.Execute(&groupingKeyData, vars); err != nil {
			return err
		}
		notification.Slack.GroupingKey = groupingKeyData.String()

		notification.Slack.NotifyBroadcast = n.NotifyBroadcast
		return nil
	}, nil
}

type SlackOptions struct {
	Username           string   `json:"username"`
	Icon               string   `json:"icon"`
	Token              string   `json:"token"`
	SigningSecret      string   `json:"signingSecret"`
	Channels           []string `json:"channels"`
	InsecureSkipVerify bool     `json:"insecureSkipVerify"`
	ApiURL             string   `json:"apiURL"`
}

type slackService struct {
	opts SlackOptions
}

var validIconEmoji = regexp.MustCompile(`^:.+:$`)

func NewSlackService(opts SlackOptions) NotificationService {
	return &slackService{opts: opts}
}

func (s *slackService) Send(notification Notification, dest Destination) error {
	apiURL := slack.APIURL
	if s.opts.ApiURL != "" {
		apiURL = s.opts.ApiURL
	}
	transport := httputil.NewTransport(apiURL, s.opts.InsecureSkipVerify)
	client := &http.Client{
		Transport: httputil.NewLoggingRoundTripper(transport, log.WithField("service", "slack")),
	}
	sl := slack.New(s.opts.Token, slack.OptionHTTPClient(client), slack.OptionAPIURL(apiURL))
	msgOptions := []slack.MsgOption{slack.MsgOptionText(notification.Message, false)}
	if s.opts.Username != "" {
		msgOptions = append(msgOptions, slack.MsgOptionUsername(s.opts.Username))
	}
	if s.opts.Icon != "" {
		if validIconEmoji.MatchString(s.opts.Icon) {
			msgOptions = append(msgOptions, slack.MsgOptionIconEmoji(s.opts.Icon))
		} else if isValidIconURL(s.opts.Icon) {
			msgOptions = append(msgOptions, slack.MsgOptionIconURL(s.opts.Icon))
		} else {
			log.Warnf("Icon reference '%v' is not a valid emoij or url", s.opts.Icon)
		}
	}

	if notification.Slack != nil {
		attachments := make([]slack.Attachment, 0)
		if notification.Slack.Attachments != "" {
			if err := json.Unmarshal([]byte(notification.Slack.Attachments), &attachments); err != nil {
				return fmt.Errorf("failed to unmarshal attachments '%s' : %v", notification.Slack.Attachments, err)
			}
		}

		blocks := slack.Blocks{}
		if notification.Slack.Blocks != "" {
			if err := json.Unmarshal([]byte(notification.Slack.Blocks), &blocks); err != nil {
				return fmt.Errorf("failed to unmarshal blocks '%s' : %v", notification.Slack.Blocks, err)
			}
		}
		msgOptions = append(msgOptions, slack.MsgOptionAttachments(attachments...), slack.MsgOptionBlocks(blocks.BlockSet...))
	}

	if _, ok := threadTSs[dest.Recipient]; !ok {
		threadTSs[dest.Recipient] = map[string]string{}
	}

	if notification.Slack.NotifyBroadcast {
		msgOptions = append(msgOptions, slack.MsgOptionBroadcast())
	}

	if lastTs, ok := threadTSs[dest.Recipient][notification.Slack.GroupingKey]; ok && lastTs != "" && notification.Slack.GroupingKey != "" {
		msgOptions = append(msgOptions, slack.MsgOptionTS(lastTs))
	}

	ctx := context.TODO()
	var err error
	for {
		err = rateLimiter.Wait(ctx)
		if err != nil {
			break
		}
		_, ts, err := sl.PostMessageContext(ctx, dest.Recipient, msgOptions...)
		if err != nil {
			if rateLimitedError, ok := err.(*slack.RateLimitedError); ok {
				rateLimiter.SetLimit(rate.Every(rateLimitedError.RetryAfter))
			} else {
				break
			}
		} else {
			if lastTs, ok := threadTSs[dest.Recipient][notification.Slack.GroupingKey]; !ok || lastTs == "" {
				threadTSs[dest.Recipient][notification.Slack.GroupingKey] = ts
			}
			// No error, so remove rate limit
			rateLimiter.SetLimit(rate.Inf)
			break
		}
	}
	return err
}

// GetSigningSecret exposes signing secret for slack bot
func (s *slackService) GetSigningSecret() string {
	return s.opts.SigningSecret
}

func isValidIconURL(iconURL string) bool {
	_, err := url.ParseRequestURI(iconURL)
	if err != nil {
		return false
	}

	u, err := url.Parse(iconURL)
	if err != nil || (u.Scheme == "" || !(u.Scheme == "http" || u.Scheme == "https")) || u.Host == "" {
		return false
	}

	return true
}
