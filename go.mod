module github.com/argoproj-labs/argocd-notifications

go 1.13

require (
	github.com/Masterminds/goutils v1.1.0 // indirect
	github.com/Masterminds/sprig v2.22.0+incompatible
	github.com/antonmedv/expr v1.8.9
	github.com/argoproj/argo-cd v1.8.0
	github.com/argoproj/gitops-engine v0.2.1
	github.com/evanphx/json-patch v4.9.0+incompatible
	github.com/ghodss/yaml v1.0.0
	github.com/go-telegram-bot-api/telegram-bot-api v4.6.4+incompatible
	github.com/golang/mock v1.4.4
	github.com/grpc-ecosystem/go-grpc-middleware v1.2.0 // indirect
	github.com/huandu/xstrings v1.3.0 // indirect
	github.com/imdario/mergo v0.3.8 // indirect
	github.com/mitchellh/copystructure v1.0.0 // indirect
	github.com/olekukonko/tablewriter v0.0.4
	github.com/opsgenie/opsgenie-go-sdk-v2 v1.0.5
	github.com/prometheus/client_golang v1.7.1
	github.com/robfig/cron v1.2.0 // indirect
	github.com/sirupsen/logrus v1.6.0
	github.com/slack-go/slack v0.6.6
	github.com/spf13/cobra v1.0.0
	github.com/stretchr/testify v1.6.1
	github.com/technoweenie/multipartstreamer v1.0.1 // indirect
	github.com/whilp/git-urls v0.0.0-20191001220047-6db9661140c0
	gomodules.xyz/notify v0.1.0
	k8s.io/api v0.19.2
	k8s.io/apimachinery v0.19.2
	k8s.io/client-go v11.0.1-0.20190816222228-6d55c1b1f1ca+incompatible
	k8s.io/utils v0.0.0-20200729134348-d5654de09c73
)

replace (
	// https://github.com/golang/go/issues/33546#issuecomment-519656923
	github.com/go-check/check => github.com/go-check/check v0.0.0-20180628173108-788fd7840127

	k8s.io/api => k8s.io/api v0.19.2
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.19.2
	k8s.io/apimachinery => k8s.io/apimachinery v0.19.2
	k8s.io/apiserver => k8s.io/apiserver v0.19.2
	k8s.io/cli-runtime => k8s.io/cli-runtime v0.19.2
	k8s.io/client-go => k8s.io/client-go v0.19.2
	k8s.io/cloud-provider => k8s.io/cloud-provider v0.19.2
	k8s.io/cluster-bootstrap => k8s.io/cluster-bootstrap v0.19.2
	k8s.io/code-generator => k8s.io/code-generator v0.19.2
	k8s.io/component-base => k8s.io/component-base v0.19.2
	k8s.io/cri-api => k8s.io/cri-api v0.19.2
	k8s.io/csi-translation-lib => k8s.io/csi-translation-lib v0.19.2
	k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.19.2
	k8s.io/kube-controller-manager => k8s.io/kube-controller-manager v0.19.2
	k8s.io/kube-proxy => k8s.io/kube-proxy v0.19.2
	k8s.io/kube-scheduler => k8s.io/kube-scheduler v0.19.2
	k8s.io/kubectl => k8s.io/kubectl v0.19.2
	k8s.io/kubelet => k8s.io/kubelet v0.19.2
	k8s.io/legacy-cloud-providers => k8s.io/legacy-cloud-providers v0.19.2
	k8s.io/metrics => k8s.io/metrics v0.19.2
	k8s.io/sample-apiserver => k8s.io/sample-apiserver v0.19.2
)
