# CLI Architecture

This document aims to provide a general overview of the Tanzu CLI architecture.

## Definition

_Plugin_ - The CLI consists of plugins, each being a cmd developed in Go and conforming to Cobra CLI standard.

_Context_ - An isolated scope of relevant client-side configurations for a combination of user identity and server identity.

_Target_ - Target is a top level entity used to make the control plane, that a user is interacting against, more explicit in command invocations.

_DiscoverySource_ - Represents a group of plugin artifacts and their distribution details that are installable by the Tanzu CLI.

_Catalog_ - A catalog holds the information of all currently installed plugins on a host OS.

_Distribution_ - A distribution is a set of plugins that may exist across multiple repositories.

_Groups_ - Plugin output is displayed within groups. Currently, Admin, Extra, Run, System, Version.

_Builder_ - Builder scaffolds CLI plugins repositories and new plugins. Builds CLI plugins for the specified arch.

## Plugins

The CLI is based on a plugin architecture. This architecture enables teams to build, own, and release their own piece of functionality as well as enable external partners to integrate with the system.

There are two category of plugins that are determined based on Plugin Discovery. Standalone Plugins and Context-Scoped Plugins

## Plugin Discovery

A plugin discovery points to a group of plugin artifacts that are installable by the Tanzu CLI. It uses an interface to fetch the list of available plugins, their supported versions and how to download them.

There are two types of plugin discovery: Standalone Discovery and Context-Scoped Discovery.

Standalone Discovery: Independent of the CLI context. E.g. OCI based plugin discovery not associated with any context

Context-Scoped Discovery - Associated with a context (generally active context) E.g., the CLIPlugin API in a kubernetes cluster

Standalone Plugins: Plugins that are discovered through standalone discovery source

Context-Scoped Plugins: Plugins that are discovered through context-scoped discovery source

The `tanzu plugin source` command is applicable to standalone plugin discovery only.

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

To list all the available plugin that are getting discovered:

```sh
tanzu plugin list
```

It will list the plugins from all the discoveries found in the local config file.

To describe a plugin use:

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

## Context

Context is an isolated scope of relevant client-side configurations for a combination of user identity and server identity.
There can be multiple contexts for the same combination of `(user, server)`. Previously, this was referred to as `Server` in the Tanzu CLI.
Going forward we shall refer to them as `Context` to be explicit. Also, the context can be managed at one place using the `tanzu context` command.
Earlier, this was distributed between the `tanzu login` command and `tanzu config server` command.

Each `Context` is associated with a `Target` which is used to determine which the control-plane(target) that context is applicable.
More details regarding Target is available in next section.

Note: This is currently behind a feature flag. To enable the flag please run `tanzu config set features.global.context-target true`

Create a new context:

```sh
# Deprecated: Login to TKG management cluster by using kubeconfig path and context for the management cluster
tanzu login --kubeconfig path/to/kubeconfig --context context-name --name mgmt-cluster

# New Command
tanzu context create --management-cluster --kubeconfig path/to/kubeconfig --context path/to/context --name mgmt-cluster
```

List known contexts:

```sh
# Deprecated
tanzu config server list

# New Command
tanzu context list
```

Delete a context:

```sh
# Deprecated
tanzu config server delete demo-cluster

# New Command
tanzu context delete demo-cluster
```

Use a context:

```sh
# Deprecated
tanzu login mgmt-cluster

# New Command
tanzu context use mgmt-cluster
```

## Target

Target is a top level entity used to make the control plane, that a user is interacting against, more explicit in command invocations.
This is done by creating a separate target specific command under root level command for Tanzu CLI. e.g. `tanzu <target>`

The Tanzu CLI supports two targets: `kubernetes`, `mission-control`. This is currently backwards compatible, i.e., the plugins are still available at the root level.

Target of a plugin is determined differently for Standalone Plugins and Context-Scoped plugins.

For Standalone Plugins, the target is determined based on `target` field defined in `CLIPlugin` CR as part of the discovery API.

For Context-scoped Plugins, the target is determined based on the `target` associated with Context itself.
E.g. all plugins discovered through the Context `test-context` will have the same target that is associated with the `test-context`.

List TKG workload clusters using `cluster` plugin associated with `kubernetes` target:

```sh
# Without target grouping (a TKG management cluster is set as the current active server)
tanzu cluster list

# With target grouping
tanzu kubernetes cluster list
```

List TMC workload clusters using `cluster` plugin associated with `tmc` target:

```sh
# With target grouping
tanzu mission-control cluster list
```

## Catalog

A catalog holds the information of all currently installed plugins on a host OS. Plugins are currently stored in $XDG_DATA_HOME/tanzu-cli. Plugins are self-describing and every plugin automatically implements a set of default hidden commands.

```sh
tanzu cluster info
```

Will output the descriptor for that plugin in json format, eg:

```json
{"name":"cluster","description":"Kubernetes cluster operations","version":"v0.0.1","buildSHA":"7e9e562-dirty","group":"Run"}
```

The catalog gets built while installing or upgrading any plugins by executing the info command on the binaries.

## Execution

When the root `tanzu` command is executed it gathers the plugin descriptors from the catalog for all the installed plugins and builds cobra commands for each one.

When this plugin specific commands are invoked, Core CLI simply executes the plugin binary for the associated plugins and passes along stdout/in/err and any environment variables.

## Versioning

By default, versioning is handled by the git tags for the repo in which the plugins are located. Versions can be overridden by setting the version field in the plugin descriptor.

All versions for a given plugin can be found by running:

```sh
tanzu plugin describe <name>
```

When installing or updating plugins a specific version can be supplied:

```sh
tanzu plugin install <name> --version v1.2.3
```

## Groups

With `tanzu --help` command, Plugins are displayed within groups. This enables the user to easily identify what functionality they may be looking for as plugins proliferate.

Currently, updating plugin groups is not available to end users as new groups must be added to Core CLI directly. This was done to improve consistency but may want to be revisited in the future.

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

The builder admin plugin is a means to build Tanzu CLI plugins. Builder provides a set of commands to bootstrap plugin repositories, add commands to them and compile them into an artifacts directory

Initialize a plugin repo:

```sh
tanzu builder init
```

Add a cli command:

```sh
tanzu builder cli add-plugin <name>
```

Compile into an artifact directory.

```sh
tanzu builder cli compile ./cmd/plugins
```

## Release

Plugins are first compiled into an artifact directory (local discovery source) using the builder plugin and then pushed up to their production discovery source.

## Default Plugin Commands

All plugins get several commands bundled with the plugin system, to provide a common set of commands:

* _Lint_: Lints the cobra command structure for flag and command names and shortcuts.
* _Docs_: Every plugin gets the ability to generate its cobra command structure.
* _Describe, Info, Version_: Get the basic details about any plugin.
