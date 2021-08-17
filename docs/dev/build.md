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

```
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

### Features and FeatureGates

Framework offers Features and FeatureGates APIs to allow developers to deliver new functionality to users rapidly but safely.
With these powerful APIs the teams can modify the system behavior without changing the code for more controlled experimentation
over the lifecyle of features, these can be incredibly useful for agile management style environments.

More detailed information on these APIs check out this [doc](../api-machinery/features-and-featuregates.md) 
