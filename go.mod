module toyou-csi

go 1.21.1

require (
	github.com/container-storage-interface/spec v1.9.0
	github.com/golang/protobuf v1.5.3
	github.com/kubernetes-csi/csi-lib-utils v0.17.0
	github.com/kubernetes-csi/csi-test v1.1.1
	github.com/kubernetes-csi/drivers v1.0.2
	github.com/stretchr/testify v1.8.4
	google.golang.org/grpc v1.59.0
	k8s.io/apimachinery v0.29.0
	k8s.io/klog v1.0.0
	k8s.io/klog/v2 v2.110.1
	k8s.io/kubernetes v1.29.0
	k8s.io/utils v0.0.0-20231127182322-b307cd553661
)

require (
	github.com/golang/glog v1.1.2 // indirect
	github.com/nxadm/tail v1.4.8 // indirect
	github.com/onsi/ginkgo v1.16.5 // indirect
	github.com/onsi/gomega v1.29.0 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	gopkg.in/tomb.v1 v1.0.0-20141024135613-dd632973f1e7 // indirect
)

require (
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/blang/semver/v4 v4.0.0 // indirect
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/emicklei/go-restful/v3 v3.11.0 // indirect
	github.com/fsnotify/fsnotify v1.7.0 // indirect
	github.com/go-logr/logr v1.3.0 // indirect
	github.com/go-openapi/jsonpointer v0.19.6 // indirect
	github.com/go-openapi/jsonreference v0.20.2 // indirect
	github.com/go-openapi/swag v0.22.3 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/google/gnostic-models v0.6.8 // indirect
	github.com/google/go-cmp v0.6.0 // indirect
	github.com/google/gofuzz v1.2.0 // indirect
	github.com/google/uuid v1.3.1 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.4 // indirect
	github.com/moby/sys/mountinfo v0.6.2 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/opencontainers/selinux v1.11.0 // indirect
	github.com/prometheus/client_golang v1.16.0 // indirect
	github.com/prometheus/client_model v0.4.0 // indirect
	github.com/prometheus/common v0.44.0 // indirect
	github.com/prometheus/procfs v0.10.1 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	golang.org/x/net v0.18.0
	golang.org/x/oauth2 v0.11.0 // indirect
	golang.org/x/sys v0.14.0 // indirect
	golang.org/x/term v0.14.0 // indirect
	golang.org/x/text v0.14.0 // indirect
	golang.org/x/time v0.3.0 // indirect
	google.golang.org/appengine v1.6.7 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20231106174013-bbf56f31fb17 // indirect
	google.golang.org/protobuf v1.31.0 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/yaml.v2 v2.4.0
	gopkg.in/yaml.v3 v3.0.1 // indirect
	k8s.io/api v0.29.0 // indirect
	k8s.io/apiextensions-apiserver v0.29.0 //indirect
	k8s.io/apiserver v0.29.0 //indirect
	k8s.io/client-go v0.29.0 // indirect
	k8s.io/cloud-provider v0.29.0 //indirect
	k8s.io/component-base v0.29.0 // indirect
	k8s.io/csi-translation-lib v0.29.0 //indirect
	k8s.io/kube-openapi v0.0.0-20231010175941-2dd684a91f00 // indirect
	k8s.io/mount-utils v0.29.0 //indirect
	sigs.k8s.io/json v0.0.0-20221116044647-bc3834ca7abd // indirect
	sigs.k8s.io/structured-merge-diff/v4 v4.4.1 // indirect
	sigs.k8s.io/yaml v1.3.0 // indirect
)

replace (
	k8s.io/api => k8s.io/api v0.29.0
	k8s.io/apimachinery => k8s.io/apimachinery v0.29.0
	k8s.io/cli-runtime => k8s.io/cli-runtime v0.29.0
	k8s.io/client-go => k8s.io/client-go v0.29.0
	k8s.io/cluster-bootstrap => k8s.io/cluster-bootstrap v0.29.0
	k8s.io/cri-api => k8s.io/cri-api v0.29.0
	k8s.io/dynamic-resource-allocation => k8s.io/dynamic-resource-allocation v0.29.0
	k8s.io/endpointslice => k8s.io/endpointslice v0.29.0
	k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.29.0
	k8s.io/kube-controller-manager => k8s.io/kube-controller-manager v0.29.0
	k8s.io/kube-proxy => k8s.io/kube-proxy v0.29.0
	k8s.io/kube-scheduler => k8s.io/kube-scheduler v0.29.0
	k8s.io/kubectl => k8s.io/kubectl v0.29.0
	k8s.io/kubelet => k8s.io/kubelet v0.29.0
	k8s.io/legacy-cloud-providers => k8s.io/legacy-cloud-providers v0.29.0
	k8s.io/metrics => k8s.io/metrics v0.29.0
	k8s.io/pod-security-admission => k8s.io/pod-security-admission v0.29.0
	k8s.io/sample-apiserver => k8s.io/sample-apiserver v0.29.0
)
