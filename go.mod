module github.com/rancher/rke

go 1.12

replace (
	github.com/knative/pkg => github.com/rancher/pkg v0.0.0-20190514055449-b30ab9de040e
	k8s.io/client-go => k8s.io/client-go v0.17.2
)

require (
	github.com/Masterminds/sprig/v3 v3.0.0
	github.com/Microsoft/go-winio v0.4.11 // indirect
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
	github.com/go-bindata/go-bindata v3.1.2+incompatible
	github.com/go-ini/ini v1.37.0
	github.com/gorilla/mux v1.7.3 // indirect
	github.com/mattn/go-colorable v0.1.2
	github.com/mcuadros/go-version v0.0.0-20180611085657-6d5863ca60fa
	github.com/morikuni/aec v0.0.0-20170113033406-39771216ff4c // indirect
	github.com/opencontainers/image-spec v1.0.1 // indirect
	github.com/pkg/errors v0.8.1
	github.com/rancher/norman v0.0.0-20200211155126-fc45a55d4dfd
	github.com/rancher/types v0.0.0-20200226215232-12620b5bd7ff
	github.com/sirupsen/logrus v1.4.2
	github.com/stretchr/testify v1.4.0
	github.com/tmc/grpc-websocket-proxy v0.0.0-20190109142713-0ad062ec5ee5 // indirect
	github.com/urfave/cli v1.20.0
	golang.org/x/crypto v0.0.0-20191112222119-e1110fd1c708
	golang.org/x/sync v0.0.0-20190911185100-cd5d95a43a6e
	gopkg.in/yaml.v2 v2.2.5
	k8s.io/api v0.17.2
	k8s.io/apimachinery v0.17.2
	k8s.io/apiserver v0.17.2
	k8s.io/client-go v12.0.0+incompatible
	k8s.io/kubectl v0.17.2
	sigs.k8s.io/yaml v1.1.0
)
