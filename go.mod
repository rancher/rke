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
	github.com/go-ini/ini v1.37.0
	github.com/gorilla/mux v1.7.3 // indirect
	github.com/mattn/go-colorable v0.1.0
	github.com/mcuadros/go-version v0.0.0-20180611085657-6d5863ca60fa
	github.com/morikuni/aec v0.0.0-20170113033406-39771216ff4c // indirect
	github.com/opencontainers/image-spec v1.0.1 // indirect
	github.com/pkg/errors v0.8.1
	github.com/rancher/kontainer-driver-metadata v0.0.0-20200218183853-2f58a2f30054
	github.com/rancher/norman v0.0.0-20200211155126-fc45a55d4dfd
	github.com/rancher/types v0.0.0-20200218191331-dc762fc27c91
	github.com/sirupsen/logrus v1.4.2
	github.com/smartystreets/goconvey v0.0.0-20190731233626-505e41936337 // indirect
	github.com/stretchr/testify v1.4.0
	github.com/tmc/grpc-websocket-proxy v0.0.0-20190109142713-0ad062ec5ee5 // indirect
	github.com/urfave/cli v1.20.0
	golang.org/x/crypto v0.0.0-20190911031432-227b76d455e7
	golang.org/x/sync v0.0.0-20190911185100-cd5d95a43a6e
	gopkg.in/ini.v1 v1.46.0 // indirect
	gopkg.in/yaml.v2 v2.2.5
	k8s.io/api v0.17.2
	k8s.io/apimachinery v0.17.2
	k8s.io/apiserver v0.17.2
	k8s.io/client-go v2.0.0-alpha.0.0.20181121191925-a47917edff34+incompatible
	k8s.io/kubectl v0.17.2
	sigs.k8s.io/yaml v1.1.0
)
