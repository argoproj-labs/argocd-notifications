package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	texttemplate "text/template"

	log "github.com/sirupsen/logrus"

	httputil "github.com/argoproj/notifications-engine/pkg/util/http"
)

type TeamsNotification struct {
	Template        string `json:"template,omitempty"`
	Title           string `json:"title,omitempty"`
	Summary         string `json:"summary,omitempty"`
	Text            string `json:"text,omitempty"`
	ThemeColor      string `json:"themeColor,omitempty"`
	Facts           string `json:"facts,omitempty"`
	Sections        string `json:"sections,omitempty"`
	PotentialAction string `json:"potentialAction,omitempty"`
}

func (n *TeamsNotification) GetTemplater(name string, f texttemplate.FuncMap) (Templater, error) {
	template, err := texttemplate.New(name).Funcs(f).Parse(n.Template)
	if err != nil {
		return nil, fmt.Errorf("error in '%s' teams.template : %w", name, err)
	}

	title, err := texttemplate.New(name).Funcs(f).Parse(n.Title)
	if err != nil {
		return nil, fmt.Errorf("error in '%s' teams.title : %w", name, err)
	}

	summary, err := texttemplate.New(name).Funcs(f).Parse(n.Summary)
	if err != nil {
		return nil, fmt.Errorf("error in '%s' teams.summary : %w", name, err)
	}

	text, err := texttemplate.New(name).Funcs(f).Parse(n.Text)
	if err != nil {
		return nil, fmt.Errorf("error in '%s' teams.text : %w", name, err)
	}

	themeColor, err := texttemplate.New(name).Funcs(f).Parse(n.ThemeColor)
	if err != nil {
		return nil, fmt.Errorf("error in '%s' teams.themeColor: %w", name, err)
	}

	facts, err := texttemplate.New(name).Funcs(f).Parse(n.Facts)
	if err != nil {
		return nil, fmt.Errorf("error in '%s' teams.facts : %w", name, err)
	}

	sections, err := texttemplate.New(name).Funcs(f).Parse(n.Sections)
	if err != nil {
		return nil, fmt.Errorf("error in '%s' teams.sections : %w", name, err)
	}

	potentialActions, err := texttemplate.New(name).Funcs(f).Parse(n.PotentialAction)
	if err != nil {
		return nil, fmt.Errorf("error in '%s' teams.potentialAction: %w", name, err)
	}

	return func(notification *Notification, vars map[string]interface{}) error {
		if notification.Teams == nil {
			notification.Teams = &TeamsNotification{}
		}

		var templateBuff bytes.Buffer
		if err := template.Execute(&templateBuff, vars); err != nil {
			return err
		}
		if val := templateBuff.String(); val != "" {
			notification.Teams.Template = val
		}

		var titleBuff bytes.Buffer
		if err := title.Execute(&titleBuff, vars); err != nil {
			return err
		}
		if val := titleBuff.String(); val != "" {
			notification.Teams.Title = val
		}

		var summaryBuff bytes.Buffer
		if err := summary.Execute(&summaryBuff, vars); err != nil {
			return err
		}
		if val := summaryBuff.String(); val != "" {
			notification.Teams.Summary = val
		}

		var textBuff bytes.Buffer
		if err := text.Execute(&textBuff, vars); err != nil {
			return err
		}
		if val := textBuff.String(); val != "" {
			notification.Teams.Text = val
		}

		var themeColorBuff bytes.Buffer
		if err := themeColor.Execute(&themeColorBuff, vars); err != nil {
			return err
		}
		if val := themeColorBuff.String(); val != "" {
			notification.Teams.ThemeColor = val
		}

		var factsData bytes.Buffer
		if err := facts.Execute(&factsData, vars); err != nil {
			return err
		}
		if val := factsData.String(); val != "" {
			notification.Teams.Facts = val
		}

		var sectionsBuff bytes.Buffer
		if err := sections.Execute(&sectionsBuff, vars); err != nil {
			return err
		}
		if val := sectionsBuff.String(); val != "" {
			notification.Teams.Sections = val
		}

		var actionsData bytes.Buffer
		if err := potentialActions.Execute(&actionsData, vars); err != nil {
			return err
		}
		if val := actionsData.String(); val != "" {
			notification.Teams.PotentialAction = val
		}

		return nil
	}, nil
}

type TeamsOptions struct {
	RecipientUrls map[string]string `json:"recipientUrls"`
}

type teamsService struct {
	opts TeamsOptions
}

func NewTeamsService(opts TeamsOptions) NotificationService {
	return &teamsService{opts: opts}
}

func (s teamsService) Send(notification Notification, dest Destination) error {
	webhookUrl, ok := s.opts.RecipientUrls[dest.Recipient]
	if !ok {
		return fmt.Errorf("no teams webhook configured for recipient %s", dest.Recipient)
	}
	transport := httputil.NewTransport(webhookUrl, false)
	client := &http.Client{
		Transport: httputil.NewLoggingRoundTripper(transport, log.WithField("service", "teams")),
	}

	message, err := teamsNotificationToReader(notification)
	if err != nil {
		return err
	}

	response, err := client.Post(webhookUrl, "application/json", bytes.NewReader(message))

	if err != nil {
		return err
	}

	defer func() {
		_ = response.Body.Close()
	}()

	bodyBytes, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return err
	}

	if string(bodyBytes) != "1" {
		return fmt.Errorf("teams webhook post error: %s", bodyBytes)
	}

	return nil
}

func teamsNotificationToMessage(n Notification) (*teamsMessage, error) {
	message := &teamsMessage{
		Type:    "MessageCard",
		Context: "https://schema.org/extensions",
		Text:    n.Message,
	}

	if n.Teams == nil {
		return message, nil
	}

	if n.Teams.Title != "" {
		message.Title = n.Teams.Title
	}

	if n.Teams.Summary != "" {
		message.Summary = n.Teams.Summary
	}

	if n.Teams.Text != "" {
		message.Text = n.Teams.Text
	}

	if n.Teams.ThemeColor != "" {
		message.ThemeColor = n.Teams.ThemeColor
	}

	if n.Teams.Sections != "" {
		unmarshalledSections := make([]teamsSection, 2)
		err := json.Unmarshal([]byte(n.Teams.Sections), &unmarshalledSections)
		if err != nil {
			return nil, fmt.Errorf("teams facts unmarshalling error %w", err)
		}
		message.Sections = unmarshalledSections
	}

	if n.Teams.Facts != "" {
		unmarshalledFacts := make([]map[string]interface{}, 2)
		err := json.Unmarshal([]byte(n.Teams.Facts), &unmarshalledFacts)
		if err != nil {
			return nil, fmt.Errorf("teams facts unmarshalling error %w", err)
		}
		message.Sections = append(message.Sections, teamsSection{
			"facts": unmarshalledFacts,
		})
	}

	if n.Teams.PotentialAction != "" {
		unmarshalledActions := make([]teamsAction, 2)
		err := json.Unmarshal([]byte(n.Teams.PotentialAction), &unmarshalledActions)
		if err != nil {
			return nil, fmt.Errorf("teams actions unmarshalling error %w", err)
		}
		message.PotentialAction = unmarshalledActions
	}

	return message, nil
}

func teamsNotificationToReader(n Notification) ([]byte, error) {
	if n.Teams != nil && n.Teams.Template != "" {
		return []byte(n.Teams.Template), nil
	}

	message, err := teamsNotificationToMessage(n)

	if err != nil {
		return nil, err
	}

	marshal, err := json.Marshal(message)

	if err != nil {
		return nil, err
	}

	return marshal, nil
}

type teamsMessage struct {
	Type            string         `json:"@type"`
	Context         string         `json:"context"`
	Title           string         `json:"title"`
	Summary         string         `json:"summary"`
	Text            string         `json:"text"`
	ThemeColor      string         `json:"themeColor,omitempty"`
	PotentialAction []teamsAction  `json:"potentialAction,omitempty"`
	Sections        []teamsSection `json:"sections,omitempty"`
}

type teamsSection = map[string]interface{}
type teamsAction map[string]interface{}
