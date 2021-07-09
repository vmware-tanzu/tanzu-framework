# Tanzu CLI Getting Started

A simple set of instructions to set up and use the Tanzu CLI.

## Installation
### Install the latest release of Tanzu CLI

`linux-amd64`,`windows-amd64`, and `darwin-amd64` are the OS-ARCHITECTURE combinations we support now.

If you want to install the latest release of the Tanzu CLI, you can run the below commands:

#### MacOS/Linux
```shell
export ARCH=$(case $(uname -m) in x86_64) echo -n amd64 ;; aarch64) echo -n arm64 ;; esac)
export OS=$(uname | awk '{print tolower($0)}')

curl -o tanzu https://storage.googleapis.com/tanzu-cli/artifacts/core/latest/tanzu-core-${OS}_${ARCH} && \
    mv tanzu /usr/local/bin/tanzu && \
    chmod +x /usr/local/bin/tanzu
```

#### Windows
#### AMD64
Windows executable can be found at https://storage.googleapis.com/tanzu-cli/artifacts/core/latest/tanzu-core-windows_amd64.exe


### Build and install the CLI and plugins locally
#### Prerequisites

* [go](https://golang.org/dl/) version 1.16

Clone Tanzu Framework and run the below command to build and install CLI and plugins locally.
```
TANZU_CLI_NO_INIT=true make build-install-cli-local
```

## Usage

```
Usage:
  tanzu [command]

Available command groups:

  Admin
    builder                 Build Tanzu components 
    test                    Test the CLI 
        
  Run
    cluster                 Kubernetes cluster operations 
    kubernetes-release      Kubernetes release operations 
    management-cluster      Kubernetes management cluster operations 
        
  System
    completion              Output shell completion code 
    config                  Configuration for the CLI 
    init                    Initialize the CLI 
    login                   Login to the platform 
    plugin                  Manage CLI plugins 
    update                  Update the CLI 
    version                 Version information 
        
  Version
    alpha                   Alpha CLI commands 
        

Flags:
  -h, --help   help for tanzu

Use "tanzu [command] --help" for more information about a command. 
```

## What's next
### Create your own plugin
To bootstrap a new plugin, follow the `builder` plugin documentation [here](../../cmd/cli/plugin-admin/builder/README.md).

Check out the [plugin implementation guide](../cli/plugin_implementation_guide.md) for more details.
