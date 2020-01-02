module github.com/rancher/rke

go 1.12

replace (
	github.com/go-resty/resty => gopkg.in/resty.v1 v1.9.0
	github.com/knative/pkg => github.com/rancher/pkg v0.0.0-20190514055449-b30ab9de040e
	k8s.io/api => k8s.io/api v0.0.0-20190918155943-95b840bb6a1f
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.0.0-20190918161926-8f644eb6e783
	k8s.io/apimachinery => k8s.io/apimachinery v0.0.0-20190913080033-27d36303b655
	k8s.io/apiserver => k8s.io/apiserver v0.0.0-20190918160949-bfa5e2e684ad
	k8s.io/cli-runtime => k8s.io/cli-runtime v0.0.0-20190918162238-f783a3654da8
	k8s.io/client-go => github.com/rancher/client-go v1.16.0-rancher.3
	k8s.io/cloud-provider => k8s.io/cloud-provider v0.0.0-20190918163234-a9c1f33e9fb9
	k8s.io/cluster-bootstrap => k8s.io/cluster-bootstrap v0.0.0-20190918163108-da9fdfce26bb
	k8s.io/code-generator => k8s.io/code-generator v0.0.0-20190912054826-cd179ad6a269
	k8s.io/component-base => k8s.io/component-base v0.0.0-20190918160511-547f6c5d7090
	k8s.io/cri-api => k8s.io/cri-api v0.0.0-20190828162817-608eb1dad4ac
	k8s.io/csi-translation-lib => k8s.io/csi-translation-lib v0.0.0-20190918163402-db86a8c7bb21
	k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.0.0-20190918161219-8c8f079fddc3
	k8s.io/kube-controller-manager => k8s.io/kube-controller-manager v0.0.0-20190918162944-7a93a0ddadd8
	k8s.io/kube-proxy => k8s.io/kube-proxy v0.0.0-20190918162534-de037b596c1e
	k8s.io/kube-scheduler => k8s.io/kube-scheduler v0.0.0-20190918162820-3b5c1246eb18
	k8s.io/kubectl => k8s.io/kubectl v0.0.0-20190918164019-21692a0861df
	k8s.io/kubelet => k8s.io/kubelet v0.0.0-20190918162654-250a1838aa2c
	k8s.io/kubernetes => k8s.io/kubernetes v1.16.0
	k8s.io/legacy-cloud-providers => k8s.io/legacy-cloud-providers v0.0.0-20190918163543-cfa506e53441
	k8s.io/metrics => k8s.io/metrics v0.0.0-20190918162108-227c654b2546
	k8s.io/sample-apiserver => k8s.io/sample-apiserver v0.0.0-20190918161442-d4c9c65c82af
)

require (
	github.com/Masterminds/sprig/v3 v3.0.0
	github.com/blang/semver v3.5.1+incompatible
	github.com/containerd/containerd v1.3.0-beta.0.0.20190808172034-23faecfb66ab // indirect
	github.com/coreos/bbolt v1.3.3 // indirect
	github.com/coreos/etcd v3.3.15+incompatible
	github.com/coreos/go-semver v0.3.0
	github.com/coreos/pkg v0.0.0-20180928190104-399ea9e2e55f // indirect
	github.com/docker/distribution v2.7.1+incompatible
	github.com/docker/docker v0.7.3-0.20190808172531-150530564a14
	github.com/docker/go-connections v0.3.0
	github.com/ghodss/yaml v1.0.0
	github.com/go-ini/ini v1.37.0
	github.com/google/btree v1.0.0 // indirect
	github.com/gorilla/mux v1.7.3 // indirect
	github.com/mattn/go-colorable v0.1.0
	github.com/mcuadros/go-version v0.0.0-20180611085657-6d5863ca60fa
	github.com/pkg/errors v0.8.1
	github.com/rancher/kontainer-driver-metadata v0.0.0-20200102184830-62373fc95553
	github.com/rancher/norman v0.0.0-20191209163739-5b9227fe3222
	github.com/rancher/types v0.0.0-20191220141525-65ed56d888c6
	github.com/sirupsen/logrus v1.4.2
	github.com/smartystreets/goconvey v0.0.0-20190731233626-505e41936337 // indirect
	github.com/stretchr/testify v1.4.0
	github.com/tmc/grpc-websocket-proxy v0.0.0-20190109142713-0ad062ec5ee5 // indirect
	github.com/urfave/cli v1.18.0
	github.com/xiang90/probing v0.0.0-20190116061207-43a291ad63a2 // indirect
	go.etcd.io/bbolt v1.3.3 // indirect
	golang.org/x/crypto v0.0.0-20190911031432-227b76d455e7
	golang.org/x/sync v0.0.0-20190423024810-112230192c58
	gopkg.in/ini.v1 v1.46.0 // indirect
	gopkg.in/yaml.v2 v2.2.4
	k8s.io/api v0.0.0
	k8s.io/apimachinery v0.0.0
	k8s.io/apiserver v0.0.0
	k8s.io/client-go v11.0.1-0.20190805182715-88a2adca7e76+incompatible
	k8s.io/kubernetes v1.16.0
	sigs.k8s.io/yaml v1.1.0
)
