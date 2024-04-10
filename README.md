# rke

*This file is auto-generated from README-template.md, please make any changes there.*

Rancher Kubernetes Engine, an extremely simple, lightning fast Kubernetes installer that works everywhere.

## Latest Release
* v1.5
  * v1.5.8 - Read the full release [notes](https://github.com/rancher/rke/releases/tag/v1.5.8).
* v1.4
  * v1.4.17 - Read the full release [notes](https://github.com/rancher/rke/releases/tag/v1.4.17).

## Download

Please check the [releases](https://github.com/rancher/rke/releases/) page.

## Requirements

Please review the [Requirements](https://rke.docs.rancher.com/os) for each node in your Kubernetes cluster.

## Getting Started

Please refer to our [RKE docs](https://rke.docs.rancher.com/) for information on how to get started!
For cluster config examples, refer to [RKE cluster.yml examples](https://rke.docs.rancher.com/example-yamls)

## Installing Rancher HA using rke

Please use [Setting up a High-availability RKE Kubernetes Cluster](https://ranchermanager.docs.rancher.com/how-to-guides/new-user-guides/kubernetes-cluster-setup/rke1-for-rancher) to install Rancher in a high-availability configuration.

## Building

RKE can be built using the `make` command, and will use the scripts in the `scripts` directory as subcommands. The default subcommand is `ci` and will use `scripts/ci`. Cross compiling can be enabled by setting the environment variable `CROSS=1`. The compiled binaries can be found in the `build/bin` directory. Dependencies are managed by Go modules and can be found in [go.mod](https://github.com/rancher/rke/blob/master/go.mod).

Read [codegen/codegen.go](./codegen/codegen.go) to check the default location for fetching `data.json`. You can override the default location as seen in the example below:

```bash
# Fetch data.json from default location
go generate

# Fetch data.json from URL using RANCHER_METADATA_URL
RANCHER_METADATA_URL=${URL} go generate

# Use data.json from local file
RANCHER_METATDATA_URL=./local/data.json go generate

# Compile RKE
make
```

To override RANCHER_METADATA_URL at runtime, populate the environment variable when running rke CLI. For example:

```bash
RANCHER_METADATA_URL=${URL} rke [commands] [options]

RANCHER_METADATA_URL=${./local/data.json} rke [commands] [options]
```
    
## License

Copyright © 2017 - 2023 SUSE LLC
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
