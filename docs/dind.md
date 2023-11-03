# rke DinD

Docker-in-Docker (DinD) feature is created mainly for testing.

DinD creates Docker containers locally which are then used as cluster nodes.

It still requires a cluster.yml with minimal information:

```
kubernetes_version: "$KUBE_VERSION"
nodes:
- address: foo
  role:
  - controlplane
  - worker
  - etcd
  user: bar
```
where KUBE_VERSION is any version in the output list of the command `rke config --list-version --all`. Note that in multinode clusters, the value of address must be unique

To create a cluster with DinD execute:
`rke up --dind`

After the creation is completed, you will see one Docker container per node running. You can use the created `kube_config_cluster.yml` to access the cluster as in a normal RKE deployment

To remove a cluster with DinD execute:
`rke remove --dind`

The Docker containers should be removed completely

# FAQ

* Do I need to create a network for the Docker containers?

No. The created Docker containers use the default Docker bridge network which allows to connect from the host to the containers on IPs in the subnet "172.17.0.0/16". However, they have no ssh service running, therefore you `docker exec` if you want to connect to them

* What image is used for those containers?

The default Docker dind image. You can find the used tag in this [link](https://github.com/rancher/rke/blob/release/v1.5/dind/dind.go#L16)

* Are all RKE features supported in DinD?

No. Only creation and removal of clusters are supported. You can of course operate the Kubernetes cluster as a normal cluster

* Is DinD production ready?

No. DinD's purpose is solely RKE testing. Anything else is out of scope
