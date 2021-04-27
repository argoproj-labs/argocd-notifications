---
name: Reproducible bug report 
about: Create a reproducible bug report. Not for support requests.
labels: 'bug'
---
## Summary 

What happened/what you expected to happen?

## Diagnostics

What Kubernetes provider are you using? 

What version of Argo CD and Argo CD Notifications are you running? 

```
# Paste the logs from the notification controller

# Logs for the entire controller:
kubectl logs -n argocd deployment/argocd-notifications-controller
```

---
<!-- Issue Author: Don't delete this message to encourage other users to support your issue! -->
**Message from the maintainers**:

Impacted by this bug? Give it a 👍. We prioritise the issues with the most 👍.
