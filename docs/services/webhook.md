# Webhook

!!! note "Requires version v0.6+"

The webhook notification service allows sending a generic HTTP request using the templatized request body and URL.
Using Webhook you might trigger a Jenkins job, update Github commit status.

Use the following steps to configure webhook:

1 Register webhook in `argocd-notifications-secret` secret under `webhook` section in `notifiers.yaml` field:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: argocd-notifications-secret
stringData:
  notifiers.yaml: |
    webhook:
    - name: <webhook-name>
      url: https://<hostname>/<optional-path>
      headers: #optional headers
      - name: <header-name>
        value: <header-value>
      basicAuth: #optional username password
        username: <username>
        password: <api-key>
type: Opaque
```

2 Use template to customize request method, path and body:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: argocd-notifications-cm
data:
  config.yaml: |
    templates:
    - name: <template-name>
      webhook:
        <webhook-name>:
          method: POST # one of: GET, POST, PUT, PATCH. Default value: GET 
          path: <optional-path-template>
          body: |
            <optional-body-template>
```

3 Create application/project subscription:

```yaml
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  annotations:
    recipients.argocd-notifications.argoproj.io: webhook:<webhook-name>
  name: <my-app>
```

## Examples

### Set Github commit status

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: argocd-notifications-secret
stringData:
  notifiers.yaml: |
    webhook:
    - name: github
      url: https://api.github.com
      headers:
      - name: Authorization
        value: token <token>

type: Opaque
```

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: argocd-notifications-cm
data:
  config.yaml: |
    triggers:
      - name: sync-operation-failed
        condition: app.status.operationState.phase in ['Error', 'Failed']
        template: sync-operation-status-change
      - name: sync-operation-succeeded
        condition: app.status.operationState.phase in ['Succeeded']
        template: sync-operation-status-change
      - name: sync-operation-running
        condition: app.status.operationState.phase in ['Running']
        template: sync-operation-status-change

    templates:
      - name: sync-operation-status-change
        webhook:
          github:
            method: POST
            path: /repos/{{call .repo.FullNameByRepoURL .app.spec.source.repoURL}}/statuses/{{.app.status.operationState.operation.sync.revision}}
            body: |
              {
                {{if eq .app.status.operationState.phase "Running"}} "state": "pending"{{end}}
                {{if eq .app.status.operationState.phase "Succeeded"}} "state": "success"{{end}}
                {{if eq .app.status.operationState.phase "Error"}} "state": "error"{{end}}
                {{if eq .app.status.operationState.phase "Failed"}} "state": "error"{{end}},
                "description": "ArgoCD",
                "target_url": "{{.context.argocdUrl}}/applications/{{.app.metadata.name}}",
                "context": "continuous-delivery/{{.app.metadata.name}}"
              }
```

### Start Jenkins Job

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: argocd-notifications-secret
stringData:
  notifiers.yaml: |
    webhook:
    - name: jenkins
      url: http://<jenkins-host>/job/<job-name>/build?token=<job-secret>
      basicAuth:
        username: <username>
        password: <api-key>

type: Opaque
```
