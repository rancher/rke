module github.com/rancher/rke

go 1.16

replace (
	github.com/knative/pkg => github.com/rancher/pkg v0.0.0-20190514055449-b30ab9de040e
	k8s.io/client-go => k8s.io/client-go v0.22.3
)

require (
	github.com/Masterminds/sprig/v3 v3.2.2
	github.com/Microsoft/hcsshim v0.8.9 // indirect
	github.com/apparentlymart/go-cidr v1.0.1
	github.com/aws/aws-sdk-go v1.38.65
	github.com/blang/semver v3.5.1+incompatible
	github.com/containerd/containerd v1.4.1-0.20201117152358-0edc412565dc // indirect
	github.com/containerd/continuity v0.0.0-20200710164510-efbc4488d8fe // indirect
	github.com/coreos/go-semver v0.3.0
	github.com/docker/distribution v2.7.1+incompatible
	github.com/docker/docker v20.10.6+incompatible
	github.com/docker/go-connections v0.4.0
	github.com/ghodss/yaml v1.0.0
	github.com/go-bindata/go-bindata v3.1.2+incompatible
	github.com/go-ini/ini v1.37.0
	github.com/gopherjs/gopherjs v0.0.0-20191106031601-ce3c9ade29de // indirect
	github.com/gorilla/mux v1.7.3 // indirect
	github.com/mattn/go-colorable v0.1.8
	github.com/mcuadros/go-version v0.0.0-20180611085657-6d5863ca60fa
	github.com/moby/sys/mount v0.2.0 // indirect
	github.com/morikuni/aec v0.0.0-20170113033406-39771216ff4c // indirect
	github.com/opencontainers/image-spec v1.0.1 // indirect
	github.com/opencontainers/runc v0.1.1 // indirect
	github.com/pkg/errors v0.9.1
	github.com/rancher/norman v0.0.0-20200517050325-f53cae161640
	github.com/sirupsen/logrus v1.8.1
	github.com/smartystreets/assertions v1.0.1 // indirect
	github.com/stretchr/testify v1.7.0
	github.com/urfave/cli v1.20.0
	go.etcd.io/etcd/client/v2 v2.305.1
	go.etcd.io/etcd/client/v3 v3.5.1
	golang.org/x/crypto v0.0.0-20210220033148-5ea612d1eb83
	golang.org/x/sync v0.0.0-20210220032951-036812b2e83c
	google.golang.org/grpc v1.38.0
	gopkg.in/yaml.v2 v2.4.0
	k8s.io/api v0.22.3
	k8s.io/apimachinery v0.22.3
	k8s.io/apiserver v0.22.3
	k8s.io/client-go v0.22.3
	k8s.io/gengo v0.0.0-20201214224949-b6c5ce23f027
	k8s.io/kubectl v0.22.3
	sigs.k8s.io/yaml v1.2.0
)
