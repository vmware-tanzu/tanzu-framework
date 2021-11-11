# CLI Architecture

This document aims to provide a general overview of the Tanzu CLI architecture.

## Definition

_Plugin_ - The CLI consists of plugins, each being a cmd developed in Go and conforming to Cobra CLI standard.

_Repository_ - Represents a group of plugin artifacts that are installable by the Tanzu CLI.

_Catalog_ - A catalog holds the information of all currently installed plugins on a host OS.

_Distribution_ - A distribution is a set of plugins that may exist across multiple repositories.

_Groups_ - Plugin output is displayed within groups. Currently, Admin, Extra, Run, System, Version.

_Builder_ - Builder scaffolds CLI plugins repositories and new plugins. Builds CLI plugins for the specified arch.

## Plugins

The CLI is based on a plugin architecture. This architecture enables teams to build, own, and release their own piece of functionality as well as enable external partners to integrate with the system.

### Current implementations

* Main Tanzu Framework plugins
  * Plugins required to provide the base functionality for the Tanzu CLI.
* Admin plugins
  * Plugins for creating, managing and testing plugins.
* Non-core plugins
  * Optional plugins developed by any non-TKG team.

## Plugin Repositories

A plugin repository represents a group of plugin artifacts that are installable by the Tanzu CLI. A repository is defined as an interface to be implemented by multiple backends like:

* Local filesystem: Your local filesystem can contain a plugin repository, which is particularly useful for development.
* GCP bucket: Published plugins reside in a GCP bucket with a simple manifest.yaml to describe the contents and versions.
* TKG Cluster: We are moving toward a model which will provide plugins as an API, packaged as [imgpkg](https://carvel.dev/imgpkg/) bundles.

Our production plugin artifacts currently come from GCP buckets and every commit to main will produce a set of dev tag plugin to GCP buckets.

Developers of plugins will typically use the local filesystem repository in the creation and testing of their feature.

Releases of plugins are released as tarballs. *Users who installed a versioned tarball release of the CLI are discouraged from consuming plugins directly from buckets.*

The CLI supports multiple repositories. Every plugin source repository produces one or more artifact repositories. During initialization, the CLI stores the known Tanzu repositories in the local config file.

A user can execute to add a repository to the config:

```sh
tanzu plugin repo add -b mybucket -p mypath
```

Then when a user executes:

```sh
tanzu plugin list
```

It will list the plugins from all of the repositories found in the local config file.

To describe a plugin in a repository use:

```sh
tanzu plugin describe <plugin-name>
```

To see specific plugin information:

```sh
tanzu <plugin> info
```

To remove a plugin:

```sh
tanzu plugin delete <plugin-name>
```

This will list all versions of the plugin along with its description.

## Plugin Discovery Sources

Discovery is the interface to fetch the list of available plugins, their supported versions and how to download them either stand-alone or scoped to a server. E.g., the CLIPlugin API in a management cluster, OCI based plugin discovery for standalone plugins, a similar REST API and a manifest file in GCP based discovery, etc. (API is defined [here](apis/config/v1alpha1/clientconfig_types.go#L111-L187)) Unsupported plugins and plugin versions are not returned by the interface. Having a separate interface for discovery helps to decouple discovery (which is usually tied to a server or user identity) from distribution (which can be shared).

The initial proposal of `tanzu plugin source` commands are global (applies for standalone plugin discovery), but if some point in the future (if we get a good use case) we can add a flag for scoping discovery source to a specific context as well.

Adding discovery sources to tanzu configuration file:

```sh
# Add a local discovery source. If URI is relative path,
# $HOME/.config/tanzu-plugins will be considered based path
tanzu plugin source add --name standalone-local --type local --uri path/to/local/discovery

# Add an OCI discovery source. URI should be an OCI image.
tanzu plugin source add --name standalone-oci --type oci --uri projects.registry.vmware.com/tkg/tanzu-plugins/standalone:latest
```

Listing available discovery sources:

```sh
tanzu plugin source list
```

Update a discovery source:

```sh
# Update a local discovery source. If URI is relative path,
# $HOME/.config/tanzu-plugins will be considered based path
tanzu plugin source update standalone-local --type local --uri new/path/to/local/discovery

# Update an OCI discovery source. URI should be an OCI image.
tanzu plugin source update standalone-oci --type oci --uri projects.registry.vmware.com/tkg/tanzu-plugins/standalone:v1.0
```

Delete a discovery source:

```sh
tanzu plugin source delete standalone-oci
```

Sample tanzu configuration file after adding discovery:

```yaml
apiVersion: config.tanzu.vmware.com/v1alpha1
clientOptions:
  cli:
    discoverySources:
    - local:
        name: standalone-local
        path: new/path/to/local/discovery
    - oci:
        name: standalone-oci
        image: projects.registry.vmware.com/tkg/tanzu-plugins/standalone:v1.0
```

## Catalog

A catalog holds the information of all currently installed plugins on a host OS. Plugins are currently stored in $XDG_DATA_HOME/tanzu-cli. Plugins are self-describing and every plugin automatically implements a set of hidden commands.

```sh
tanzu cluster info
```

Will output the descriptor for that plugin in json format, eg:

```json
{"name":"cluster","description":"Kubernetes cluster operations","version":"v0.0.1","buildSHA":"7e9e562-dirty","group":"Run"}
```

The catalog builds itself by executing the info command on all of the binaries found in the configured XDG directory. This data is cached and this cache is managed by the CLI machinery.

Catalogs offer the ability to install a plugin for any given repo or set of repos. As well as updating any plugin for any given a repository.

```sh
tanzu plugin install serverless
```

```sh
tanzu plugin update serverless
```

Catalogs also contain the notion of a set of plugins called a distribution. A distribution is simply a set of plugins that may exist across multiple repositories. The CLI currently contains a default distribution which is the default set of plugins that should be installed on initialization. This is done so that the CLI can be easily tailored to specific company or persona needs.

The above initialization process can be bypassed by setting `TANZU_CLI_NO_INIT=true` during runtime or with a linker flag during build time.

## Components

CLI components aim to be the [Clarity of CLIs](https://clarity.design/), providing a common set of reusable functionality for plugin implementations. By standardizing on these components we ensure consistent UX throughout the product and make it easy to make changes to the experience across all plugins.

Currently implemented components:

* Prompt
* Select
* Table printing
* Question

## Execution

When the root command is executed it gathers the plugin descriptors from all the binaries in the plugin directory and builds cobra commands for each one. For installation, for each configured repository the plugin manifest is consumed to provide install and version information. Once installed, the plugins are saved to a local PluginDescriptor cache. This cache is what is referenced during tanzu plugin list, as well.

Those commands are added to the root command alongside any commands in the core binary. Each cobra command simply executes the binary its associated with and passes along stdout/in/err and any environment variables.

## Versioning

By default versioning is handled by the git tags for the repo in which the plugins are located. If no tag is present the version defaults to ‘dev’, versions can be overridden by setting the version field in the plugin descriptor.

All versions for a given plugin can be found by running:

```sh
tanzu plugin describe <name>
```

When installing or updating plugins a specific version can be supplied:

```sh
tanzu plugin install <name> --version v1.2.3
```

Or a version selection algorithm can be used:

```sh
tanzu update --include-unstable
```

A version selector is simply an interface which finds a version in a set of versions. The current implementations are:

* LatestStable -- find the latest stable version
* LatestAny -- find the latest version including any unstable versions

An unstable version is “dev” or a semver string containing a -suffix after the vMajor.Minor.Patch (e.g. v1.3.0-rc.1, v1.3.0-latest)

Conversely, a stable version is one that does not contain such a suffix.

## Groups

Plugins are displayed within groups. This enables the user to easily identify what functionality they may be looking for as plugins proliferate.

Currently updating plugin groups is not available to end users as new groups must be added to Framework directly. This was done to improve consistency but may want to be revisited in the future.

## Testing

Every plugin requires a test which the compiler enforces. Plugin tests are a nested binary under the plugin which should implement the test framework.

Plugin tests can be run by installing the admin test plugin, which provides the ability to run tests for any of the currently installed plugins. It will fetch the test binaries for each plugin from its respective repo.

Execute the test plugin:

```sh
tanzu test plugin <name>
```

## Docs

Every plugin requires a README.md file at the top level of its directory which is enforced by the compiler. This file should serve as a guide for how the plugin is to be used.

In the future, we should have every plugin implement a `docs` command which outputs the generated cobra docs.

## Builder

The builder admin plugin is a means to build Tanzu products. Builder provides a set of commands to bootstrap plugin repositories, add commands to them and compile them into an artifacts repository

Initialize a plugin repo:

```sh
tanzu builder init
```

Add a cli command:

```sh
tanzu builder cli add-plugin <name>
```

Compile into an artifact repository.

```sh
tanzu builder cli compile ./cmd/plugins
```

## Release

Plugins are first compiled into an artifact repository using the builder plugin and then pushed up to their production repository (currently GCP buckets) using the repos CI mechanism.

The CI is triggered when a tag is created which pushes up a production release for that tag, as well as on merges to main a release is triggered which pushes the artifacts to the ‘dev’ path.

For air gapped releases the Cayman build system is used to produce tarballs which contain the artifacts repositories as well as a script which installs the needed plugins from these local repos.

## Default Plugin Commands

All plugins get several commands bundled with the plugin system, to provide a common set of commands:

* _Lint_: Lints the cobra command structure for flag and command names and shortcuts.
* _Docs_: Every plugin gets the ability to generate its cobra command structure.
* _Describe, Info, Version_: Get the basic details about any plugin.
