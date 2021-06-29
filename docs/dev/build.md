# Developing With Tanzu Framework

Framework is meant to be extended, and these docs will help outline the available dev-centric features.

## Local Builds

There are Make targets provided to help establish a good workflow for local development.
With these workflows you can iterate on changes to APIs, Framework CLI or CLI Plugins.

### Building APIs

To build the APIs in Framework, the following commands exist:

`make generate`: Generates API and boilerplate code via controller-gen.

`make manifests`: Runs controller-gen to output manifests to `config/crd/bases`.

`make deploy`: Runs kustomize on config/ output and pipes to kubectl to deploy to your current kubeconfig
context.

### Framework CLI

The CLI has specific targets for local development due to its distributed nature.

`make build-install-cli-local`: cleans, builds and installs plugins.

`make test`: Performs a suite of tests on the CLI and addons.

### Building Plugins

Consider these command while building plugins:

`make build-install-cli-all`: cleans, builds and installs plugins.

`make build-cli-local`: Only builds the Tanzu CLI locally.
