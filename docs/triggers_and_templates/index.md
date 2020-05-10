# Overview

Triggers and templates are configured in the `config.yaml` field of the `argocd-notification-cm` ConfigMap:

```yaml
{!argocd-notifications-cm.yaml!}
```

## Triggers

The trigger defines the condition when the notification should be sent. The definition includes name, condition
and notification template reference.

The following trigger sends a notification when application sync status changes to `Unknown`:

```
  - name: on-sync-status-unknown
    condition: app.status.sync.status == 'Unknown'
    template: app-sync-status
    enabled: true
```

* **name** - a unique trigger identifier.
* **template** - the name of the template that defines the notification content.
* **condition** - a predicate expression that returns true if the notification should be sent. The trigger condition
evaluation is powered by [antonmedv/expr](https://github.com/antonmedv/expr). The condition language syntax is described
at [Language-Definition.md](https://github.com/antonmedv/expr/blob/master/docs/Language-Definition.md).
* **enabled** - flag that indicates if trigger is enabled or not. By default trigger is enabled.

## Templates

The notification template is used to generate the notification content. The template is leveraging
[html/template](https://golang.org/pkg/html/template/) golang package and allow to define notification title and body.
The template is meant to be reusable and can be referenced by multiple triggers.

The following template is used to notify the user about application sync status.

```
  - name: app-sync-status
    title: Application {{.app.metadata.name}} sync status is {{.app.status.sync.status}}
    body: |
      Application {{.app.metadata.name}} sync is {{.app.status.sync.status}}.
      Application details: {{.context.argocdUrl}}/applications/{{.app.metadata.name}}.
```

Each template has access to the `app` and `context` fields:

- `app` holds the application object.
- `context` is user defined string map and might include any string keys and values.

