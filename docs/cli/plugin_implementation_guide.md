# Tanzu CLI Plugin Implementation Guide

## Developing

The Tanzu CLI was built to be extensible across teams and be cohesive across SKUs. To this end, the Tanzu CLI provides
tools to make creating and compiling new plugins straightforward.

The [Tanzu CLI Styleguide](/docs/cli/style_guide.md) describes the user interaction patterns to be followed,
and general guidance, for CLI contribution.

------------------------------

### Plugins

The Tanzu CLI is a modular design that consists of plugins. To bootstrap a new plugin, you can use the `builder` admin
plugin as described below.

This architecture enables teams to build, own, and release their own piece of functionality as well as enable external
partners to integrate with the system.

Current implementations:

* [Tanzu Framework plugins](https://github.com/vmware-tanzu/tanzu-framework/tree/main/cmd/cli/plugin)
* [Admin plugins](https://github.com/vmware-tanzu/tanzu-framework/tree/main/cmd/cli/plugin-admin)
* Advanced plugins

### Repository (Legacy method)

NOTE: This is not applicable if [context-aware plugin discovery](/docs/design/context-aware-plugin-discovery-design.md) is enabled within Tanzu CLI.

A plugin repository represents a group of plugin artifacts that are installable by the Tanzu CLI. A repository is
defined as an interface.

Current interface implementations:

* Local filesystem
* GCP bucket

Our production plugin artifacts currently come from GCP buckets. Developers of plugins will typically use the local
filesystem repository in the creation and testing of their feature.

#### Adding Admin Repository (Legacy method)

NOTE: This is not applicable if [context-aware plugin discovery](/docs/design/context-aware-plugin-discovery-design.md) is enabled within Tanzu CLI.

The admin repository contains the Builder plugin - a plugin which helps scaffold and compile plugins.

To add the admin repository use `tanzu plugin repo add -n admin -b tanzu-cli-admin-plugins -p artifacts-admin`

To add the builder plugin use `tanzu plugin install builder`

#### Installing Admin Plugins

With the [context-aware plugin discovery](/docs/design/context-aware-plugin-discovery-design.md) enabled, Tanzu CLI admin plugins should now be installed from local source as follows:

1. Download the latest admin plugin tarball or zip file from [release](https://github.com/vmware-tanzu/tanzu-framework/releases/latest) page (`tanzu-framework-plugins-admin-linux-amd64.tar.gz` or `tanzu-framework-plugins-admin-darwin-amd64.tar.gz` or `tanzu-framework-plugins-admin-darwin-amd64.tar.gz`) and extract it (using `linux` as example OS for next steps)
1. Run `tanzu plugin list --local /path/to/admin-plugins/linux-amd64-admin` to list available admin plugins
1. Run `tanzu plugin install all --local /path/to/admin-plugins/linux-amd64-admin` command to install all admin plugins. User can use `_plugin-name_` instead of `all` to install a specific plugin.
1. Run `tanzu plugin list` to verify the installed plugins

NOTE: We are working on enhancing this user experience by publishing admin artifacts as OCI image discovery source as well. More details is in [this issue](https://github.com/vmware-tanzu/tanzu-framework/issues/1376).

### Context

Context is an isolated scope of relevant client-side configurations for a combination of user identity and server identity. There can be multiple contexts for the same combination of `(user, server)`. Previously, this was referred to as `Server` in the Tanzu CLI. Going forward we shall refer to them as `Context` to be explicit.

If a plugin wants to access the context it should use the [provided libraries](/pkg/v1/config/context.go) for forwards compatibility. For example, to get the current active context use the below snippet:

```go
ctx, err := config.GetCurrentContext(v1alpha1.CtxTypeK8s)
```

**Note:** The Tanzu CLI ensures backwards compatibility between `Server` and `Context`.

### Plugin Discovery Source

Discovery is the interface to fetch the list of available plugins, their supported versions and how to download them either standalone or scoped to a context(server). E.g., the CLIPlugin resource in a management cluster, OCI based plugin discovery for standalone plugins, a similar REST API etc. provides the list of available plugins and details about the supported versions. Having a separate interface for discovery helps to decouple discovery (which is usually tied to a server or user identity) from distribution (which can be shared).

Plugins can be of two different types:

  1. Standalone plugins: independent of the CLI context and are discovered using standalone discovery source
  
      This type of plugins are not associated with the `tanzu login` workflow and are available to the Tanzu CLI independent of the CLI context.

  2. Context(server) scoped plugins: scoped to one or more contexts and are discovered using kubernetes or other server associated discovery source

      This type of plugins are associated with the `tanzu login` workflow and are discovered from the management-cluster or global server endpoint.
      In terms of management-clusters, this type of plugins are mostly associated with the installed packages.

      Example:

      As a developer of a `velero` package, I would like to create a Tanzu CLI plugin that can be used to configure and manage installed `velero` package configuration.
      This usecase can be handled with context scoped plugins by installing `CLIPlugin` CR related to `velero` plugin on the management-cluster as part of `velero` package installation.

      ```sh
      # Login to a management-cluster
      $ tanzu login

      # Installs velero package to the management-cluster along with `velero` CLIPlugin resource
      $ tanzu package install velero-pkg --package-name velero.tanzu.vmware.com

      # Plugin list should show a new `velero` plugin available
      $ tanzu plugin list
        NAME     DESCRIPTION                    SCOPE       DISCOVERY          VERSION    STATUS
        velero   Backup and restore operations  Context     cluster-default    v0.1.0     not installed

      # Install velero plugin
      $ tanzu plugin install velero
      ```

Default standalone plugins discovery source automatically gets added to the tanzu config files and plugins from this discovery source are automatically gets discovered.

```sh
$ tanzu plugin list
  NAME                DESCRIPTION                                 SCOPE       DISCOVERY             VERSION      STATUS
  login               Login to the platform                       Standalone  default               v0.11.0-dev  not installed
  management-cluster  Kubernetes management-cluster operations    Standalone  default               v0.11.0-dev  not installed
```

To add the plugin discovery source, assuming the admin plugin's manifests are released as a carvel-package at OCI image `projects.registry.vmware.com/tkg/tanzu-plugins/admin-plugins:v0.11.0-dev` than we use below command to add that discovery source to tanzu configuration.

```sh
 tanzu plugin source add --name admin --type oci --uri projects.registry.vmware.com/tkg/tanzu-plugins/admin-plugins:v0.11.0-dev
```

We can check newly added discovery source with

```sh
$ plugin source list
  NAME     TYPE  SCOPE
  default  oci   Standalone
  admin    oci   Standalone
```

This will allow tanzu CLI to discover new available plugins in the newly added discovery source.

```sh
$ tanzu plugin list
  NAME                DESCRIPTION                                                        SCOPE       DISCOVERY             VERSION      STATUS
  login               Login to the platform                                              Standalone  default               v0.11.0-dev  not installed
  management-cluster  Kubernetes management-cluster operations                           Standalone  default               v0.11.0-dev  not installed
  builder             Builder plugin for CLI                                             Standalone  admin                 v0.11.0-dev  not installed
  test                Test plugin for CLI                                                Standalone  admin                 v0.11.0-dev  not installed
```

To install the builder plugin use `tanzu plugin install builder`

#### Bootstrap A New CLI plugin

`tanzu builder init <repo-name>` will create a new plugin repository.

`cd <repo-name> && tanzu builder cli add-plugin <plugin-name>` will add a `main` package for the new plugin.

CLI plugins have to instantiate a [Plugin descriptor](https://github.com/vmware-tanzu/tanzu-framework/blob/main/apis/cli/v1alpha1/catalog_types.go#L69)
for creating a new plugin, which then bootstraps the plugin with some [sub-commands](https://github.com/vmware-tanzu/tanzu-framework/tree/main/pkg/v1/cli/command/plugin).

Plugins are pulled from registered repositories. On a merge to main, all the plugins in this repo are built and pushed
to a public repository. It is useful to leverage a local repo when developing.

#### Building a Plugin

The Tanzu CLI itself is responsible for building plugins. You can build and install your new plugin for the current host OS with the provided make targets as follows:

1. Initialize go module.

```shell
make init
```

1. Create an initial commit.

```shell
git add -A
git commit -m "Initialize plugin repository"
```

1. Add your plugin name to `PLUGINS` variable in the `Makefile`.

```shell
# Add list of plugins separated by space
PLUGINS ?= "<PLUGIN_NAME>"
```

1. Build the plugin.

```sh
make build-install-local
```

This will build plugin artifacts under `./artifacts` and generates plugin publishing directory `{HOSTOS}/{HOSTARCH}` under `./artifacts/published`.
Using `make build-install-local` installs the plugins for the user, but it internally invokes following command to install the plugins.

```sh
tanzu plugin install <plugin-name> --local ./artifacts/published/${HOSTOS}-${HOSTARCH}
```

If your plugin talks to a Tanzu cluster, a helper function `GetCurrentClusterConfig` defined in
[config](../../pkg/v1/config) package can be used to get the logged in Tanzu cluster's config.

Plugins are installed into `$XDG_DATA_HOME`, [read more](https://specifications.freedesktop.org/basedir-spec/basedir-spec-latest.html)

The CLI can be updated to the latest version of all plugins using:

```sh
tanzu update
```

#### Tests

Every CLI plugin should have a nested test executable. The executable should utilize the test framework found in
`pkg/v1/test/cli`.
Tests are written to ensure the stability of the commands and are compiled alongside the plugins. Tests can be ran by
the admin `test` plugin.

#### Docs

Every plugin requires a README document that explains its basic usage.

### Distributions

A distribution is simply a set of plugins that may exist across multiple repositories. The CLI currently contains a
default distribution which is the default set of plugins that should be installed on initialization.

Distributions allow the CLI to be presented in accordance with different product offerings. When creating a new local
catalog, you can specify the distro you wish the catalog to enforce for the CLI.

On boot, the CLI will check that the distro is present within the given set of plugins or it will install them.

Initialization of the distributions can be prevented by setting the env var `TANZU_CLI_NO_INIT=true`

### Release

When a git tag is created on the repositories, it will version all the plugins in that repository to the current tag.
The plugin binaries built for that tag will be namespaced under the tag semver.

All merges to main will be under the `dev` namespace in the artifacts repository.

The release directory structure for the available plugins under the repository can be generated by running the following command:

```sh
make release
```

This will generate the published artifact directories under `./artifacts/published/` (default location) with OS-ARCH specific
directory structure generated.

------------------------------

## Repositories

Framework exists in
[https://github.com/vmware-tanzu/tanzu-framework](https://github.com/vmware-tanzu/tanzu-framework)
any plugins that are considered open source should exist in that repository as well.

Other repositories should follow the model seen in
(TODO:add example url) and vendor the repository.
Ideally these plugins should exist in the same area as the API definitions.

------------------------------

## CLI Behavior

### Components

CLI commands should utilize the plugin component library in `pkg/cli/component` for interactive features like prompts
or table printing.

### Asynchronous Requests

Commands should be written in such a way as to return as quickly as possible.
When a request is not expected to return immediately, as is often the case with declarative commands, the command should
return immediately with an exit code indicating the server's response.

The completion notice should include an example of the `get` command the user would need in order to poll the resource
to check the state/status of the operation.

### Tab Completion

TBD

### Templates

TBD

### Config file

`~/.config/tanzu/config.yaml`

------------------------------

## CLI Design

Please see the [Tanzu CLI Styleguide](/docs/cli/style_guide.md)
