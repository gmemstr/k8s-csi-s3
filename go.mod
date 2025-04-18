module git.gmem.ca/arch/k8s-csi-s3

go 1.23.1

toolchain go1.24.1

require (
	github.com/container-storage-interface/spec v1.11.0
	github.com/coreos/go-systemd/v22 v22.5.0
	github.com/godbus/dbus/v5 v5.1.0
	github.com/golang/glog v1.2.4
	github.com/kubernetes-csi/csi-test v2.0.0+incompatible
	github.com/minio/minio-go/v7 v7.0.88
	github.com/mitchellh/go-ps v1.0.0
	github.com/onsi/ginkgo v1.16.4
	github.com/onsi/gomega v1.35.1
	golang.org/x/net v0.37.0
	google.golang.org/grpc v1.71.0
	google.golang.org/protobuf v1.36.5
	k8s.io/mount-utils v0.0.0
)

require (
	github.com/dustin/go-humanize v1.0.1 // indirect
	github.com/fsnotify/fsnotify v1.7.0 // indirect
	github.com/go-ini/ini v1.67.0 // indirect
	github.com/go-logr/logr v1.4.2 // indirect
	github.com/goccy/go-json v0.10.5 // indirect
	github.com/google/go-cmp v0.6.0 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/klauspost/compress v1.17.11 // indirect
	github.com/klauspost/cpuid/v2 v2.2.9 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/minio/crc64nvme v1.0.1 // indirect
	github.com/minio/md5-simd v1.1.2 // indirect
	github.com/moby/sys/mountinfo v0.7.2 // indirect
	github.com/moby/sys/userns v0.1.0 // indirect
	github.com/nxadm/tail v1.4.8 // indirect
	github.com/onsi/ginkgo/v2 v2.21.0 // indirect
	github.com/rs/xid v1.6.0 // indirect
	golang.org/x/crypto v0.36.0 // indirect
	golang.org/x/sys v0.31.0 // indirect
	golang.org/x/text v0.23.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250115164207-1a7da9e5054f // indirect
	gopkg.in/tomb.v1 v1.0.0-20141024135613-dd632973f1e7 // indirect
	gopkg.in/yaml.v2 v2.3.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	k8s.io/klog/v2 v2.130.1 // indirect
	k8s.io/utils v0.0.0-20241104100929-3ea5e8cea738 // indirect
)

replace k8s.io/api => k8s.io/api v0.32.3

replace k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.32.3

replace k8s.io/apimachinery => k8s.io/apimachinery v0.32.3

replace k8s.io/apiserver => k8s.io/apiserver v0.32.3

replace k8s.io/cli-runtime => k8s.io/cli-runtime v0.32.3

replace k8s.io/client-go => k8s.io/client-go v0.32.3

replace k8s.io/cloud-provider => k8s.io/cloud-provider v0.32.3

replace k8s.io/cluster-bootstrap => k8s.io/cluster-bootstrap v0.32.3

replace k8s.io/code-generator => k8s.io/code-generator v0.32.3

replace k8s.io/component-base => k8s.io/component-base v0.32.3

replace k8s.io/component-helpers => k8s.io/component-helpers v0.32.3

replace k8s.io/controller-manager => k8s.io/controller-manager v0.32.3

replace k8s.io/cri-api => k8s.io/cri-api v0.32.3

replace k8s.io/cri-client => k8s.io/cri-client v0.32.3

replace k8s.io/csi-translation-lib => k8s.io/csi-translation-lib v0.32.3

replace k8s.io/dynamic-resource-allocation => k8s.io/dynamic-resource-allocation v0.32.3

replace k8s.io/endpointslice => k8s.io/endpointslice v0.32.3

replace k8s.io/externaljwt => k8s.io/externaljwt v0.32.3

replace k8s.io/kms => k8s.io/kms v0.32.3

replace k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.32.3

replace k8s.io/kube-controller-manager => k8s.io/kube-controller-manager v0.32.3

replace k8s.io/kube-proxy => k8s.io/kube-proxy v0.32.3

replace k8s.io/kube-scheduler => k8s.io/kube-scheduler v0.32.3

replace k8s.io/kubectl => k8s.io/kubectl v0.32.3

replace k8s.io/kubelet => k8s.io/kubelet v0.32.3

replace k8s.io/metrics => k8s.io/metrics v0.32.3

replace k8s.io/mount-utils => k8s.io/mount-utils v0.32.3

replace k8s.io/pod-security-admission => k8s.io/pod-security-admission v0.32.3

replace k8s.io/sample-apiserver => k8s.io/sample-apiserver v0.32.3

replace k8s.io/sample-cli-plugin => k8s.io/sample-cli-plugin v0.32.3

replace k8s.io/sample-controller => k8s.io/sample-controller v0.32.3
