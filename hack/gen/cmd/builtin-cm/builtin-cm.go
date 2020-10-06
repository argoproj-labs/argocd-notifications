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
			Name: "argocd-notifications-builtin-cm",
		},
		Data: make(map[string]string),
	}
	wd, err := os.Getwd()
	dieOnError(err, "Failed to get current working directory")
	target := path.Join(wd, "manifests/controller/argocd-notifications-builtin-cm.yaml")

	templatesDir := path.Join(wd, "builtin/templates")
	triggersDir := path.Join(wd, "builtin/triggers")

	cnf, err := tools.BuildConfigFromFS(templatesDir, triggersDir)
	dieOnError(err, "Failed to build builtin config")

	configBytes, err := yaml.Marshal(cnf)
	dieOnError(err, "Failed to marshal final builtin config")

	cm.Data["config.yaml"] = string(configBytes)
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
