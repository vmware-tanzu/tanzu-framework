# Unstable Versions

The Tanzu CLI offers the ability to filter CLI plugins based on the Semantic Version of the plugin. By default, the CLI will filter any non-stable versions of plugins from being used but this behavior can be updated by config options. To update, use the config command:
```
tanzu config set unstable-versions all
```

The options for the command are:
- `none`: Allows only stable plugin versions. Default.
- `alpha`: All stable plugins and plugins with a prerelease tag matching "alpha" are allowed.
- `experimental`: All stable, alpha and all prerelease versions are allowed, minus any build tagged versions.
- `all`: Allows all plugin versions, regardless of prerelease or build tags.
