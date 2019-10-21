module github.com/rancher/rke

go 1.12

replace (
	github.com/go-resty/resty => gopkg.in/resty.v1 v1.9.0
	k8s.io/client-go => k8s.io/client-go v0.0.0-20190918160344-1fbdaa4c8d90
)

require (
	github.com/Microsoft/go-winio v0.0.0-20180501170546-ab35fc04b636 // indirect
	github.com/blang/semver v3.5.1+incompatible
	github.com/containerd/containerd v1.3.0-beta.0.0.20190808172034-23faecfb66ab // indirect
	github.com/coreos/bbolt v1.3.3 // indirect
	github.com/coreos/etcd v3.3.15+incompatible
	github.com/coreos/go-semver v0.3.0
	github.com/coreos/pkg v0.0.0-20180928190104-399ea9e2e55f // indirect
	github.com/docker/distribution v2.7.1-0.20190205005809-0d3efadf0154+incompatible
	github.com/docker/docker v0.7.3-0.20190808172531-150530564a14
	github.com/docker/go-connections v0.3.0
	github.com/go-ini/ini v1.37.0
	github.com/google/btree v1.0.0 // indirect
	github.com/gorilla/mux v1.7.3 // indirect
	github.com/mattn/go-colorable v0.1.0
	github.com/mcuadros/go-version v0.0.0-20180611085657-6d5863ca60fa
	github.com/morikuni/aec v0.0.0-20170113033406-39771216ff4c // indirect
	github.com/opencontainers/go-digest v1.0.0-rc1 // indirect
	github.com/opencontainers/image-spec v0.0.0-20170929214853-7c889fafd04a // indirect
	github.com/pkg/errors v0.8.1
	github.com/rancher/kontainer-driver-metadata v0.0.0-20191021164950-6514452b9732
	github.com/rancher/norman v0.0.0-20191003174345-0ac7dd6ccb36
	github.com/rancher/types v0.0.0-20191003184925-ab3403a2c8ca
	github.com/sirupsen/logrus v1.4.2
	github.com/smartystreets/goconvey v0.0.0-20190731233626-505e41936337 // indirect
	github.com/tmc/grpc-websocket-proxy v0.0.0-20190109142713-0ad062ec5ee5 // indirect
	github.com/urfave/cli v1.18.0
	github.com/xiang90/probing v0.0.0-20190116061207-43a291ad63a2 // indirect
	go.etcd.io/bbolt v1.3.3 // indirect
	golang.org/x/crypto v0.0.0-20190611184440-5c40567a22f8
	golang.org/x/sync v0.0.0-20190423024810-112230192c58
	gopkg.in/ini.v1 v1.46.0 // indirect
	gopkg.in/yaml.v2 v2.2.4
	k8s.io/api v0.0.0-20190918155943-95b840bb6a1f
	k8s.io/apimachinery v0.0.0-20190913080033-27d36303b655
	k8s.io/client-go v11.0.1-0.20190409021438-1a26190bd76a+incompatible
)
