# Tanzu Framework

## What is Framework?
Framework is both the foundation for Tanzu editions and a set of
building blocks that can be used to extend Tanzu. Framework enables
APIs, code, and documentation to be shared.

This model provides a few primary benefits.

* Enables re-use of existing work in Tanzu editions.
* Improvements to Framework benefit all editions.
* Fosters cross-team collaboration.
* Allows extensions to Tanzu to be built following the same patterns as
  components of Tanzu.

## Features

- [Tanzu CLI and core Plugins](docs/cli/commands/)
- [Capabilities API](apis/run/v1alpha1/capability_types.go)
- [Features](apis/config/v1alpha1/feature_types.go) and [FeatureGates](apis/config/v1alpha1/featuregate_types.go) APIs
- [CLI development and build tools](docs/cli/cli-architecture.md)
- [Controllers development resources](docs/api-machinery/)
- [User and developer documentation and guides](docs/dev/)
- [Tanzu Kubernetes Release API](apis/run/v1alpha1/tanzukubernetesrelease_types.go)
- [Tanzu Addons](addons/)

## Why it is important?
It acts as the foundation for Tanzu editions enabling the re-use of the existing work
in Tanzu Framework.

## What does this repository contain?
Framework acts as the central location for APIs, controllers,
documentation, and CLI plugins used by all Tanzu editions.

## Getting Started
The best way to get started with the Tanzu CLI is to see it in action.

Get bootstrapped with the CLI getting started [guide](docs/cli/getting-started.md).

## How to contribute
Check out the [contribution guidelines](CONTRIBUTING.md) to learn more about how to contribute.

## Troubleshooting
Reference our [troubleshooting](docs/dev/troubleshooting.md) tips if you find yourself in need of more debug options.

## Roadmap
Check out Framework projects [Roadmap](ROADMAP.md), and consider contributing!
