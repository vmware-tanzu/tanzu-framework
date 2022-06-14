# Builder

Scaffolds and builds Tanzu plugin repositories

## Usage

### Init

`tanzu builder init <repo-name>` will initialize a new plugin repository with scaffolding for:

* Tanzu Framework CLI integration
* GolangCI linting config
* GitHub or GitLab CI config
* A Makefile

For more details, this command supports a `--dry-run` flag which will show everything created:

```sh
tanzu builder init <repo-name> --dry-run
```

### Add-plugin

`tanzu builder cli add-plugin <plugin-name>` adds a new plugin to your repository. The plugins command will live in the `./cmd/plugin/<plugin-name>` directory.

### Compile

`tanzu builder cli compile` will compile a repository and create the artifacts to be used with tanzu cli.

The artifact output directory structure will be created to match the expected layout. This will include some plugin
metadata used in the publishing and installation of plugins in a `manifest.yaml` file and a `plugin.yaml` file for
each included plugin.

Plugins will find that their `make build` command will suffice for most compile cases, but there are many flags at your disposal as well:

```txt
--artifacts string   path to output artifacts (default "artifacts")
--corepath string    path for core binary
--ldflags string     ldflags to set on build
--match string       match a plugin name to build, supports globbing (default "*")
--path string        path of the plugins directory (default "./cmd/cli/plugin")
--target string      only compile for a specific target, use 'local' to compile for host os (default "all")
--version string     version of the root cli (required)
```

### Publish

`tanzu builder publish` is used to publish the compiled plugin to a target output. The two supported targets for
publishing are "local" for the legacy local filesystem output, or "oci" for the recommended OCI image bundle.

Arguments to the `publish` command are:

```txt
--input-artifact-dir string                  artifact directory which is a output of 'tanzu builder cli compile' command
--local-output-discovery-dir string          local output directory where CLIPlugin resource yamls for discovery will be placed. Applicable to 'local' type
--local-output-distribution-dir string       local output directory where plugin binaries will be placed. Applicable to 'local' type
--oci-discovery-image string                 image path to publish oci image with CLIPlugin resource yamls. Applicable to 'oci' type
--oci-distribution-image-repository string   image path prefix to publish oci image for plugin binaries. Applicable to 'oci' type
--os-arch string                             list of os-arch (default "darwin-amd64 linux-amd64 windows-amd64")
--plugins string                             list of plugin names. Example: 'login management-cluster cluster'
--type string                                type of discovery and distribution for publishing plugins. Supported: local
--version string                             recommended version of the plugins
```
