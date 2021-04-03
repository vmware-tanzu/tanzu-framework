# Builder

Scaffolds and builds Tanzu plugin repositories

## Usage

### Init

`tanzu builder init <repo-name>` will initialize a new plugin repository with scaffolding for:

- Core CLI integration
- GolangCI linting config
- Github and Gitlab CI config
- A Makefile

For more details, this command supports a `--dry-run` flag which will show everything created:
```
tanzu builder init <repo-name> --dry-run
```

### Add-plugin

`tanzu builder cli add-plugin <plugin-name>` adds a new plugin to your repository. The plugins command will live in the ./cmd/plugin/<plugin-name> directory.

### Compile

`tanzu builder cli compile` will compile a repository and create the artifacts to be used with tanzu cli.

Plugins will find that their `make build` command will suffice for most compile cases, but there are many flags at your disposal as well:

```
--artifacts string   path to output artifacts (default "artifacts")
--corepath string    path for core binary
--ldflags string     ldflags to set on build
--match string       match a plugin name to build, supports globbing (default "*")
--path string        path of the plugins directory (default "./cmd/cli/plugin")
--target string      only compile for a specific target, use 'local' to compile for host os (default "all")
--version string     version of the root cli (required)
```
