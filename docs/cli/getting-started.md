# Tanzu CLI Getting Started

A simple set of instructions to set up and use the Tanzu CLI.

## Binary installation

### Install the latest release of Tanzu CLI

`linux-amd64`,`windows-amd64`, and `darwin-amd64` are the OS-ARCHITECTURE
combinations we support now.

#### macOS/Linux

- Download the latest [release](https://github.com/vmware-tanzu/tanzu-framework/releases/latest)

- Extract the downloaded tar file

  - for macOS:

    ```sh
    mkdir tanzu && tar -xvf tanzu-framework-darwin-amd64.tar -C tanzu
    ```

  - for Linux:

    ```sh
    mkdir tanzu && tar -xvf tanzu-framework-linux-amd64.tar -C tanzu
    ```

- Install the `tanzu` CLI

  Note: Replace `v0.8.0` with the version you've downloaded.

  - for macOS:

    ```sh
    install tanzu/cli/core/v0.8.0/tanzu-core-darwin_amd64 /usr/local/bin/tanzu
    ```

  - for Linux:

    ```sh
    sudo install tanzu/cli/core/v0.8.0/tanzu-core-linux_amd64 /usr/local/bin/tanzu
    ```

- Set `TANZU_CLI_NO_INIT=true`

  ```sh
  export TANZU_CLI_NO_INIT=true
  ```

- If you have a previous version of tanzu CLI already installed and the config file ~/.config/tanzu/config.yaml is present, run this command to make sure the default plugin repo points to the right path.

  ```sh
  tanzu plugin repo update -b tanzu-cli-framework core
  ```

- Install the downloaded plugins

  ```sh
  tanzu plugin install --local tanzu/cli all
  ```

- Verify installed plugins

  ```sh
  tanzu plugin list
  ```

#### Windows

- Download the latest [release](https://github.com/vmware-tanzu/tanzu-framework/releases/latest)

- Open PowerShell as an administrator, change to the download directory and run:

  ```sh
  mkdir tanzu
  tar -xvf tanzu-framework-windows-amd64.tar -C tanzu
  cd .\tanzu\
  ```

- Save following in `install.bat` in current directory and run `install.bat`

  Note: Replace `v0.8.0` (line number 3) with the version you've downloaded.

  ```sh
  SET TANZU_CLI_DIR=%ProgramFiles%\tanzu
  mkdir "%TANZU_CLI_DIR%"
  copy /B /Y cli\core\v0.8.0\tanzu-core-windows_amd64.exe "%TANZU_CLI_DIR%\tanzu.exe"
  set PATH=%PATH%;%TANZU_CLI_DIR%
  SET PLUGIN_DIR=%LocalAppData%\tanzu-cli
  mkdir %PLUGIN_DIR%
  SET TANZU_CACHE_DIR=%LocalAppData%\.cache\tanzu
  rmdir /Q /S %TANZU_CACHE_DIR%
  set TANZU_CLI_NO_INIT=true
  tanzu plugin repo update -b tanzu-cli-framework core
  tanzu plugin install --local cli all
  tanzu plugin list
  ```

- Add `Program Files\tanzu` to your PATH.

## Delete a selected plugin

If you want to delete a given plugin (one use case is when a plugin has become obsolete), you can run the following command:

```sh
tanzu plugin delete <PLUGIN_NAME>
```

With `v0.8.0` release, the plugin `imagepullsecret` is deprecated and renamed `secret`. The new plugin `secret` will be installed following
the instructions listed above. Remove the installed deprecated plugin if it exists using:

```sh
tanzu plugin delete imagepullsecret
```

## Build the CLI and plugins from source

If you want the very latest, you can also build and install tanzu CLI, and its plugins, from source.

### Prerequisites

- [go](https://golang.org/dl/) version 1.16

- Clone Tanzu Framework and run the below command to build and install CLI and
  plugins locally for your platform.

  ```sh
  TANZU_CLI_NO_INIT=true make build-install-cli-local
  ```

- When the build is done, the tanzu CLI binary and the plugins will be produced locally in the `artifacts` directory.
  The CLI binary will be in a directory similar to the following:

  ```bash
  ./artifacts/<OS>/<ARCH>/cli/core/<version>/tanzu-core-<os_arch>
  ```

- For instance, the following is a build for MacOS:

  ```bash
  ./artifacts/darwin/amd64/cli/core/latest/tanzu-core-darwin_amd64
  ```

- If you additionally want to build and install CLI and plugins for all platforms, run:

  ```sh
  TANZU_CLI_NO_INIT=true make build-install-cli-all
  ```

The CLI currently contains a default distribution which is the default set of plugins that should be installed on
initialization. Initialization of the distributions can be prevented by setting the env var `TANZU_CLI_NO_INIT=true`.
Check out this [doc](../cli/plugin_implementation_guide.md#Distributions) to learn more about distributions in Tanzu CLI

## Usage

```sh
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

   ```sh
   tanzu management-cluster create --ui
   ```

   The above would open a management cluster provisioning UI and you can select the
   deployment infrastructure and create the cluster.

1. To validate the creation of the management cluster

   ```sh
   tanzu management-cluster get
   ```

1. Get the management cluster's kubeconfig

   ```sh
   tanzu management-cluster kubeconfig get ${MGMT_CLUSTER_NAME} --admin
   ```

1. Set kubectl context

   ```sh
   kubectl config use-context ${MGMT_CLUSTER_NAME}-admin@${MGMT_CLUSTER_NAME}
   ```

1. Next create the workload cluster

   1. Create a new workload clusterconfig file by copying the management cluster config file
      `~/.config/tanzu/tkg/clusterconfigs/<MGMT-CONFIG-FILE>` and changing the `CLUSTER_NAME` parameter
      to the workload cluster name, you can also edit other parameters as required.
   1. Create workload cluster

      ```sh
      tanzu cluster create ${WORKLOAD_CLUSTER_NAME} --file ~/.config/tanzu/tkg/clusterconfigs/workload.yaml
      ```

   1. Validate workload cluster creation

      ```sh
      tanzu cluster list
      ```

1. Do cool things with the provisioned clusters.
1. Clean up

   1. To delete workload cluster

      ```sh
      tanzu cluster delete ${WORKLOAD_CLUSTER_NAME}
      ```

      Management cluster can only be deleted after deleting all the workload clusters.

   1. To delete management cluster

      ```sh
      tanzu management-cluster delete ${MGMT_CLUSTER_NAME}
      ```

## What's next

Tanzu CLI is built to be extensible, if you wish to extend Tanzu CLI, you can do
that by writing your CLI plugins.

### Create your own plugin

To bootstrap a new plugin, follow the `builder` plugin documentation [here](../../cmd/cli/plugin-admin/builder/README.md).

Check out the [plugin implementation guide](../cli/plugin_implementation_guide.md) for more details.
