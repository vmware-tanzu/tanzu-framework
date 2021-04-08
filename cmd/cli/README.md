# Tanzu CLI

## Summary
Our objective is to support building CLI experiences built on a shared framework to enable a consistent, unified product experience for our users. We have designed the CLI around a pluggable model to support a broader set of products as they are ready to adopt it.

For more detail, please see the [Tanzu CLI purpose doc](https://docs.google.com/document/d/1X34ZNkPG_kEMSySpFjAQsmX2Xn1dXTksbVxXUgUk-QM/edit?usp=sharing).

## Overview
The CLI is based on the kubectl plugin architecture. This architecture enables teams to build, own, and release their own piece of functionality as well as enable external partners to integrate with the system.

For more detail, please see the [Tanzu CLI architecture doc](https://docs.google.com/document/d/1qCarTtSUxJzYJweiHsOQhObTc2L4f9smAXAlIheMFfI/edit#).

---

## Installation

### MacOS
```shell
curl -o tanzu https://storage.googleapis.com/tanzu-cli/artifacts/core/latest/tanzu-core-darwin_amd64 && \
    mv tanzu /usr/local/bin/tanzu && \
    chmod +x /usr/local/bin/tanzu
```
### Linux
#### i386
```shell
curl -o tanzu https://storage.googleapis.com/tanzu-cli/artifacts/core/latest/tanzu-core-linux_386 && \
    mv tanzu /usr/local/bin/tanzu && \
    chmod +x /usr/local/bin/tanzu
```
#### AMD64
```shell
curl -o tanzu https://storage.googleapis.com/tanzu-cli/artifacts/core/latest/tanzu-core-linux_amd64 && \
    mv tanzu /usr/local/bin/tanzu && \
    chmod +x /usr/local/bin/tanzu
```

### Windows
Windows executable can be found at https://storage.googleapis.com/tanzu-cli/artifacts/core/latest/tanzu-core-windows_amd64.exe


---

## Usage
```
tanzu [command]

Available command groups:

  Manage
    clustergroup            A group of Kubernetes clusters

  Run
    cluster                 Kubernetes cluster operations
    kubernetes-release      Kubernetes release operations
    management-cluster      Commands for creating and managing TKG clusters

  System
    completion              Output shell completion code
    config                  Configuration for the CLI
    init                    Initialize the CLI
    login                   Login to the platform
    plugin                  Manage CLI plugins
    update                  Update the CLI
    version                 Version information


Flags:
  -h, --help   help for tanzu

Use "tanzu [command] --help" for more information about a command.
```
## Documentation
TODO

## Contribution

👍🎉 First off, thanks for taking the time to contribute! 🎉👍

The first step to contributing is to come say hello at SIG meetings. You can also take a peek at our [issues](https://github.com/vmware-tanzu-private/core/issues). 

The plugin author guide provides details on the getting the process started, but in short:
* The core framework and components are written in go
* Every plugin requires a README that explains its usage.
* Core plugins are required to conform to [design guidance](https://github.com/vmware-tanzu-private/core/blob/main/docs/cli/style_guide.md) to ensure a consistent user experience.
* The plugin will live in an approved repo to enable review and testing.
* Every CLI plugin requires a [nested test executable](https://github.com/vmware-tanzu-private/core/blob/main/docs/cli/plugin_implementation_guide.md#tests).

Non-core/3rd party plugins are welcome and encouraged, and can be added by users via `tanzu plugin repo add NAME`.  
Note that non-core plugins will not be included in plugin distributions distributed by VMware and commands will be namespaced differently in help.

For more detail, please see the [Tanzu CLI Implementation Guide](/docs/cli/plugin_implementation_guide.md)

## License
TODO
