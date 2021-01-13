package tools

import (
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"strings"
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

	"github.com/argoproj-labs/argocd-notifications/shared/k8s"
	testingutil "github.com/argoproj-labs/argocd-notifications/testing"
)

func newTestContext(stdout io.Writer, stderr io.Writer, data map[string]string, apps ...runtime.Object) (*commandContext, func(), error) {
	cm := v1.ConfigMap{
		TypeMeta: metav1.TypeMeta{APIVersion: "v1", Kind: "ConfigMap"},
		ObjectMeta: metav1.ObjectMeta{
			Name: k8s.ConfigMapName,
		},
		Data: data,
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
		stdin:         strings.NewReader(""),
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
	cmData := map[string]string{
		"trigger.my-trigger": `
- when: app.metadata.name == 'guestbook'
  send: [my-template]`,
		"template.my-template": `
message: hello {{.app.metadata.name}}`,
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	ctx, closer, err := newTestContext(&stdout, &stderr, cmData, testingutil.NewApp("guestbook"))
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
	cmData := map[string]string{
		"trigger.my-trigger1": `
- when: 'true'
  send: [my-template]`,
		"trigger.my-trigger2": `
- when: 'false'
  send: [my-template]`,
		"template.my-template": `
message: hello {{.app.metadata.name}}`,
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	ctx, closer, err := newTestContext(&stdout, &stderr, cmData)
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
