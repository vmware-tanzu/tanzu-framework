# Tanzu CLI Getting Started

A simple set of instructions to set up and use the Tanzu CLI.

## Installation
### Install the latest release of Tanzu CLI

`linux-amd64`,`windows-amd64`, and `darwin-amd64` are the OS-ARCHITECTURE 
combinations we support now.

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

Clone Tanzu Framework and run the below command to build and install CLI and 
plugins locally for your platform.
```
TANZU_CLI_NO_INIT=true make build-install-cli-local
```

If you additionaly want to build and install CLI and plugins for all platforms, run:
```
TANZU_CLI_NO_INIT=true make build-install-cli-all
```

The CLI currently contains a default distribution which is the default set of plugins that should be installed on
initialization. Initialization of the distributions can be prevented by setting the env var `TANZU_CLI_NO_INIT=true`.
Check out this [doc](../cli/plugin_implementation_guide.md#Distributions) to learn more about distributions in Tanzu CLI

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

## Creating clusters

Tanzu CLI allows you to create clusters on a variety of infrastructure platforms 
such as vSphere, Azure, AWS and on Docker.

1. Initialize the Tanzu kickstart UI by running the below command to create the 
management cluster.
```
tanzu management-cluster create --ui
```

The above would open a management cluster provisioning UI and you can select the
deployment infrastructure and create the cluster.

2. To validate the creation of the management cluster
```
tanzu management-cluster get
```

3. Get the management cluster's kubeconfig
```
tanzu management-cluster kubeconfig get ${MGMT_CLUSTER_NAME} --admin
```

4. Set kubectl context
```
kubectl config use-context ${MGMT_CLUSTER_NAME}-admin@${MGMT_CLUSTER_NAME}
```

5. Next create the workload cluster 
   1. Create a new workload clusterconfig file by copying the management cluster config file
   `~/.config/tanzu/tkg/clusterconfigs/<MGMT-CONFIG-FILE>` and changing the `CLUSTER_NAME` parameter
   to the workload cluster name, you can also edit other parameters as required.
   2. Create workload cluster
   ```
    tanzu cluster create ${WORKLOAD_CLUSTER_NAME} --file ~/.config/tanzu/tkg/clusterconfigs/workload.yaml
   ```
   3. Validate workload cluster creation
   ```
    tanzu cluster list 
   ```
   
6. Do cool things with the provisioned clusters.
7. Clean up

   1. To delete workload cluster
   ```
    tanzu cluster delete ${WORKLOAD_CLUSTER_NAME}
   ```
   Management cluster can only be deleted after deleting all the workload clusters.

   2. To delete management cluster
   ```
    tanzu management-cluster delete ${MGMT_CLUSTER_NAME}
   ```

## What's next

Tanzu CLI is built to be extensible, if you wish to extend Tanzu CLI, you can do
that by writing your CLI plugins.

### Create your own plugin
To bootstrap a new plugin, follow the `builder` plugin documentation [here](../../cmd/cli/plugin-admin/builder/README.md).

Check out the [plugin implementation guide](../cli/plugin_implementation_guide.md) for more details.
