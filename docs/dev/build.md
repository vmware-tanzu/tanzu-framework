# Developing With Tanzu Framework

Framework is meant to be extended, and these docs will help outline the available dev-centric features.

## Local Builds

There are Make targets provided to help establish a good workflow for local development.
With these workflows you can iterate on changes to APIs, Framework CLI or CLI Plugins.

### Building APIs

To build the APIs in Framework, the following commands exist:

`make generate`: Generates API and boilerplate code via controller-gen.

`make manifests`: Runs controller-gen to output manifests to `config/crd/bases`.

`make install`: To install CRDs into the cluster.

API controllers that exist in the Framework repo

* [Addons](https://github.com/vmware-tanzu/tanzu-framework/tree/main/addons)
* [Capabilities](https://github.com/vmware-tanzu/tanzu-framework/tree/main/pkg/v1/sdk/capabilities)
* [FeatureGates](https://github.com/vmware-tanzu/tanzu-framework/tree/main/pkg/v1/sdk/features)
* [TKR](https://github.com/vmware-tanzu/tanzu-framework/tree/main/pkg/v1/tkr)

Each controller directory has its own Dockerfile, Makefile and manifests needed to build the image and
deploy the controller to the cluster.

### Framework CLI

The CLI has specific targets for local development due to its distributed nature.

`make build-install-cli-local`: cleans, builds and installs plugins locally for
your platform

`make test`: Performs a suite of tests on the CLI and API controllers.

### Building Plugins

The CLI builder can accept directories using a single, global Go module
or multiple Go modules within sub directories.

Generally, the directory structure when building plugins may look like:

```sh
plugins-directory
|- foo-plugin
|- bar-plugin
```

where `foo-plugin` and `bar-plugin` are within a single, global, top level Go module
or are both individually, their own Go module. Both are accepted.

Consider these command while building plugins:

`make build-install-cli-all`: cleans, builds and installs plugins.

`make build-install-cli-local`: cleans, builds, installs CLI and plugins locally
for your platform.

`make build-cli-local`: Only builds the Tanzu CLI locally.

Check out the [plugin implementation guide](../cli/plugin_implementation_guide.md)
for more details on how to write plugins for Tanzu CLI.

### Capabilities

Framework provides Capability discovery
[GO package](https://github.com/vmware-tanzu/tanzu-framework/tree/main/pkg/v1/sdk/capabilities/discovery)
and Capability API to query a cluster's capabilities. It can be used to understand the API surface area and query for
objects in the cluster.

For more detailed information on Capability functionality offered by Framework check out this
[doc](../api-machinery/capability-discovery.md)

### Features and FeatureGates

Framework offers Features and FeatureGates APIs to allow developers to deliver new functionality to users rapidly but
safely. With these powerful APIs the teams can modify the system behavior without changing the code for more controlled
experimentation over the lifecyle of features, these can be incredibly useful for agile management style environments.

For more detailed information on these APIs check out this [doc](../api-machinery/features-and-featuregates.md)

### Context-aware API-driven Plugin Discovery

The Tanzu CLI is an amalgamation of all the Tanzu infrastructure elements under one unified core CLI experience. The core CLI supports a plugin model where the developers of different Tanzu services (bundled or SaaS) can distribute plugins that target functionalities of the services they own. When users switch between different services via the CLI context, we want to surface only the relevant plugins for the given context for a crisp user experience.

This feature is gated by `features.global.context-aware-cli-for-plugins` CLI feature flag and can be turned on/off as described [here](../cli/config-features.md). When this feature is enabled, the CLI will not honor plugins installed prior to this feature being enabled, instead users will need to install plugins again with either `tanzu plugin install <plugin_name>` or `tanzu plugin sync`.

For more detailed information on these design check out this [doc](../design/context-aware-plugin-discovery.md)

### Troubleshooting

#### Errors building docker images on RHEL/Fedora with "failed to solve with frontend dockerfile.v0"

When building on Fedora, you may see the following error:

``` shell
➜ docker build .
[+] Building 5.1s (4/4) FINISHED
 => [internal] load build definition from Dockerfile 0.1s
 => => transferring dockerfile: 1.43kB 0.0s
 => [internal] load .dockerignore 0.2s
 => => transferring context: 2B 0.0s
 => resolve image config for docker.io/docker/dockerfile:1.4 1.7s
 => docker-image://docker.io/docker/dockerfile:1.4@sha256:443aab4ca21183e069e7d8b2dc68006594f40bddf1b15bbd83f5137bd93e80e2 2.7s
 => => resolve docker.io/docker/dockerfile:1.4@sha256:443aab4ca21183e069e7d8b2dc68006594f40bddf1b15bbd83f5137bd93e80e2 0.1s
 => => sha256:24d064a369eda7bc7839b6c1c227eac7212d06ca09a8235a4bed467f8acf180d 528B / 528B 0.0s
 => => sha256:84495a15555de1a8f4738f58268fa8949547068198f8d0fa2a3e3a693d7f923f 2.37kB / 2.37kB 0.0s
 => => sha256:09768fef35f2ee387f57e401ae685727d12d1c70c6fd8545a422850167bf1940 9.94MB / 9.94MB 2.1s
 => => sha256:443aab4ca21183e069e7d8b2dc68006594f40bddf1b15bbd83f5137bd93e80e2 2.00kB / 2.00kB 0.0s
 => => extracting sha256:09768fef35f2ee387f57e401ae685727d12d1c70c6fd8545a422850167bf1940 0.2s
failed to solve with frontend dockerfile.v0: failed to solve with frontend gateway.v0: exit code: 1
```

Tanzu Framework uses buildkit to build Docker images in order to speed up builds. Unfortunately, there are unresolved issues with SELinux and Docker
Buildkit, and SELinux must be currently set to permissive with `sudo setenforce 0`.

See [this upstream issue](https://github.com/moby/buildkit/issues/2295) for more information.
