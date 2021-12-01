# Context-aware API-driven Plugin Discovery

## Abstract

The Tanzu CLI is an amalgamation of all the Tanzu infrastructure elements under one unified core CLI experience. The core CLI supports a plugin model where the developers of different Tanzu services (bundled or SaaS) can distribute plugins that target functionalities of the services they own. When users switch between different services via the CLI context, we want to surface only the relevant plugins for the given context for a crisp user experience.

## Key Concepts

- CLI - The Tanzu command line interface, built to be pluggable.
- Service - Any tanzu service, user-managed or SaaS. E.g., TKG, TCE, TMC, etc
- Server - An instance of service. E.g., A single TKG management-cluster, a specific TMC endpoint, etc.
- Context - an isolated scope of relevant client-side configurations for a combination of user identity and server identity. There can be multiple contexts for the same combination of {user, server}. This is currently referred to as `Server` in the Tanzu CLI, which can also mean an instance of a service. Hence, we shall use `Context` to avoid confusion.
- Plugin - A scoped piece of functionality that can be used to extend the CLI. Usually refers to a single binary that is invoked by the root Tanzu CLI.
- Scope - the context association level of a plugin
- Stand-alone - independent of the CLI context
- Context-scoped - scoped to one or more contexts
- Discovery - the interface to fetch the list of available plugins and their supported versions
- Distribution - the interface to deliver a plugin for user download
- Scheme - the specific mechanism to discover or download a plugin
- Discovery Scheme - e.g., REST API, CLIPlugin kubernetes API, manifest YAML
- Distribution Scheme - e.g., OCI image, Google Cloud Storage, S3
- Discovery Source - the source server of a plugin metadata for discovery, e.g., a REST API server, a management cluster, a local manifest file, OCI compliant image containing manifest file
- Distribution Repository - the repository of plugin binary for distribution, e.g., an OCI compliant image registry, Google Cloud Storage, an S3 compatible object storage server
- Plugin Descriptor - the metadata about a single plugin version that is installed locally and is used by the core to construct a sub-command under `tanzu`

## Background

## Goals

## Non Goals

## High-Level Design

## Detailed Design

## Alternatives Considered

## Security Considerations

## Compatibility

## Implementation

### Installing Plugins from Local Source

Generally we expect most of the users to install the plugins from the default OCI based discovery mechanism. However tanzu CLI handles another way of installing plugins by downloading `tar.gz` files for the plugins and using that to list and install available plugins.

To install the plugin with local source, download the plugin `tar.gz` from the release artifacts for your distribution and untar it to a location on your local machine. You can use the directory where the `tar.gz` has been extracted with the `tanzu plugin list --local` and `tanzu plugin install --local` command as mentioned below.

- List the plugins without `--local`: Lists all available default plugins from the default OCI registry

  ```sh
  $ tanzu plugin list
    NAME                DESCRIPTION                                                        SCOPE       DISCOVERY  VERSION      STATUS
    login               Login to the platform                                              Standalone  default    v0.13.0-dev  not installed
    management-cluster  Kubernetes management-cluster operations                           Standalone  default    v0.13.0-dev  not installed
    package             Tanzu package management                                           Standalone  default    v0.13.0-dev  not installed
    pinniped-auth       Pinniped authentication operations (usually not directly invoked)  Standalone  default    v0.13.0-dev  not installed
    secret              Tanzu secret management                                            Standalone  default    v0.13.0-dev  not installed
  ```

- List the plugins with `--local` pointing to local plugin directory: Lists only available plugins from given local file path

  ```sh
  $ tanzu plugin list --local /tmp/admin/tanzu-plugins/
    NAME     DESCRIPTION                 SCOPE       DISCOVERY  VERSION      STATUS
    builder  Build Tanzu components      Standalone             v0.13.0-dev  not installed
    codegen  Tanzu code generation tool  Standalone             v0.13.0-dev  not installed
    test     Test the CLI                Standalone             v0.13.0-dev  not installed
  ```

- Install plugins `--local` flag: Only installs plugins from given local file path and doesn't install plugins from default OCI registry

  ```sh
  $ tanzu plugin install all --local /tmp/admin/tanzu-plugins/
  Installing plugin 'builder:v0.13.0-dev'
  Installing plugin 'codegen:v0.13.0-dev'
  Installing plugin 'test:v0.13.0-dev'
  Successfully installed all plugins
  ✔  successfully installed 'all' plugin
  ```

- List all installed plugins: Plugins installed with `tanzu plugin install all --local /tmp/admin/tanzu-plugins/` will not display `DISCOVERY` information, but will be listed as part of `list` command output

  ```sh
  $ tanzu plugin list
    NAME                DESCRIPTION                                                        SCOPE       DISCOVERY  VERSION      STATUS
    login               Login to the platform                                              Standalone  default    v0.13.0-dev  not installed
    management-cluster  Kubernetes management-cluster operations                           Standalone  default    v0.13.0-dev  not installed
    package             Tanzu package management                                           Standalone  default    v0.13.0-dev  not installed
    pinniped-auth       Pinniped authentication operations (usually not directly invoked)  Standalone  default    v0.13.0-dev  not installed
    secret              Tanzu secret management                                            Standalone  default    v0.13.0-dev  not installed
    builder             Build Tanzu components                                             Standalone             v0.13.0-dev  installed
    codegen             Tanzu code generation tool                                         Standalone             v0.13.0-dev  installed
    test                Test the CLI                                                       Standalone             v0.13.0-dev  installed
  ```

- List plugins with `--local` that are discovered with default discovery. Let's assume `/tmp/default-package-secret/tanzu-plugins/` contains `package` and `secret` plugins

  ```sh
  $ tanzu plugin list --local /tmp/default-package-secret/tanzu-plugins/
    NAME                DESCRIPTION                  SCOPE       DISCOVERY  VERSION      STATUS
    package             Tanzu package management     Standalone             v0.12.0-dev  not installed
    secret              Tanzu secret management      Standalone             v0.13.0-dev  not installed
  ```

- Install plugins with `--local` that are discovered with default discovery

  ```sh
  $ tanzu plugin install all --local /tmp/default-package-secret/tanzu-plugins/
  Installing plugin 'package:v0.12.0-dev'
  Installing plugin 'secret:v0.13.0-dev'
  Successfully installed all plugins
  ✔  successfully installed 'all' plugin
  ```

- List all installed plugins: As `package` and `secret` plugins can also be discovered with default discovery it will be displayed as `installed` with discovery information mentioned. Status of `package` plugin should list `update available` as the installed version is `v0.12.0-dev` but version `v0.13.0-dev` is discovered from the default discovery

  ```sh
  $ tanzu plugin list
    NAME                DESCRIPTION                                                        SCOPE       DISCOVERY  VERSION      STATUS
    login               Login to the platform                                              Standalone  default    v0.13.0-dev  not installed
    management-cluster  Kubernetes management-cluster operations                           Standalone  default    v0.13.0-dev  not installed
    package             Tanzu package management                                           Standalone  default    v0.12.0-dev  update available
    pinniped-auth       Pinniped authentication operations (usually not directly invoked)  Standalone  default    v0.13.0-dev  not installed
    secret              Tanzu secret management                                            Standalone  default    v0.13.0-dev  installed
    builder             Build Tanzu components                                             Standalone             v0.13.0-dev  installed
    codegen             Tanzu code generation tool                                         Standalone             v0.13.0-dev  installed
    test                Test the CLI                                                       Standalone             v0.13.0-dev  installed
  ```

- Running `tanzu plugin sync` will sync all the plugins that can be discovered with the discovery. In this case, `login`, `management-cluster`, `pinniped-auth` will get installed. `package` plugin will get updated to newly available version. `secret` plugin installation will be skipped as no new version is available. And `builder`, `codegen`, `test` plugins will not get considered with sync command as there is no `discovery` source associated with this plugins

  ```sh
  $ tanzu plugin sync
  Installing plugin 'login:v0.13.0-dev'
  Installing plugin 'management-cluster:v0.13.0-dev'
  Installing plugin 'package:v0.13.0-dev'
  Installing plugin 'pinniped-auth:v0.13.0-dev'
  Successfully installed all plugins
  ✔  successfully installed 'all' plugin
  ```

- Clean all installed plugins with `tanzu plugin clean` will delete all installed plugins

- Listing plugin after deleting all plugin will not show locally installed plugins because those plugins did not had any `discovery` source associated with them

  ```sh
  $ tanzu plugin list
    NAME                DESCRIPTION                                                        SCOPE       DISCOVERY  VERSION      STATUS
    login               Login to the platform                                              Standalone  default    v0.13.0-dev  not installed
    management-cluster  Kubernetes management-cluster operations                           Standalone  default    v0.13.0-dev  not installed
    package             Tanzu package management                                           Standalone  default    v0.13.0-dev  not installed
    pinniped-auth       Pinniped authentication operations (usually not directly invoked)  Standalone  default    v0.13.0-dev  not installed
    secret              Tanzu secret management                                            Standalone  default    v0.13.0-dev  not installed
  ```

## Open Issues
