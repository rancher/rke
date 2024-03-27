# Title

* Decoupling RKE1 and KDM

## Date

- Created: 2023-12-04
- Last updated: 2023-12-04

## Status

Discussing

## Context

### What does RKE1 require from KDM?

RKE1 deploys a Kubernetes cluster, which means installing Kubernetes components and add-ons such as nginx ingress controller or CNI plugins. Each RKE1 release is capable of deploying several Kubernetes releases, for example RKE1 `Release v1.4.11` can deploy:
```
v1.26.9-rancher1-1
v1.25.14-rancher1-1
v1.24.17-rancher1-1
v1.23.16-rancher2-3
```
All of these Kubernetes releases deploy the same add-ons but with different versions. RKE1 knows about what addons versions are mapped per Kubernetes release by querying the `data/data.json` file of KDM. That file is embedded in RKE1 and for example, for `v1.26.9-rancher1-1`:
```
  "v1.26.9-rancher1-1": {
   "etcd": "rancher/mirrored-coreos-etcd:v3.5.6",
   "alpine": "rancher/rke-tools:v0.1.96",
   "nginxProxy": "rancher/rke-tools:v0.1.96",
   "certDownloader": "rancher/rke-tools:v0.1.96",
   "kubernetesServicesSidecar": "rancher/rke-tools:v0.1.96",
   "kubedns": "rancher/mirrored-k8s-dns-kube-dns:1.22.20",
   "dnsmasq": "rancher/mirrored-k8s-dns-dnsmasq-nanny:1.22.20",
   "kubednsSidecar": "rancher/mirrored-k8s-dns-sidecar:1.22.20",
   "kubednsAutoscaler": "rancher/mirrored-cluster-proportional-autoscaler:1.8.6",
   "coredns": "rancher/mirrored-coredns-coredns:1.9.4",
   "corednsAutoscaler": "rancher/mirrored-cluster-proportional-autoscaler:1.8.6",
   "nodelocal": "rancher/mirrored-k8s-dns-node-cache:1.22.20",
   "kubernetes": "rancher/hyperkube:v1.26.9-rancher1",
   "flannel": "rancher/mirrored-flannel-flannel:v0.21.4",
   "flannelCni": "rancher/flannel-cni:v0.3.0-rancher8",
   "calicoNode": "rancher/mirrored-calico-node:v3.25.0",
   "calicoCni": "rancher/calico-cni:v3.25.0-rancher1",
   "calicoControllers": "rancher/mirrored-calico-kube-controllers:v3.25.0",
   "calicoCtl": "rancher/mirrored-calico-ctl:v3.25.0",
   "calicoFlexVol": "rancher/mirrored-calico-pod2daemon-flexvol:v3.25.0",
   "canalNode": "rancher/mirrored-calico-node:v3.25.0",
   "canalCni": "rancher/calico-cni:v3.25.0-rancher1",
   "canalControllers": "rancher/mirrored-calico-kube-controllers:v3.25.0",
   "canalFlannel": "rancher/mirrored-flannel-flannel:v0.21.4",
   "canalFlexVol": "rancher/mirrored-calico-pod2daemon-flexvol:v3.25.0",
   "weaveNode": "weaveworks/weave-kube:2.8.1",
   "weaveCni": "weaveworks/weave-npc:2.8.1",
   "podInfraContainer": "rancher/mirrored-pause:3.7",
   "ingress": "rancher/nginx-ingress-controller:nginx-1.7.0-rancher1",
   "ingressBackend": "rancher/mirrored-nginx-ingress-controller-defaultbackend:1.5-rancher1",
   "ingressWebhook": "rancher/mirrored-ingress-nginx-kube-webhook-certgen:v20230312-helm-chart-4.5.2-28-g66a760794",
   "metricsServer": "rancher/mirrored-metrics-server:v0.6.3",
   "windowsPodInfraContainer": "rancher/mirrored-pause:3.7",
   "aciCniDeployContainer": "noiro/cnideploy:6.0.3.1.81c2369",
   "aciHostContainer": "noiro/aci-containers-host:6.0.3.1.81c2369",
   "aciOpflexContainer": "noiro/opflex:6.0.3.1.81c2369",
   "aciMcastContainer": "noiro/opflex:6.0.3.1.81c2369",
   "aciOvsContainer": "noiro/openvswitch:6.0.3.1.81c2369",
   "aciControllerContainer": "noiro/aci-containers-controller:6.0.3.1.81c2369"
  },
```
Apart from that, RKE1 also relies on `data/data.json` to know what manifests it should deploy to install those addons. There are two important parts in `data/data.json`` including this information. Part 1 specifies the logic to pick up a manifest depending on the Kubernetes release. For example, for calico, we can see:
```
  "calico": {
...
   "\u003e=1.21.0-rancher1-1 \u003c1.22.0-rancher1-1": "calico-v3.19.0",
   "\u003e=1.22.0-rancher1-1 \u003c1.22.17-rancher1-1": "calico-v3.21.1",
   "\u003e=1.22.17-rancher1-1 \u003c1.22.17-rancher1-2": "calico-v3.22.5",
   "\u003e=1.22.17-rancher1-2 \u003c1.23.0-rancher1-1": "calico-v3.22.5-rancher2",
   "\u003e=1.23.0-rancher1-1 \u003c1.23.15-rancher1-1": "calico-v3.21.1",
   "\u003e=1.23.15-rancher1-1 \u003c1.23.16-rancher1-1": "calico-v3.22.5",
   "\u003e=1.23.16-rancher1-1 \u003c1.24.0-rancher1-1": "calico-v3.22.5-rancher2",
   "\u003e=1.24.0-rancher1-1 \u003c1.24.9-rancher1-1": "calico-v3.21.1",
   "\u003e=1.24.10-rancher1-1 \u003c1.25.0-rancher1-1": "calico-v3.22.5-rancher2",
   "\u003e=1.24.9-rancher1-1 \u003c1.24.10-rancher1-1": "calico-v3.22.5",
   "\u003e=1.25.0-rancher1-1 \u003c1.26.0-rancher1-1": "calico-v3.24.1",
   "\u003e=1.26.0-rancher1-1 \u003c1.27.0-rancher1": "calico-v3.25.0",
   "\u003e=1.27.0-rancher1-1": "calico-v3.26.1",
   "\u003e=1.8.0-rancher0 \u003c1.13.0-rancher0": "calico-v1.8"
  },
```
It's a bit cryptic but each line is stating a min and max Kubernetes release. RKE1 checks the Kubernetes release we are using to know what calico manifest to install. For example, for `v1.26.9-rancher1-1`, we will use the manifest `calico-v3.25.0`.

Part 2 includes the actual manifest, which is also part of `data/data.json`. I will not show it here because it is too long but you can expect a long string representing all required manifests after the variable `calico-v3.25.0=`

### How does RKE1 consume KDM?

As stated in the [README.md](https://github.com/rancher/rke#building), before building RKE1 we must execute `go generate`, which will fetch `data/data.json` from a default URL set by the [defaultURL variable](https://github.com/rancher/rke/blob/release/v1.5/codegen/codegen.go#L13). We can modify that URL with env variables when executing `go generate` to embed a different `data/data.json` into the rke binary.

Note that it is also possible to override the URL at runtime with an environment variable.

### How is the release process right now

It all starts with KDM creating a branch with the name `dev-vX.Y-YEAR-MONTH-patches`, for example [dev-v2.7-2023-11-patches](https://github.com/rancher/kontainer-driver-metadata/tree/dev-v2.7-2023-11-patches).

Then, a PR is created updating that KDM branch:
* The new images that will be used (both Kubernetes and addons)
* The new manifests if needed (e.g. if there is a new addon version that changes the manifest)
* The logic to pick up manifests if needed (e.g. if there is a new addon manifest)

Here is [an example](https://github.com/rancher/kontainer-driver-metadata/pull/1257/files)

When merged in KDM, the default URL in RKE1 is changed, pointing at the previously described branch. Here is [an example](https://github.com/rancher/rke/commit/ca8b578f55ffd46c5ff9485a4fcedcb9107f9425).

Once the default URL PR is merged, RKE1 releases an RC. This RC is used by QA (infracloud team) to run its RKE1/KDM tests. If a regression is found, KDM or RKE1 code is updated and a new RC is released. Once QA signs off an RC, that RC can be considered for an "official" RKE1 release. However, we don't release RKE1 until KDM officially releases. KDM is consumed by other projects, such as Rancher Manager or RKE2, so this final KDM release can be delayed.

When KDM releases, it merges the changes of `dev-vX.Y-YEAR-MONTH-patches` into `release-vX.Y` and the default URL in RKE1 is changed again. Here is [an example](https://github.com/rancher/rke/commit/6e8ec25f09f2ff88f66b0fe643f9d40b84f43756). After that, RKE1 is officially released.

Note that, as of today, KDM branches are created after upstream Kubernetes releases are out.


### Problems of the current release process

It is impossible to have a stable cadence of releases in RKE1 because RKE1 depends on the KDM release and the KDM release cedence is unpredictable. The main reason for this unpredicatability is because Rancher Manager uses KDM for its functioning and will not allow it to be released until all tests are working. Rancher Manager is a very complex software that could require a lot of time (weeks/months) until it is ready to release.

As a consequence:
* It is impossible to know when does the releases process start, which makes it difficult to plan what gets in or not. e.g. maybe the release cycle is super short because we are already late and only a few issues can be added.
* It is impossible to clearly state when does code freeze start or finishes
* It is impossible for QA to plan testing ahead
* Your PR might have been merged more than a month ago and now you get questions from QA about that PR, which is half forgotten
* It could take months for an important update to happen in an addon (e.g. CVE)
* Big version jumps are sometimes made, which could be risky for regressions
* Human/conscious decisions taken, which makes automation complicated. For example, as releases can be delayed for a long time, new unplanned stuff gets sometimes sneaked into a release while waiting for KDM to officially release (e.g. CVE fix for k8s-dns)


## Proposal

Two related proposals:
* Decouple RKE1 from Rancher Manager unpredictable release cycle as much as possible. To do so, we need to make RKE1 able to release without waiting on KDM to be officially released.
* Do monthly releases based on the upstream Kubernetes release cadence (following rke2/k3s cadence model)

My proposal is the following:
* In KDM, continue creating a branch with the name `dev-vX.Y-YEAR-MONTH-patches`
* Same as with the current release, RKE1 changes the default URL PR and points to that new branch
* Same as with the current release, RKE1 releases an RC and gets tested by QA
* If QA signs off, that RC becomes the official release

In other words, RKE1 will not wait until Rancher Manager is thorough with its testing to make KDM release official

### Potential problems:

1 - After RKE1 release, `dev-vX.Y-YEAR-MONTH-patches` gets a new PR because rke2 (or other project) requires something else (e.g. bugfix)

RKE1 would have been released using an old `dev-vX.Y-YEAR-MONTH-patches`. However, the difference would be rke2 related code. Therefore, the important RKE1 code in KDM will still be the same. The difference will exist for a short span of time, as it will be removed in the next monthly release `dev-vX.Y-YEAR-MONTH+1-patches` which will include the rke2 bits

2 - After RKE1 release, Rancher Manager QA detects a bug in the RKE1 version (how do fix this now??)

Given the long Rancher Manager release cycle, if it does not really block Rancher Manager release, possibly we can fix the bug in the next monthly release.
If it is necessary to quickly fix the bug, we can make an r2 of `dev-vX.Y-YEAR-MONTH-patches` by updating that branch with the necessary changes. This will probably require adding a patch version in the release version:
* RKE1 v1.4.10 with such fix would become something like RKE1 v1.4.10r2
* Kubernetes releases included in that RKE1 v1.4.10r2 would change:
  * v1.25.4-rancher1-1 ==> v1.25.4-rancher2-1

In any case, this is a problem that RKE1 currently could have had but luckily never experienced

3 - By the time Rancher Manager finally releases with RKE1 v1.5.0, RKE1 has already released v1.5.1 and v1.5.2, is that a problem?

It shouldn't be a problem. Rancher Manager vendors in RKE1 with version v1.5.0 and it will be able to deploy that one. In the next release cycle, they can pick to deploy RKE1 v1.5.2. In this case, v1.5.1 will only be available via the rke binary and not via Rancher Manager. This is a change in today's behaviour, but that is fine.

4 - Rancher Manager KDM's data.json will not be in sync with RKE1 KDM's data.json, is that a problem?

It shouldn't. RKE1 KDM's data.json will be based on `dev-vX.Y-YEAR-MONTH-patches` branch and Rancher KDM's data.json will be based on `release-vX.Y` branch. As long as the vendored RKE1 version in Rancher Manager is covered in KDM's data.json of `release-vX.Y` branch, Rancher Manager will not have any problem to deploy RKE1. 

Besides, at some point, when Rancher's QA gives the sign off, both branches will be in sync. So the difference will anyway not be large.

5 - Rancher Manager finally released and now `dev-v2.7-2024-01-patches` branch became `release-v2.7` branch. In the meanwhile `dev-v2.7-2024-02-patches` and `dev-v2.7-2024-03-patches` were released. The new Rancher Manager will focus on `dev-v2.7-2024-04-patches`, will the added code in `02` and `03` branches be lost?

No, it will not. `data.json` includes all the information of the previous branches



## Decision

*(This section describes our response to these forces. It is stated in full sentences, with active/indicate voice. "We do...")*

## Consequences

*(This section describes the resulting context, after applying the decision. All consequences should be listed here, not just the "positive" ones. A particular decision may have positive, negative, and neutral consequences, but all of them affect the team and project in the future.)*
