# Tanzu CLI plugin implementation guide

## Developing
The Tanzu CLI was built to be extensible across teams and be presented across skus.

The [Tanzu CLI Styleguide](/docs/cli/style_guide.md) describes the user interaction patterns to be followed, and general guidance, for CLI contribution. 

------------------------------

### Plugins
The CLI is made of plugins. To add a new plugin copy an existing one in `cmd/cli/plugin` and change the values for the new plugin name.

Plugins are pulled from registered repositories, on a merge to master all the plugins in this repo are built and pushed to a public repository. When developing it's useful to leverage a local repo.

To build use:
```
make build
```
This will build a local repository under `./artifacts`. Plugins can be installed from this repository using:
```
tanzu plugin install <plugin-name> --local ./artifacts -u
```

Plugins are installed into `$XDG_DATA_HOME`, [read more](https://specifications.freedesktop.org/basedir-spec/basedir-spec-latest.html)

The CLI can be updated to the latest version of all plugins using:
```
tanzu update
```

#### Tests
Every CLI plugin should have a nested test executable. The executable should utilize the test framework found in `pkg/v1/test/cli`. Tests should be written 
to cover each command. Tests are compiled alongside the plugins. Tests can be ran by the admin `test` plugin.

#### Docs
Every plugin requires a guide that explains its usage. 

### Distributions
The CLI comes with the notion of a distribution, which is a set of plugins that should always be installed on boot.

This allows the CLI to be presented in accordance with different product offerings. When creating a new local catalog, you can specify the distro you wish the catalog to enforce for the CLI.

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
The core framework exists in https://github.com/vmware-tanzu-private/core any
plugins that are considered open source should exist in that repository as well.

Other repositories should follow the model seen in
(TODO:add example url) and vendor the core repository.
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
~/.tanzu/config.yaml
Issue #263  would move it to ~/.config/tanzu/config.yaml

------------------------------

## CLI Design
Please see the [Tanzu CLI Styleguide](/docs/cli/style_guide.md)
