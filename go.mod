module github.com/w6d-io/mongodb

go 1.15

require (
	github.com/avast/retry-go v3.0.0+incompatible
	github.com/go-logr/logr v0.3.0
	github.com/google/uuid v1.2.0
	github.com/jetstack/cert-manager v1.2.0
	github.com/onsi/ginkgo v1.14.1
	github.com/onsi/gomega v1.10.2
	go.uber.org/zap v1.16.0
	gopkg.in/yaml.v3 v3.0.0-20200615113413-eeeca48fe776
	k8s.io/api v0.19.7
	k8s.io/apimachinery v0.19.7
	k8s.io/client-go v0.19.7
	knative.dev/pkg v0.0.0-20210406170139-b8e331a6abf3 // indirect
	sigs.k8s.io/controller-runtime v0.7.2
)
