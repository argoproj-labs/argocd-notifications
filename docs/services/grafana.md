# Grafana

To be able to create Grafana annotation with argocd-notifications you have to create an [API Key](https://grafana.com/docs/grafana/latest/http_api/auth/#create-api-key) inside your [Grafana](https://grafana.com).

![sample](https://user-images.githubusercontent.com/958983/76374272-1cfe9500-6319-11ea-8477-b62d14ac3c9b.png)

1. Login to your Grafana instance as `admin`
2. On the left menu, go to Configuration / API Keys
3. Click "Add API Key" 
4. Fill the Key with name `ArgoCD Notification`, role `Editor` and Time to Live `10y` (for example)
5. Click on Add button
6. Copy your API Key and define it in `argocd-notifications-cm` ConfigMap

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: argocd-notifications-cm
data:
  service.grafana: |
    apiUrl: https://grafana.example.com/api
    apiKey: <grafana-api-key> 
```

7. Create subscription for your Grafana integration

```yaml
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  annotations:
    subscribe.argocd-notifications.argoproj.io: grafana:tag1|tag2 # list of tags separated with |
```