package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"

	"github.com/argoproj-labs/argocd-notifications/cmd/tools"
	"github.com/ghodss/yaml"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func main() {
	cm := v1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "argocd-notifications-cm",
		},
		Data: make(map[string]string),
	}
	wd, err := os.Getwd()
	dieOnError(err, "Failed to get current working directory")
	target := path.Join(wd, "manifests/controller/argocd-notifications-cm.yaml")

	templatesDir := path.Join(wd, "builtin/templates")
	triggersDir := path.Join(wd, "builtin/triggers")

	templates, triggers, err := tools.BuildConfigFromFS(templatesDir, triggersDir)
	dieOnError(err, "Failed to build builtin config")

	for _, trigger := range triggers {
		name := trigger.Name
		trigger.Name = ""
		t, err := yaml.Marshal(trigger)
		dieOnError(err, "Failed to marshal trigger")
		cm.Data[fmt.Sprintf("trigger.%s", name)] = string(t)
	}

	for _, template := range templates {
		name := template.Name
		template.Name = ""
		t, err := yaml.Marshal(template)
		dieOnError(err, "Failed to marshal template")
		cm.Data[fmt.Sprintf("template.%s", name)] = string(t)
	}

	d, err := yaml.Marshal(cm)
	dieOnError(err, "Failed to marshal final configmap")

	err = ioutil.WriteFile(target, d, 0644)
	dieOnError(err, "Failed to write builtin configmap")

}

func dieOnError(err error, msg string) {
	if err != nil {
		fmt.Printf("[ERROR] %s: %v", msg, err)
		os.Exit(1)
	}
}
