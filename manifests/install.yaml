apiVersion: v1
kind: ServiceAccount
metadata:
  name: argocd-notifications-controller
---
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: argocd-notifications-controller
rules:
- apiGroups:
  - argoproj.io
  resources:
  - applications
  - appprojects
  verbs:
  - get
  - list
  - watch
  - update
  - patch
- apiGroups:
  - ""
  resources:
  - secrets
  - configmaps
  verbs:
  - get
  - list
  - watch
---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: argocd-notifications-controller
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: argocd-notifications-controller
subjects:
- kind: ServiceAccount
  name: argocd-notifications-controller
---
apiVersion: v1
kind: ConfigMap
metadata:
  creationTimestamp: null
  name: argocd-notifications-cm
---
apiVersion: v1
kind: Secret
metadata:
  name: argocd-notifications-secret
type: Opaque
---
apiVersion: v1
kind: Service
metadata:
  labels:
    app.kubernetes.io/name: argocd-notifications-controller-metrics
  name: argocd-notifications-controller-metrics
spec:
  ports:
  - name: metrics
    port: 9001
    protocol: TCP
    targetPort: 9001
  selector:
    app.kubernetes.io/name: argocd-notifications-controller
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: argocd-notifications-controller
spec:
  selector:
    matchLabels:
      app.kubernetes.io/name: argocd-notifications-controller
  strategy:
    type: Recreate
  template:
    metadata:
      labels:
        app.kubernetes.io/name: argocd-notifications-controller
    spec:
      containers:
      - command:
        - /app/argocd-notifications-backend
        - controller
        image: argoprojlabs/argocd-notifications:v1.0.2
        imagePullPolicy: Always
        name: argocd-notifications-controller
        volumeMounts:
        - mountPath: /app/config/tls
          name: tls-certs
        workingDir: /app
      securityContext:
        runAsNonRoot: true
      serviceAccountName: argocd-notifications-controller
      volumes:
      - configMap:
          name: argocd-tls-certs-cm
        name: tls-certs
