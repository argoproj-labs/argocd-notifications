package tools

import (
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/ghodss/yaml"

	"github.com/argoproj-labs/argocd-notifications/pkg/templates"
	"github.com/argoproj-labs/argocd-notifications/triggers"
)

func BuildConfigFromFS(templatesDir string, triggersDir string) ([]templates.NotificationTemplate, []triggers.NotificationTrigger, error) {
	var notificationTemplates []templates.NotificationTemplate
	err := filepath.Walk(templatesDir, func(p string, info os.FileInfo, e error) error {
		if e != nil {
			return e
		}
		if info.IsDir() {
			return nil
		}
		data, err := ioutil.ReadFile(p)
		if err != nil {
			return err
		}
		template := templates.NotificationTemplate{
			Name: strings.Split(path.Base(p), ".")[0],
		}
		if err := yaml.Unmarshal(data, &template); err != nil {
			return err
		}
		notificationTemplates = append(notificationTemplates, template)
		return nil
	})
	if err != nil {
		return nil, nil, err
	}

	var notificationTriggers []triggers.NotificationTrigger
	err = filepath.Walk(triggersDir, func(p string, info os.FileInfo, e error) error {
		if e != nil {
			return e
		}
		if info.IsDir() {
			return nil
		}
		data, err := ioutil.ReadFile(p)
		if err != nil {
			return err
		}
		trigger := triggers.NotificationTrigger{
			Name: strings.Split(path.Base(p), ".")[0],
		}
		if err := yaml.Unmarshal(data, &trigger); err != nil {
			return err
		}
		notificationTriggers = append(notificationTriggers, trigger)
		return nil
	})
	if err != nil {
		return nil, nil, err
	}
	return notificationTemplates, notificationTriggers, nil
}
