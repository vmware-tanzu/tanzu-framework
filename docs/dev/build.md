# Developing With Tanzu Framework

Framework is meant to be extended, and these docs will help outline the available dev-centric features.

## Local Builds

There are Make targets provided to help establish a good workflow for local development.
With these workflows you can iterate on changes to APIs, Framework CLI or CLI Plugins.

Note: Go v1.17 or greater is required to build Framework components.

### Building APIs

To build the APIs in Framework, the following commands exist:

`make generate`: Generates API and boilerplate code via controller-gen.

`make manifests`: Runs controller-gen to output manifests to `config/crd/bases`.

`make install`: To install CRDs into the cluster.

API controllers that exist in the Framework repo

* [FeatureGates](https://github.com/vmware-tanzu/tanzu-framework/tree/main/featuregates)
* [Capabilities](https://github.com/vmware-tanzu/tanzu-framework/tree/main/capabilities)

Each controller directory has its own Dockerfile, Makefile and manifests needed to build the image and
deploy the controller to the cluster.

### Capabilities

Framework provides Capability discovery
[Go package](https://github.com/vmware-tanzu/tanzu-framework/tree/main/capabilities/client/pkg/discovery)
and Capability API to query a cluster's capabilities. It can be used to understand the API surface area and query for
objects in the cluster.

For more detailed information on Capability functionality offered by Framework check out this
[doc](../runtime-core/capability-discovery.md)

### Features and FeatureGates

Framework offers Features and FeatureGates APIs to allow developers to deliver new functionality to users rapidly but
safely. With these powerful APIs the teams can modify the system behavior without changing the code for more controlled
experimentation over the lifecyle of features, these can be incredibly useful for agile management style environments.

For more detailed information on these APIs check out this [doc](../runtime-core/features-and-featuregates.md)
