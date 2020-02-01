package testing

import (
	"time"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

const (
	TestNamespace = "default"
)

func WithAnnotations(annotations map[string]string) func(obj *unstructured.Unstructured) {
	return func(app *unstructured.Unstructured) {
		app.SetAnnotations(annotations)
	}
}

func WithProject(project string) func(app *unstructured.Unstructured) {
	return func(app *unstructured.Unstructured) {
		_ = unstructured.SetNestedField(app.Object, project, "spec", "project")
	}
}

func WithSyncStatus(status string) func(app *unstructured.Unstructured) {
	return func(app *unstructured.Unstructured) {
		_ = unstructured.SetNestedField(app.Object, status, "status", "sync", "status")
	}
}

func WithSyncOperationPhase(phase string) func(app *unstructured.Unstructured) {
	return func(app *unstructured.Unstructured) {
		_ = unstructured.SetNestedField(app.Object, phase, "status", "operationState", "phase")
	}
}

func WithSyncOperationStartAt(t time.Time) func(app *unstructured.Unstructured) {
	return func(app *unstructured.Unstructured) {
		ts := t.Format(time.RFC3339)
		_ = unstructured.SetNestedField(app.Object, ts, "status", "operationState", "startedAt")
	}
}

func WithHealthStatus(status string) func(app *unstructured.Unstructured) {
	return func(app *unstructured.Unstructured) {
		_ = unstructured.SetNestedField(app.Object, status, "status", "health", "status")
	}
}

func NewApp(name string, modifiers ...func(app *unstructured.Unstructured)) *unstructured.Unstructured {
	app := unstructured.Unstructured{}
	app.SetGroupVersionKind(schema.GroupVersionKind{Group: "argoproj.io", Kind: "application", Version: "v1alpha1"})
	app.SetName(name)
	app.SetNamespace(TestNamespace)
	for i := range modifiers {
		modifiers[i](&app)
	}
	return &app
}

func NewProject(name string, modifiers ...func(app *unstructured.Unstructured)) *unstructured.Unstructured {
	proj := unstructured.Unstructured{}
	proj.SetGroupVersionKind(schema.GroupVersionKind{Group: "argoproj.io", Kind: "appproject", Version: "v1alpha1"})
	proj.SetName(name)
	proj.SetNamespace(TestNamespace)
	for i := range modifiers {
		modifiers[i](&proj)
	}
	return &proj
}
