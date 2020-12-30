The notification template is used to generate the notification content and configured in `argocd-notifications-cm` ConfigMap. The template is leveraging
[html/template](https://golang.org/pkg/html/template/) golang package and allow to define notification title and body.
Templates are meant to be reusable and can be referenced by multiple triggers.

The following template is used to notify the user about application sync status.

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: argocd-notifications-cm
data:
  template.my-custom-template-slack-template: |
    title: Application {{.app.metadata.name}} sync status is {{.app.status.sync.status}}
    body: |
      Application {{.app.metadata.name}} sync is {{.app.status.sync.status}}.
      Application details: {{.context.argocdUrl}}/applications/{{.app.metadata.name}}.
```

Each template has access to the following fields:

- `app` holds the application object.
- `context` is user defined string map and might include any string keys and values.
- `serviceType` holds the notification service type name. The field can be used to conditionally
render service specific fields.


## Functions

Templates have access to the set of built-in functions:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: argocd-notifications-cm
data:
  template.my-custom-template-slack-template: |
    title: Application {{.app.metadata.name}} sync status is {{.app.status.sync.status}}
    body: "Author: {{(call .repo.GetCommitMetadata .app.status.sync.revision).Author}}"
```

{!functions.md!}

## Notification Service Specific Messages

Templates might define notification service-specific fields, for example, attachments for Slack or URL path and body for Webhook.
See corresponding service [documentation](./services/overview.md) for more information.
