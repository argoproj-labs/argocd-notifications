# Grafana

!!! note "Requires version v0.6+"

To be able to create Grafana annotation with argocd-notifications you have to create an [API Key](https://grafana.com/docs/grafana/latest/http_api/auth/#create-api-key) inside your [Grafana](https://grafana.com).

1. Login to your Grafana instance as `admin`
2. On the left menu, go to Configuration / API Keys
3. Click "Add API Key" 
4. Fill the Key with name `ArgoCD Notification`, role `Editor` and Time to Live `10y` (for example)
5. Click on Add button
6. Copy your API Key and define it in `notifiers.yaml`

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: argocd-notifications-secret
stringData:
  notifiers.yaml: |
    grafana:
      apiUrl: https://grafana.example.com/api
      apiKey: <grafana-api-key> 
```
