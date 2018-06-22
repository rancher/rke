# rke

Rancher Kubernetes Engine, an extremely simple, lightning fast Kubernetes installer that works everywhere.

## Download

Please check the [releases](https://github.com/rancher/rke/releases/) page.

## Requirements

- Docker versions `1.11.2` up to `1.13.1` and `17.03.x` are validated for Kubernetes versions 1.8, 1.9 and 1.10
- OpenSSH 7.0+ must be installed on each node for stream local forwarding to work.
- The SSH user used for node access must be a member of the `docker` group:

```bash
usermod -aG docker <user_name>
```

- Ports 6443, 2379, and 2380 should be opened between cluster nodes.
- Swap disabled on worker nodes.

## Getting Started

Starting out with RKE? Check out this [blog post](http://rancher.com/an-introduction-to-rke/) or the [Quick Start Guide](http://rancher.com/docs/rke/v0.1.x/en/quick-start-guide).

Please refer to our [RKE docs](http://staging.rancher.com/docs/rke/v0.1.x/en/) for information on how to get started!

## Deploying Rancher 2.x using rke

Using RKE's pluggable user addons, it's possible to deploy Rancher 2.x server in HA with a single command. Detailed instructions can be found [here](https://rancher.com/docs/rancher/v2.x/en/installation/ha-server-install/).

## License

Copyright (c) 2018 [Rancher Labs, Inc.](http://rancher.com)

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

[http://www.apache.org/licenses/LICENSE-2.0](http://www.apache.org/licenses/LICENSE-2.0)

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
