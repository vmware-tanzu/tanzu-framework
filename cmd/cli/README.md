# Tanzu CLI

The [Tanzu CLI](https://docs.google.com/document/d/1btWfZ9Z_Y7SmUmPis69hR_u4TunAPKT7ZDSZy5mvU6Y/edit?usp=sharing)

## Install
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


## Developing
The Tanzu CLI was built to be extensible across teams and be presented across skus.

### Plugins
The CLI is made of plugins. Each plugin is currently located in the `./plugin` directory.   

To add a plugin simply copy the format of an existing one, and utilize the plugin library in `./pkg/cli/commands/plugin`.   

Plugin designs should be approved by the CLI product manager @Morgan Fine, if you wish to expose a plugin in an alpha state please add it to 
the alpha plugin in `./plugin/alpha`.   

Plugins are pulled from registered repositories, on a merge to master all the plugins in this repo are built and pushed to a public repository. When developing its useful to use a local repo.

To build use:
```
make build
```
This will build a local repository under `./artifacts`. Plugins can be installed from this repository using:
```
tanzu plugin install <plugin-name> --local ./artifacts
```

Plugins are installed into `$XDG_DATA_HOME`, [read more](https://specifications.freedesktop.org/basedir-spec/basedir-spec-latest.html)

The CLI can be updated to the latest version of all plugins using:
```
tanzu update
```
An example external plugin repo can be seen at https://gitlab.eng.vmware.com/olympus/cli-plugins

### Distributions

The CLI comes with the notion of a distribution, which is a set of plugins that should always be installed on boot.

This allows the CLI to be presented in accordance with different product offerings. When creating a new local catalog, you can specify the distro you wish the catalog to enforce for the CLI.

On boot, the CLI will check that the distro is present within the given set of plugins or it will install them.