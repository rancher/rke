module github.com/rancher/rke

go 1.12

replace github.com/go-resty/resty => gopkg.in/resty.v1 v1.9.0

require (
	github.com/Azure/go-ansiterm v0.0.0-20170929234023-d6e3b3328b78 // indirect
	github.com/Microsoft/go-winio v0.0.0-20180501170546-ab35fc04b636 // indirect
	github.com/blang/semver v0.0.0-20190414102917-ba2c2ddd8906
	github.com/containerd/containerd v1.3.0-beta.0.0.20190808172034-23faecfb66ab // indirect
	github.com/coreos/bbolt v1.3.3 // indirect
	github.com/coreos/etcd v0.0.0-20180109221743-52f73c5a6cb0
	github.com/coreos/go-semver v0.3.0
	github.com/coreos/go-systemd v0.0.0-20161114122254-48702e0da86b // indirect
	github.com/coreos/pkg v0.0.0-20180928190104-399ea9e2e55f // indirect
	github.com/dgrijalva/jwt-go v3.2.0+incompatible // indirect
	github.com/docker/distribution v2.7.1-0.20190205005809-0d3efadf0154+incompatible
	github.com/docker/docker v0.7.3-0.20190808172531-150530564a14
	github.com/docker/go-connections v0.3.0
	github.com/docker/go-units v0.3.2 // indirect
	github.com/go-ini/ini v1.37.0
	github.com/google/btree v1.0.0 // indirect
	github.com/gorilla/websocket v1.2.0 // indirect
	github.com/grpc-ecosystem/go-grpc-prometheus v1.2.0 // indirect
	github.com/grpc-ecosystem/grpc-gateway v1.7.0 // indirect
	github.com/jonboulle/clockwork v0.1.0 // indirect
	github.com/mattn/go-colorable v0.1.0
	github.com/mattn/go-isatty v0.0.4 // indirect
	github.com/mcuadros/go-version v0.0.0-20180611085657-6d5863ca60fa
	github.com/morikuni/aec v0.0.0-20170113033406-39771216ff4c // indirect
	github.com/opencontainers/go-digest v1.0.0-rc1 // indirect
	github.com/opencontainers/image-spec v0.0.0-20170929214853-7c889fafd04a // indirect
	github.com/pkg/errors v0.8.1
	github.com/rancher/kontainer-driver-metadata v0.0.0-20190911170536-f9acf8fc853c
	github.com/rancher/norman v0.0.0-20190821234528-20a936b685b0
	github.com/rancher/types v0.0.0-20190827214052-704648244586
	github.com/sirupsen/logrus v1.4.2
	github.com/smartystreets/goconvey v0.0.0-20190731233626-505e41936337 // indirect
	github.com/soheilhy/cmux v0.1.4 // indirect
	github.com/tmc/grpc-websocket-proxy v0.0.0-20190109142713-0ad062ec5ee5 // indirect
	github.com/ugorji/go v0.0.0-20171231121548-ccfe18359b55 // indirect
	github.com/urfave/cli v1.18.0
	github.com/xiang90/probing v0.0.0-20190116061207-43a291ad63a2 // indirect
	go.etcd.io/bbolt v1.3.3 // indirect
	golang.org/x/crypto v0.0.0-20190605123033-f99c8df09eb5
	golang.org/x/sync v0.0.0-20190423024810-112230192c58
	google.golang.org/grpc v1.2.1-0.20190807214610-36ddeccf1860 // indirect
	gopkg.in/check.v1 v1.0.0-20180628173108-788fd7840127 // indirect
	gopkg.in/ini.v1 v1.46.0 // indirect
	gopkg.in/yaml.v2 v2.2.2
	gotest.tools v2.2.0+incompatible // indirect
	k8s.io/api v0.0.0-20190805182251-6c9aa3caf3d6
	k8s.io/apimachinery v0.0.0-20190404173353-6a84e37a896d
	k8s.io/client-go v11.0.1-0.20190805182715-88a2adca7e76+incompatible
)
