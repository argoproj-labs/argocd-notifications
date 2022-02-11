module github.com/argoproj-labs/argocd-notifications

go 1.16

require (
	github.com/argoproj/argo-cd/v2 v2.1.9
	github.com/argoproj/notifications-engine v0.3.1-0.20211117165611-0e1f1eda5f52
	github.com/evanphx/json-patch v4.11.0+incompatible
	github.com/ghodss/yaml v1.0.0
	github.com/go-redis/cache/v8 v8.11.3 // indirect
	github.com/golang/mock v1.5.0
	github.com/grpc-ecosystem/go-grpc-middleware v1.2.0 // indirect
	github.com/olekukonko/tablewriter v0.0.4
	github.com/prometheus/client_golang v1.11.0
	github.com/robfig/cron v1.2.0 // indirect
	github.com/sirupsen/logrus v1.8.1
	github.com/slack-go/slack v0.6.6
	github.com/spf13/cobra v1.1.3
	github.com/stretchr/testify v1.7.0
	github.com/whilp/git-urls v0.0.0-20191001220047-6db9661140c0
	k8s.io/api v0.21.0
	k8s.io/apimachinery v0.21.0
	k8s.io/client-go v11.0.1-0.20190816222228-6d55c1b1f1ca+incompatible
	k8s.io/utils v0.0.0-20210111153108-fddb29f9d009
)

replace (
	// https://github.com/golang/go/issues/33546#issuecomment-519656923
	github.com/go-check/check => github.com/go-check/check v0.0.0-20180628173108-788fd7840127

	github.com/go-redis/cache/v8 => github.com/go-redis/cache/v8 v8.4.3

	k8s.io/api => k8s.io/api v0.21.0
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.21.0
	k8s.io/apimachinery => k8s.io/apimachinery v0.21.0
	k8s.io/apiserver => k8s.io/apiserver v0.21.0
	k8s.io/cli-runtime => k8s.io/cli-runtime v0.21.0
	k8s.io/client-go => k8s.io/client-go v0.21.0
	k8s.io/cloud-provider => k8s.io/cloud-provider v0.21.0
	k8s.io/cluster-bootstrap => k8s.io/cluster-bootstrap v0.21.0
	k8s.io/code-generator => k8s.io/code-generator v0.21.0
	k8s.io/component-base => k8s.io/component-base v0.21.0
	k8s.io/component-helpers => k8s.io/component-helpers v0.21.0
	k8s.io/controller-manager => k8s.io/controller-manager v0.21.0
	k8s.io/cri-api => k8s.io/cri-api v0.21.0
	k8s.io/csi-translation-lib => k8s.io/csi-translation-lib v0.21.0
	k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.21.0
	k8s.io/kube-controller-manager => k8s.io/kube-controller-manager v0.21.0
	k8s.io/kube-proxy => k8s.io/kube-proxy v0.21.0
	k8s.io/kube-scheduler => k8s.io/kube-scheduler v0.21.0
	k8s.io/kubectl => k8s.io/kubectl v0.21.0
	k8s.io/kubelet => k8s.io/kubelet v0.21.0
	k8s.io/legacy-cloud-providers => k8s.io/legacy-cloud-providers v0.21.0
	k8s.io/metrics => k8s.io/metrics v0.21.0
	k8s.io/mount-utils => k8s.io/mount-utils v0.21.0
	k8s.io/sample-apiserver => k8s.io/sample-apiserver v0.21.0
)
