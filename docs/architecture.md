# Architecture

## Authors

* Mohamed Elsayed ([@moelsayed](https://github.com/moelsayed))
* Hussein Galal ([@galal-hussein](https://github.com/galal-hussein))
* Sebastiaan Van Steenis ([@superseb](https://github.com/superseb)), creator of this document

RKE was created to depend solely on Docker/the Docker socket. The requirements are:
* Supported OS + SSH daemon
* Docker daemon
* OS user with access to Docker socket
* A host to run `rke` binary with a `cluster.yml` configuration file

Everything that RKE does is done through the Docker socket. Any functionality that RKE contains, wether it is the TCP port checking or just running the Kubernetes containers, it is all run in Docker containers.

```
RKE binary -> ssh -> user@Supported OS -> Docker socket
```
## Basics

1. Create cluster configuration file (default `cluster.yml`)
2. Create the cluster (by running `rke up`)
3. When finished, use kubeconfig file (default `kube_config_cluster.yml`) to connect to your cluster
4. State is saved in state file (default `cluster.rkestate`), this file is necessary for every next interaction with the cluster using `rke`

## Paths

* `/etc/kubernetes(/ssl)`: All Kubernetes files like certificates, kubeconfig files and configuration files like audit policy/admission policy
* `/var/lib/rancher/rke/log`: Contains symbolic links to each k8s Docker container log file (not pods, that lives in `/var/log/pods`)
* `/opt/rke/`: etcd snapshots

## Create cluster configuration file

- Create `cluster.yml` yourself
- Use `rke config` to create `cluster.yml` interactively

See https://rke.docs.rancher.com/example-yamls

Minimal `cluster.yml`:

```
nodes:
- address: 1.2.3.4
  user: ubuntu
  role:
  - controlplane
  - etcd
  - worker
```

## Create the cluster

Run `rke up` to create the cluster. After any change in the cluster configuration file, you will need to run `rke up` again to apply the changes. The following flow is regarding new clusters, we go into detail for existing clusters below.


* Generate certificates needed for Kubernetes containers
* Create cluster state file and save locally
* Setup tunnel for node(s) (SSH to Docker socket)
* Network checks regarding open ports are performed using listening and check containers on nodes, TCP port checks are done for all components used (TCP/2379 for etcd, TCP/10250 for kubelet, TCP/6443 for kube-apiserver)
* Certificates are deployed to the nodes (container `cert-deployer` using `rancher/rke-tools` image)
* Create kubeconfig file and save locally
* Deploy files to hosts, for example `/etc/kubernetes/admission.yaml` and `/etc/kubernetes/audit-policy.yaml`
* Pulls `rancher/hyperkube` image as it is huge and might get interrupted if pulling is part of running the container
* Gets default options for k8s version that is configured from KDM (see https://github.com/rancher/kontainer-driver-metadata/blob/dev-v2.7/pkg/rke/k8s_service_options.go)
* Creates and starts `etcd-fix-perm` to set correct permissions on etcd data directory (only etcd role)
* Creates and starts `etcd` container, run `rke-log-linker` to link log files (only etcd role)
* Creates and starts `etcd-rolling-snapshots` container (this runs etcd snapshot on a configured interval, default `--retention=72h --creation=12h`), uses `rancher/rke-tools` container, see `main.go` for implementation. Run `rke-log-linker` to link log files (only etcd role)
* Creates and start `rke-bundle-cert` container (Saves a bundle of certificates in an archive, is legacy)
* Creates `service-sidekick` (uses `rancher/rke-tools` container, specifies `VOLUME /opt/rke-tools` in `Dockerfile` and that volume is re-used for k8s containers)
* Create and start k8s containers (depending on role, `etcd` components first, `controlplane` components second)
* Create `rke-job-deployer` ServiceAccount to deploy add-ons
* Create `system:node` ClusterRoleBinding (Node Authorization: https://kubernetes.io/docs/reference/access-authn-authz/node/)
* Create kube-apiserver proxy ClusterRole (`proxy-clusterrole-kubeapiserver`) and ClusterRoleBinding (`proxy-role-binding-kubernetes-master`)
* Updates `cluster.rkestate` file and saves it as ConfigMap to Kubernetes cluster (`full-cluster-state` in `kube-system`)
* Create and start k8s containers (`worker` components)
* Deploy CNI, save templated YAML as ConfigMap and deploy using Job `rke-network-plugin`
* Deploy CoreDNS, save templated YAML as ConfigMap and deploy using Job `rke-coredns-addon`
* Deploy Metrics Server, save templated YAML as ConfigMap and deploy using Job `rke-metrics-addon`
* Deploy Ingress Controller, save templated YAML as ConfigMap and deploy using Job `rke-ingress-controller`
* Deploy optional specified user addons
* Done


### Docker containers inspect output for learning/debugging:

Network port checks (`rke-etcd-port-listener`)

Host port is mapped to TCP/1337, and inside the listener, netcat (`nc`) is listening on TCP/1337 to verify connectivity

```
            "Image": "rancher/rke-tools:v0.1.89",
            "Cmd": [
                "nc",
                "-kl",
                "-p",
                "1337",
                "-e",
                "echo"
            ],
            "Ports": {
                "1337/tcp": [
                    {
                        "HostIp": "0.0.0.0",
                        "HostPort": "2380"
                    },
                    {
                        "HostIp": "0.0.0.0",
                        "HostPort": "2379"
                    }
                ],
                "80/tcp": null
            },
            
```

Network port checks (`rke-port-checker`)

Environment variables for all relevant nodes for that port are passed to the `rke-port-checker` and are checked using netcat (`nc`)

```
            "Image": "rancher/rke-tools:v0.1.89",
            "Cmd": [
                "sh",
                "-c",
                "for host in $HOSTS; do for port in $PORTS ; do echo \"Checking host ${host} on port ${port}\" >&1 & nc -w 5 -z $host $port > /dev/null || echo \"${host}:${port}\" >&2 & done; wait; done"
            ],
            "Env": [
                "HOSTS=172.26.14.216",
                "PORTS=2379",
```

Certificate deployer (`cert-deployer`)

`cert-deployer` is a binary built in the `rancher/rke-tools` repository that uses environment variables starting with `KUBE_` and translates the value of the environment variables into separate files.

```
            "Image": "rancher/rke-tools:v0.1.89",
            "Cmd": [
                "cert-deployer"
            ],
            "Env": [
                "KUBE_APISERVER_PROXY_CLIENT_KEY=-----BEGIN RSA PRIVATE KEY-----\nMIIEpAIBAAKCAQEAp3EAQWF9EQlvM3R2GEU5OC/wiYeF8zLf0qiCiMCxNVUaYx1+\nqtk6kUFyKdlgaRBqjKPr1u/ELkxA8KU8ffT9O2NAGpmuGgis9Ef+IszXbtqFMSXo\nJenLUqQDh/KGoG/dCBycMcvjhmu2d3EtY5Evl74UZhK1wbO3nS3rMm0ozxH6mq2R\nsWadaY02rQUegUN3D/7Q7LsYjvATVIcLLWk4lc
XrUAFjganggXkGUtoQdflfgEXA\n2IKzQnm8bS2l5GOGu414ss4n6agbEyOs8+z7sAMHSF/F3gaxqsGAZjBhCJyulDhe\nWqUp4sIrDlXLYeEr72VQj7hBR7dyVzJtWji+pQIDAQABAoIBAB9BF4Qct2SbtzcK\nkRScrz6OrD5vnpAzudWvgJYYKbvDw+YmVkN7wtPkPHQVUEqsNsdDvbzkCmF9+E0y\n+qSkOzR/pTKR5w6S+f2rBoPmanwVq/Dtm3SgPESTutkAayK9Xqup83nUDgdESc3n\nwUopipGveE1JRsX+TtK3BIToHU
rxSd9x5j556tFEYwIsR2THsBuva+f1kbPwRJp/\n9Ab4jHezo09KjfxFggrA/3Ve/R7PyYjFtX5fwGjb9icIRm/pJqmjwgv0WanAXVrg\nr1BGWGmM+9gblbbRyjFEZc1P57xBZfinPrZuTzMo/LIle1oIuISDb3gxawEjSfls\n6UNu+MECgYEA05c9EQVyKrnqdReXfelYxZMqxwtoePyLg/KnzcP4PjdkrjE5kQ4n\nyzf9XJyh8YTD7LvU8l48zQ6Bu6LWbVV0otxDNGijTAvPqsdOldP+LAXq1IzWyGw6\nyeohjpW4zP3jLr
wEXhY2q5M/s/o4Ho/RlF+GQswmkOOsAU65DWmGZ3UCgYEAypWe\njGzAgEOLJQ8XgDfW1f9WwddqEZJPSXL6ieF3b23XuikZtCGxtNmV5qS8XoXK3B+C\nhrI/ZTzwTTyikblxzCU4JomcGlTgroHSvhTJMOS5Q9lTgRvWuWs68w4abmQPJhvF\nmnwznM/0pVhq/Wup4BG0Ee+qJgfgREwSG0WCRHECgYBbiiW4NHP0+iPt7nvy1D48\nk/PA0zWqig/N0PA5/Btsx0g+eDtgfxBGQf3R0E3bkEW3KHfzN0P0rt7/j25XNM5W\nGx
bUGKT1JHL+fmWIOoPPBexXcmsFoJU6f5lu92VRAlIECQGWtuOGDRlVQt5+klfo\naf9K7MmOi4EBu84heFLWdQKBgQCHPaENr+BHAFBY2h1fPGfQjth1KYCm4FzL9NUq\nzPj1y4eTwLJnHYNL72HyCpGyLHFDyElT8JT/2dG2Tj9dN0av+TzmBUHQFk+0T/jH\naorxeA/yKphjfZk4SUyeTBD7FxNB5pJhUn8GNZHl/APY0FIkwszKmIunPeTK01nX\nGO0hEQKBgQDHB6B0hK1Ob48nvqkWYLDmPIsRFCk+HWjVKNjUvrPqsPqo
O+EZt9ae\nVb4+Onb3pp51Y/PCgaHSSnWDa57qGpNdTGGp/W6dJRs2oH9ai6UZ7QI9NfOD7B8A\nVMCO9zsPvcpffZHhuJtU6JcqPXjxy+9+M86eng+wZQsyZWx1hJTKWw==\n-----END RSA PRIVATE KEY-----\n",
                "KUBE_APISERVER_PROXY_CLIENT=-----BEGIN CERTIFICATE-----\nMIIDJDCCAgygAwIBAgIIF9ss+F3ZQSYwDQYJKoZIhvcNAQELBQAwKjEoMCYGA1UE\nAxMfa3ViZS1hcGlzZXJ2ZXItcmVxdWVzdGhlYWRlci1jYTAeFw0yMzA4MDIxNDQw\nMzBaFw0zMzA3MzAxNDQwMzFaMCYxJDAiBgNVBAMTG2t1YmUtYXBpc2VydmVyLXBy\nb3h5LWNsaWVudDCCASIwDQYJKoZIhvcNAQEBBQADggEPAD
CCAQoCggEBAKdxAEFh\nfREJbzN0dhhFOTgv8ImHhfMy39KogojAsTVVGmMdfqrZOpFBcinZYGkQaoyj69bv\nxC5MQPClPH30/TtjQBqZrhoIrPRH/iLM127ahTEl6CXpy1KkA4fyhqBv3QgcnDHL\n44ZrtndxLWORL5e+FGYStcGzt50t6zJtKM8R+pqtkbFmnWmNNq0FHoFDdw/+0Oy7\nGI7wE1SHCy1pOJXF61ABY4Gp4IF5BlLaEHX5X4BFwNiCs0J5vG0tpeRjhruNeLLO\nJ+moGxMjrPPs+7ADB0hfxd4GsarBgGYwYQ
icrpQ4XlqlKeLCKw5Vy2HhK+9lUI+4\nQUe3clcybVo4vqUCAwEAAaNSMFAwDgYDVR0PAQH/BAQDAgWgMB0GA1UdJQQWMBQG\nCCsGAQUFBwMCBggrBgEFBQcDATAfBgNVHSMEGDAWgBRR1ePFAu2FarZUO9SwV9lu\nIgnxuzANBgkqhkiG9w0BAQsFAAOCAQEAmhcGo60ALsJ1I/DO5+EcSGfmo4UupCb3\nm+OCSF3hSsU7HR+w1mutRL+xCQvvA/2xtVxYQlqzmAIwgPCOI/uY3SSbzzi98XIq\nRssUuv+plb11tzQBgx+Kv4
wtZ5pPnkFXOrMUbkEl7xXvbrTjSdos9e+tzlvZjzLY\nZcQ7cNbhrgVLEF+vYR1UuqvjkZiCe/V/4JqWrH+OLwlP0eLuRI8yndXtycvGB62a\ne6tecWByY3lxp8BvRlHxYfauwlvSFafstgwOxzZRpPmFgNjsfEMY3B2QWERD1/rQ\nnwud/3QaAecGBwIlku2lun9PH6yAxGmu4QVdhvI2YH9z6G6HFf0Ufw==\n-----END CERTIFICATE-----\n",
                "KUBECFG_KUBE_APISERVER_PROXY_CLIENT=apiVersion: v1\nkind: Config\nclusters:\n- cluster:\n    api-version: v1\n    certificate-authority: /etc/kubernetes/ssl/kube-ca.pem\n    server: \"https://127.0.0.1:6443\"\n  name: \"local\"\ncontexts:\n- context:\n    cluster: \"local\"\n    user: \"kube-apiserve
r-proxy-client-local\"\n  name: \"local\"\ncurrent-context: \"local\"\nusers:\n- name: \"kube-apiserver-proxy-client-local\"\n  user:\n    client-certificate: /etc/kubernetes/ssl/kube-apiserver-proxy-client.pem\n    client-key: /etc/kubernetes/ssl/kube-apiserver-proxy-client-key.pem",
                "KUBE_ETCD_172_26_14_216_KEY=-----BEGIN RSA PRIVATE KEY-----\nMIIEowIBAAKCAQEA523e0SBqWa1fhrck1Hgxm1tP3TOG3NAdW5dRt72Elp3U+OnV\n5M6M3A1kOXX7UOsit82H4bm0+VzKybod1hrrEK5veTthKF2lwdHuwHmWyVRhlhsw\n1oTA8NdTm3Hfjeep0GZG745WVp/s0xXozgImSSRJW+cQk0hyO9DwjSkSyRO1UNR1\nzOva4zvtxV0VttJwExLPmNLmFDJZuk7OfjqY0lmZoN
RoPA3rcuSyJImozUkLwZ0D\nz5Dt2TfIKoJ4xUTxbEbaOG36ZzWJZJ8SiBnwI6JetUHOxyrmuNkSTLmjUTV73R6O\nBs9pT9ECBMKWeswBBcXEJmN2HC33tMV+gmB7KQIDAQABAoIBAHF3iPt3rSzyuBdQ\nzBnwJEJLbsjBbqnsz7gMZOB1ZwCBud2gqGJacu2hEzapBeMSph8AAlNFvdlVCYgG\nXIKRCBdRrw39cxFbeN2ilDCCbM+hM4dpJXTH+eEbcb6RAk6M+tFWlAj3/JTULEUC\nRPZcT3Ek/WK104aiyn9RXd+X98HlnP
0bvd7D2HDAof+ellTPAd+UtSECEQL4PUzK\nhlMeRAa47/M4M/iEjYhlCJhXND6/Ad5lPPtW6eqWXOWAsZ1ikp7vzIXus40jXPkV\nhV1n0KmEzS+om0U8OUfXQ/mu/1Hn9z5lyRC4l7fRzJVmrr/v/5RK/1qGU6K40sCr\nMuiq8VECgYEA+4q52yOEKMgbasEXMD+Aipoq2uBXH5/PRYmNtTsORyw6UiMLAuMO\n2qPkpQ1yqdx9ZWvCuHFe4hCSEYff+i/CcE/MkulSJBaiTUuNtLebZdlC31j0YL8D\nakfDyyRMVx77Pl/H6g
MYixfayTUALcshQWGclBpcTMcyuAPBZtufbG8CgYEA64fk\nBX86tCQKqPAw8XQXLFPKYFO4uUC32NfPgqN5o4sO4KZ7SuQZLnh9EEQfJkhgT8Tr\n9Fy6+aN8gvISAnCCR/2kOMuHmtFE2mrSmQjkaE7a1s3eHxxUwFI64mHMatCOXS4S\nu6xA+5kd7UobzIqU/HnZzlNulutltbh9Ut08DecCgYAybf6S64zsbCnq/ik6+BA6\nOWxME1wEMBLq+wfZBKz5IenTW8kyW/k3ZlJJsOeDHHxbX/5a4gfGxNG0CAykaPzP\nbYAzF+
nq6ErDulj/mSvjgGpCwt/Doaf6n8amLHHNqZ1vRN6ckOBTyoWHf0O46peR\nNxOgMaS9k9YcREx65Z8RqwKBgQDh6eUn9KJNGWj0I/b+EhkMFo6+GG/NmSr+nfnX\nV5Ab8wzhJC6MZf9VWJK04HJ0WOWwfbTJHYzmWA7c1u25U0tTXBGBvI8kS2fcjKvV\nx/a1qjUz5iEQ/C66jeUXMTFOnx5+d+vWAWIPMg3HhdbmOWKwTPxCcDpaHg3f4Mas\njbHFrwKBgCtgowPUWdh0VMM0sUOg+fJie5fKD+mVNP88jLFjqwtum+nbyTx2
Axxj\nDPi5lpHC8JxTQGt+fTNPziBqzyds5bYx6I3hw18klT8TIhsMdHG+Q7Ayx4v8+Jrq\nOmUa9WFsSYRrHKZqBmAYXXnWyRDk0X6iqCLA6x7JQ56Evu4srL9u\n-----END RSA PRIVATE KEY-----\n", 
```

File deployer (`file-deployer`)

Generic way to deploy files to nodes, this example shows deploying `/etc/kubernetes/audit-policy.yaml`

```
            "Image": "rancher/rke-tools:v0.1.89",
            "Env": [                  
                "FILE_DEPLOY=apiVersion: audit.k8s.io/v1\nkind: Policy\nmetadata:\n  creationTimestamp: null\nrules:\n- level: Metadata\n",
            "Cmd": [
                "sh",
                "-c",
                "t=$(mktemp); echo -e \"$FILE_DEPLOY\" > $t && mv $t /etc/kubernetes/audit-policy.yaml && chmod 600 /etc/kubernetes/audit-policy.yaml"
            ],
```

## Debug

### etcd-rolling-snapshots

Run a separate container called `etcd-rolling-snapshots-test` with adjusted `--creation` and `--retention` for testing

```
echo $(docker run --rm -v /var/run/docker.sock:/var/run/docker.sock  assaflavie/runlike etcd-rolling-snapshots | sed 's/creation=.../creation=1m /' | sed 's/retention=.../retention=1m /' | sed 's/etcd-rolling-snapshots/etcd-rolling-snapshots-test/' | sed 's/$/ --debug/') | bash
```

## K8s containers

- 3 roles (planes), each has their set of k8s containers to be created
- all cluster resources (pods etc) are created using jobs

### Every node

Every node gets (https://github.com/rancher/rke/blob/v1.4.8/cluster/plan.go#L115):

- service-sidekick
- kubelet
- kube-proxy

Every non-controlplane node gets (https://github.com/rancher/rke/blob/v1.4.8/cluster/plan.go#L122):

- nginx-proxy (used to load balance kubelet connections between controlplane nodes, `/etc/kubernetes/ssl/kubecfg-kube-node.yaml` -> `server: "https://127.0.0.1:6443"`)

### controlplane

Every controlplane node gets (https://github.com/rancher/rke/blob/v1.4.8/cluster/plan.go#L125):

- kube-apiserver
- kube-controller-manager
- kube-scheduler

### etcd

Every etcd node gets (https://github.com/rancher/rke/blob/v1.4.8/cluster/plan.go#L132):

- etcd
- etcd-rolling-snapshots (configured by default)

### service-sidekick

This container is in `Created` state, it will not show up in `docker ps` (only in `docker ps -a`). It is a container based on `rancher/rke-tools`, and the purpose is to serve it's volume to k8s containers. Each k8s container is started with `--volumes from=service-sidekick`.

The volume defined the `Dockerfile` of `rancher/rke-tools` (https://github.com/rancher/rke-tools/blob/v0.1.89/package/Dockerfile#L42) is `/opt/rke-tools`. The default entrypoint for k8s containers is `/opt/rke-tools/entrypoint.sh` (https://github.com/rancher/rke/blob/v1.4.8/cluster/plan.go#L46).

## RKE and kontainer-driver-metadata (KDM)

https://github.com/rancher/kontainer-driver-metadata was created to provide out-of-band updates for Rancher Kubernetes. RKE embeds the generated `data.json` (https://github.com/rancher/rke/blob/release/v1.4/data/data.json) into RKE using go-bindata (https://github.com/rancher/rke/blob/release/v1.4/data/bindata.go). You can also point RKE to a different KDM source, see https://github.com/rancher/rke/#building.

Metadata initialization happens at https://github.com/rancher/rke/blob/v1.4.8/metadata/metadata.go, and it initializes:

- System images (what images are used in what Rancher Kubernetes version)
- Add-on templates (the add-on templates like CNI/metrics-server/CoreDNS etc)
- Service options (default arguments for Kubernetes components)
- Docker options (supported Docker versions per Kubernetes version)

### System images

Example:

```
		"v1.26.7-rancher1-1": {
			Etcd:                      "rancher/mirrored-coreos-etcd:v3.5.6",
			Kubernetes:                "rancher/hyperkube:v1.26.7-rancher1",
			Alpine:                    "rancher/rke-tools:v0.1.89",
			NginxProxy:                "rancher/rke-tools:v0.1.89",
			CertDownloader:            "rancher/rke-tools:v0.1.89",
			KubernetesServicesSidecar: "rancher/rke-tools:v0.1.89",
			KubeDNS:                   "rancher/mirrored-k8s-dns-kube-dns:1.22.20",
			DNSmasq:                   "rancher/mirrored-k8s-dns-dnsmasq-nanny:1.22.20",
			KubeDNSSidecar:            "rancher/mirrored-k8s-dns-sidecar:1.22.20",
			KubeDNSAutoscaler:         "rancher/mirrored-cluster-proportional-autoscaler:1.8.6",
			Flannel:                   "rancher/mirrored-flannel-flannel:v0.21.4",
			FlannelCNI:                "rancher/flannel-cni:v0.3.0-rancher8",
			CalicoNode:                "rancher/mirrored-calico-node:v3.25.0",
			CalicoCNI:                 "rancher/calico-cni:v3.25.0-rancher1",
			CalicoControllers:         "rancher/mirrored-calico-kube-controllers:v3.25.0",
			CalicoCtl:                 "rancher/mirrored-calico-ctl:v3.25.0",
			CalicoFlexVol:             "rancher/mirrored-calico-pod2daemon-flexvol:v3.25.0",
			CanalNode:                 "rancher/mirrored-calico-node:v3.25.0",
			CanalCNI:                  "rancher/calico-cni:v3.25.0-rancher1",
			CanalControllers:          "rancher/mirrored-calico-kube-controllers:v3.25.0",
			CanalFlannel:              "rancher/mirrored-flannel-flannel:v0.21.4",
			CanalFlexVol:              "rancher/mirrored-calico-pod2daemon-flexvol:v3.25.0",
			WeaveNode:                 "weaveworks/weave-kube:2.8.1",
			WeaveCNI:                  "weaveworks/weave-npc:2.8.1",
			AciCniDeployContainer:     "noiro/cnideploy:5.2.7.1.81c2369",
			AciHostContainer:          "noiro/aci-containers-host:5.2.7.1.81c2369",
			AciOpflexContainer:        "noiro/opflex:5.2.7.1.81c2369",
			AciMcastContainer:         "noiro/opflex:5.2.7.1.81c2369",
			AciOpenvSwitchContainer:   "noiro/openvswitch:5.2.7.1.81c2369",
			AciControllerContainer:    "noiro/aci-containers-controller:5.2.7.1.81c2369",
			AciGbpServerContainer:     "noiro/gbp-server:5.2.7.1.81c2369",
			AciOpflexServerContainer:  "noiro/opflex-server:5.2.7.1.81c2369",
			PodInfraContainer:         "rancher/mirrored-pause:3.7",
			Ingress:                   "rancher/nginx-ingress-controller:nginx-1.7.0-rancher1",
			IngressBackend:            "rancher/mirrored-nginx-ingress-controller-defaultbackend:1.5-rancher1",
			IngressWebhook:            "rancher/mirrored-ingress-nginx-kube-webhook-certgen:v20230312-helm-chart-4.5.2-28-g66a760794",
			MetricsServer:             "rancher/mirrored-metrics-server:v0.6.3",
			CoreDNS:                   "rancher/mirrored-coredns-coredns:1.9.4",
			CoreDNSAutoscaler:         "rancher/mirrored-cluster-proportional-autoscaler:1.8.6",
			WindowsPodInfraContainer:  "rancher/mirrored-pause:3.7",
			Nodelocal:                 "rancher/mirrored-k8s-dns-node-cache:1.22.20",
		},
```

### Add-on templates

Add-on templates are linked to Rancher Kubernetes versions, not to an image or add-on version. (see [templates/templates.go](https://github.com/rancher/kontainer-driver-metadata/blob/dev-v2.7/pkg/rke/templates/templates.go))

#### CNI

CNI gets deployed in [cluster/network.go](https://github.com/rancher/rke/blob/v1.4.8/cluster/network.go#L346), this is where the configuration options from `cluster.yml` get passed to the template. The rendered template is then deployed using the generic `doAddonDeploy` function (https://github.com/rancher/rke/blob/v1.4.8/cluster/addons.go#L478). This creates a ConfigMap with the template and uses a Job to deploy the template from the ConfigMap. The logic also accounts for updating templates (https://github.com/rancher/rke/blob/v1.4.8/cluster/addons.go#L485).

#### Other add-ons

Other default add-ons are deployed in [cluster/addons.go](https://github.com/rancher/rke/blob/v1.4.8/cluster/addons.go#L162).

### ServiceOptions

Depending on Rancher Kubernetes version, the default arguments are loaded for each Kubernetes container:

- etcd
- kube-apiserver
- kube-controller-manager
- kubelet
- kube-proxy
- kube-scheduler

See [k8s_service_options.go](https://github.com/rancher/kontainer-driver-metadata/blob/dev-v2.7/pkg/rke/k8s_service_options.go) for all configured service options. Initialized in RKE at [metadata/metadata.go](https://github.com/rancher/rke/blob/v1.4.8/metadata/metadata.go#L99) and used in [cluster/plan.go](https://github.com/rancher/rke/blob/v1.4.8/cluster/plan.go#L1193)

### Docker options

Configures what Docker versions can be used for what Kubernetes minor version (https://github.com/rancher/kontainer-driver-metadata/blob/dev-v2.7/pkg/rke/k8s_docker_info.go)

```
func loadK8sVersionDockerInfo() map[string][]string {
	return map[string][]string{
		"1.8":  {"1.11.x", "1.12.x", "1.13.x", "17.03.x"},
		"1.9":  {"1.11.x", "1.12.x", "1.13.x", "17.03.x", "18.06.x", "18.09.x", "19.03.x", "20.10.x"},
		"1.10": {"1.11.x", "1.12.x", "1.13.x", "17.03.x", "18.06.x", "18.09.x", "19.03.x", "20.10.x"},
		"1.11": {"1.11.x", "1.12.x", "1.13.x", "17.03.x", "18.06.x", "18.09.x", "19.03.x", "20.10.x"},
		"1.12": {"1.11.x", "1.12.x", "1.13.x", "17.03.x", "17.06.x", "17.09.x", "18.06.x", "18.09.x", "19.03.x", "20.10.x"},
		"1.13": {"1.11.x", "1.12.x", "1.13.x", "17.03.x", "17.06.x", "17.09.x", "18.06.x", "18.09.x", "19.03.x", "20.10.x"},
		"1.14": {"1.13.x", "17.03.x", "17.06.x", "17.09.x", "18.06.x", "18.09.x", "19.03.x", "20.10.x"},
		"1.15": {"1.13.x", "17.03.x", "17.06.x", "17.09.x", "18.06.x", "18.09.x", "19.03.x", "20.10.x"},
		"1.16": {"1.13.x", "17.03.x", "17.06.x", "17.09.x", "18.06.x", "18.09.x", "19.03.x", "20.10.x"},
		"1.17": {"1.13.x", "17.03.x", "17.06.x", "17.09.x", "18.06.x", "18.09.x", "19.03.x", "20.10.x"},
		"1.18": {"1.13.x", "17.03.x", "17.06.x", "17.09.x", "18.06.x", "18.09.x", "19.03.x", "20.10.x"},
		"1.19": {"1.13.x", "17.03.x", "17.06.x", "17.09.x", "18.06.x", "18.09.x", "19.03.x", "20.10.x"},
		"1.20": {"1.13.x", "17.03.x", "17.06.x", "17.09.x", "18.06.x", "18.09.x", "19.03.x", "20.10.x"},
		"1.21": {"1.13.x", "17.03.x", "17.06.x", "17.09.x", "18.06.x", "18.09.x", "19.03.x", "20.10.x"},
		"1.22": {"1.13.x", "17.03.x", "17.06.x", "17.09.x", "18.06.x", "18.09.x", "19.03.x", "20.10.x"},
		"1.23": {"1.13.x", "17.03.x", "17.06.x", "17.09.x", "18.06.x", "18.09.x", "19.03.x", "20.10.x", "23.0.x", "24.0.x"},
		"1.24": {"1.13.x", "17.03.x", "17.06.x", "17.09.x", "18.06.x", "18.09.x", "19.03.x", "20.10.x", "23.0.x", "24.0.x"},
		"1.25": {"1.13.x", "17.03.x", "17.06.x", "17.09.x", "18.06.x", "18.09.x", "19.03.x", "20.10.x", "23.0.x", "24.0.x"},
		"1.26": {"1.13.x", "17.03.x", "17.06.x", "17.09.x", "18.06.x", "18.09.x", "19.03.x", "20.10.x", "23.0.x", "24.0.x"},
	}
}
```

### K8s version info



## Links

- [kdmq](https://github.com/superseb/kdmq): CLI tool to query KDM info
- [KDM markdown tables](https://github.com/superseb/ranchertools/tree/master/kdm/v2.7): to visualize what versions are used in what Kubernetes version
- [Rancher release notes/version tables](https://github.com/superseb/ranchertools/tree/master/release-notes): Aggregated release notes per project and versions tables for k3s/RKE1/RKE2
