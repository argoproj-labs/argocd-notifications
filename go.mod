module github.com/argoproj-labs/argocd-notifications

go 1.13

require (
	github.com/Masterminds/goutils v1.1.0 // indirect
	github.com/Masterminds/semver v1.5.0 // indirect
	github.com/Masterminds/sprig v2.22.0+incompatible
	github.com/antonmedv/expr v1.4.1
	github.com/argoproj/argo-cd v1.5.4
	github.com/argoproj/pkg v0.0.0-20200424003221-9b858eff18a1 // indirect
	github.com/evanphx/json-patch v4.5.0+incompatible // indirect
	github.com/ghodss/yaml v1.0.0
	github.com/gobwas/glob v0.2.3 // indirect
	github.com/gogo/protobuf v1.3.1 // indirect
	github.com/golang/groupcache v0.0.0-20180513044358-24b0969c4cb7 // indirect
	github.com/golang/mock v1.3.1
	github.com/googleapis/gnostic v0.3.1 // indirect
	github.com/grpc-ecosystem/go-grpc-middleware v1.2.0 // indirect
	github.com/huandu/xstrings v1.3.0 // indirect
	github.com/imdario/mergo v0.3.8 // indirect
	github.com/jstemmer/go-junit-report v0.9.1 // indirect
	github.com/kballard/go-shellquote v0.0.0-20180428030007-95032a82bc51 // indirect
	github.com/konsorten/go-windows-terminal-sequences v1.0.2 // indirect
	github.com/magiconair/properties v1.8.0
	github.com/mitchellh/copystructure v1.0.0 // indirect
	github.com/olekukonko/tablewriter v0.0.4
	github.com/opsgenie/opsgenie-go-sdk-v2 v1.0.5
	github.com/prometheus/client_golang v1.6.0
	github.com/robfig/cron v1.2.0 // indirect
	github.com/sirupsen/logrus v1.4.2
	github.com/slack-go/slack v0.6.6
	github.com/spf13/cobra v0.0.5
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/stretchr/testify v1.4.0
	github.com/whilp/git-urls v0.0.0-20191001220047-6db9661140c0
	golang.org/x/net v0.0.0-20191004110552-13f9640d40b9 // indirect
	golang.org/x/oauth2 v0.0.0-20191202225959-858c2ad4c8b6 // indirect
	golang.org/x/time v0.0.0-20191024005414-555d28b269f0 // indirect
	gomodules.xyz/notify v0.1.0
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/src-d/go-git.v4 v4.13.1 // indirect
	k8s.io/api v0.0.0-20191114100352-16d7abae0d2a
	k8s.io/apimachinery v0.0.0-20191028221656-72ed19daf4bb
	k8s.io/client-go v0.0.0-20191114101535-6c5935290e33
	k8s.io/klog v1.0.0 // indirect
	k8s.io/utils v0.0.0-20191114200735-6ca3b61696b6
)

// https://github.com/golang/go/issues/33546#issuecomment-519656923
replace github.com/go-check/check => github.com/go-check/check v0.0.0-20180628173108-788fd7840127
