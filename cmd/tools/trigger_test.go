package tools

import (
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"testing"

	"github.com/ghodss/yaml"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic"
	dynamicfake "k8s.io/client-go/dynamic/fake"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"

	"github.com/argoproj-labs/argocd-notifications/shared/settings"
	testingutil "github.com/argoproj-labs/argocd-notifications/testing"
	"github.com/argoproj-labs/argocd-notifications/triggers"
)

func newTestContext(stdout io.Writer, stderr io.Writer, config settings.Config, apps ...runtime.Object) (*commandContext, func(), error) {
	configData, err := yaml.Marshal(config)
	if err != nil {
		return nil, nil, err
	}
	cm := v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name: settings.ConfigMapName,
		},
		Data: map[string]string{
			"config.yaml": string(configData),
		},
	}
	cmData, err := yaml.Marshal(cm)
	if err != nil {
		return nil, nil, err
	}
	tmpFile, err := ioutil.TempFile("", "*-cm.yaml")
	if err != nil {
		return nil, nil, err
	}
	defer func() {
		_ = tmpFile.Close()
	}()
	_, err = tmpFile.Write(cmData)
	if err != nil {
		return nil, nil, err
	}

	ctx := &commandContext{
		stdout:        stdout,
		stderr:        stderr,
		secretPath:    ":empty",
		configMapPath: tmpFile.Name(),
		getK8SClients: func() (kubernetes.Interface, dynamic.Interface, string, error) {
			dynamicClient := dynamicfake.NewSimpleDynamicClient(runtime.NewScheme(), apps...)
			return fake.NewSimpleClientset(), dynamicClient, "default", nil
		},
	}
	return ctx, func() {
		_ = os.RemoveAll(tmpFile.Name())
	}, nil
}

func TestTriggerRun(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	ctx, closer, err := newTestContext(&stdout, &stderr, settings.Config{
		Triggers: []triggers.NotificationTrigger{{
			Name:      "my-trigger",
			Condition: "app.metadata.name == 'guestbook'",
			Template:  "my-template",
		}},
		Templates: []triggers.NotificationTemplate{{
			Name: "my-template",
		}},
	}, testingutil.NewApp("guestbook"))
	if !assert.NoError(t, err) {
		return
	}
	defer closer()

	command := newTriggerRunCommand(ctx)
	err = command.RunE(command, []string{"my-trigger", "guestbook"})
	assert.NoError(t, err)
	assert.Empty(t, stderr.String())
	assert.Contains(t, stdout.String(), "true")
}

func TestTriggerGet(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	ctx, closer, err := newTestContext(&stdout, &stderr, settings.Config{
		Triggers: []triggers.NotificationTrigger{{
			Name:      "my-trigger1",
			Template:  "my-template",
			Condition: "true",
		}, {
			Name:      "my-trigger2",
			Template:  "my-template",
			Condition: "false",
		}},
		Templates: []triggers.NotificationTemplate{{
			Name: "my-template",
		}},
	})
	if !assert.NoError(t, err) {
		return
	}
	defer closer()

	command := newTriggerGetCommand(ctx)
	err = command.RunE(command, nil)
	assert.NoError(t, err)
	assert.Empty(t, stderr.String())
	assert.Contains(t, stdout.String(), "my-trigger1")
	assert.Contains(t, stdout.String(), "my-trigger2")
}
