package tools

import (
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/argoproj-labs/argocd-notifications/shared/settings"
	"github.com/argoproj-labs/argocd-notifications/triggers"
	"github.com/ghodss/yaml"
)

func BuildConfigFromFS(templatesDir string, triggersDir string) (*settings.Config, error) {
	notificationTemplates := []triggers.NotificationTemplate{}
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
		template := triggers.NotificationTemplate{
			Name: strings.Split(path.Base(p), ".")[0],
		}
		if err := yaml.Unmarshal(data, &template); err != nil {
			return err
		}
		notificationTemplates = append(notificationTemplates, template)
		return nil
	})
	if err != nil {
		return nil, err
	}

	notificationTriggers := []triggers.NotificationTrigger{}
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
		return nil, err
	}
	cnf := &settings.Config{
		Triggers:  notificationTriggers,
		Templates: notificationTemplates,
	}
	return cnf, nil
}
