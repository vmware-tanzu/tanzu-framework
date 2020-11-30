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
The CLI is made of plugins. To bootstrap a plugin use the `builder` admin plugin.   

`tanzu builer init <repo-name>` will create a new plugin repository.    

`tanzu builder cli add-plugin <plugin-name>` will add a new cli plugin. 

The CI will publish the plugins to a GCP Bucket with repo name prefix. It expects the the repo secret `GCP_BUCKET_SA` to have a GCP service account token that has write access to that repo. If you need help provisioning this, please ping the #tanzu-cli-api slack. To add your repository to the default set make a PR to this repo and add it to the `KnownRepositories` list.

Plugin designs should go through the plugin review process, if you wish to expose a plugin in an alpha state please add it to 
the alpha plugin in `./plugin/alpha`.   

Plugins are pulled from registered repositories, on a merge to master all the plugins in this repo are built and pushed to a public repository. When developing it's useful to leverage a local repo.

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