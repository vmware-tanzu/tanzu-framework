# Tanzu CLI Plugin Implementation Guide

## Developing
The Tanzu CLI was built to be extensible across teams and be cohesive across SKUs. To this end, the Tanzu CLI provides tools to make creating and compiling new plugins straightforward.

The [Tanzu CLI Styleguide](/docs/cli/style_guide.md) describes the user interaction patterns to be followed, and general guidance, for CLI contribution.

------------------------------

### Plugins
The Tanzu CLI is a modular design that consists of plugins. To bootstrap a new plugin, you can use the `builder` admin plugin as described below.

This architecture enables teams to build, own, and release their own piece of functionality as well as enable external partners to integrate with the system.

Current implementations:
- [Tanzu Framework plugins](https://github.com/vmware-tanzu/tanzu-framework/tree/main/cmd/cli/plugin)
- [Admin plugins](https://github.com/vmware-tanzu/tanzu-framework/tree/main/cmd/cli/plugin-admin)
- [Advanced plugins](https://gitlab.eng.vmware.com/tanzu/cli-plugins)

### Repository
A plugin repository represents a group of plugin artifacts that are installable by the Tanzu CLI. A repository is defined as an interface.

Current interface implementations:
- Local filesystem
- GCP bucket

Our production plugin artifacts currently come from GCP buckets. Developers of plugins will typically use the local filesystem repository in the creation and testing of their feature.

#### Adding Admin Repository
The admin repository contains the Builder plugin - a plugin which helps scaffold and compile plugins.

To add the admin repository use `tanzu plugin repo add -n admin -b tanzu-cli-admin-plugins -p artifacts-admin`

To add the builder plugin use `tanzu plugin install builder`

#### Bootstrap A New CLI plugin
`tanzu builder init <repo-name>` will create a new plugin repository.

`tanzu builder cli add-plugin <plugin-name>` will add a new cli plugin.

Plugins are pulled from registered repositories. On a merge to main, all the plugins in this repo are built and pushed to a public repository.
It is useful to leverage a local repo when developing.

#### PostInstallHook Implementation for a Plugin
A plugin might need to setup initial configuration once plugin is installed. Tanzu CLI exposes this functionality with PluginDescriptor by providing `PostInstallHook` function implementation.
Note: The same function will be invoked with `tanzu config init` command for all the installed plugin as well.

Sample usage with the `management-cluster` plugin: https://github.com/vmware-tanzu/tanzu-framework/blob/main/cmd/cli/plugin/managementcluster/main.go

#### Building a Plugin

The Tanzu CLI itself is responsible for building plugins. You can build your new plugin with the provided make targets:
```
make build
```
This will build plugin artifacts under `./artifacts`. Plugins can be installed from this repository using:
```
tanzu plugin install <plugin-name> --local ./artifacts -u
```

Plugins are installed into `$XDG_DATA_HOME`, [read more](https://specifications.freedesktop.org/basedir-spec/basedir-spec-latest.html)

The CLI can be updated to the latest version of all plugins using:
```
tanzu update
```

#### Tests
Every CLI plugin should have a nested test executable. The executable should utilize the test framework found in `pkg/v1/test/cli`.
Tests are written to ensure the stability of the commands and are compiled alongside the plugins. Tests can be ran by the admin `test` plugin.

#### Docs
Every plugin requires a README document that explains its basic usage.

### Distributions
A distribution is simply a set of plugins that may exist across multiple repositories. The CLI currently contains a default distribution
which is the default set of plugins that should be installed on initialization.

Distributions allow the CLI to be presented in accordance with different product offerings. When creating a new local catalog, you can specify
the distro you wish the catalog to enforce for the CLI.

On boot, the CLI will check that the distro is present within the given set of plugins or it will install them.

Initialization of the distributions can be prevented by setting the env var `TANZU_CLI_NO_INIT=true`

### Release
When a git tag is created on the repositories, it will version all the plugins in that repository to the current tag. The plugin binaries built for that
tag will be namespaced under the tag semver.

All merges to main will be under the `dev` namespace in the artifacts repository.

When listing or installing plugins, a `version finder` is used to parse the available versions of the plugin. By defaultc the version finder will attempt to
find the latest stable semver, which excludes semvers with build suffixes e.g. `1.2.3-rc.1`. If you wish to include unstable builds you can use the `--include-unstable` flag which will look for the latest version regardless of build suffixes.

------------------------------

## Repositories
Framework exists in https://github.com/vmware-tanzu/tanzu-framework any
plugins that are considered open source should exist in that repository as well.

Other repositories should follow the model seen in
(TODO:add example url) and vendor the repository.
Ideally these plugins should exist in the same area as the API definitions.

------------------------------

## CLI Behavior
### Components
CLI commands should utilize the plugin component library in `pkg/cli/component` for interactive features like prompts or table printing.

### Asynchronous Requests
Commands should be written in such a way as to return as quickly as possible.
When a request is not expected to return immediately, as is often the case with declarative commands, the command should return immediately with an exit code indicating the server's response.

The completion notice should include an example of the `get` command the user would need in order to poll the resource to check the state/status of the operation.

### Tab Completion
TBD

### Templates
TBD

### Config file
~/.config/tanzu/config.yaml

------------------------------

## CLI Design
Please see the [Tanzu CLI Styleguide](/docs/cli/style_guide.md)
