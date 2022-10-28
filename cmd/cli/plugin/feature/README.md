# Feature

Feature plugin gives access to features using featuregates.

## Usage

Feature plugin has three commands:

1. list - allows to list the features that are gated by a particular
   FeatureGate.
2. activate - allows to activate a feature.
3. deactivate - allows to deactivate a feature.

Feature plugin is able to list all discoverable features on the cluster.
Optionally, a FeatureGate may be specified by using the `featuregate` flag.

Example:

```sh
# list Features associated with tkg-system FeatureGate
tanzu feature list
# list Features associated with tkg-system-sample FeatureGate
tanzu feature list --featuregate=tkg-system-sample
```

```sh
>>> tanzu feature --help
Operate on features and featuregates

Usage:
  tanzu feature [command]

Available Commands:
  activate      Activate Features
  deactivate    Deactivate Features
  list          List Features

Flags:
  -h, --help   help for feature

Use "tanzu feature [command] --help" for more information about a command.
```

### list command

```sh
>>> tanzu feature list --help
List features

Usage:
  tanzu feature list [flags]

Examples:
  
    # List feature(s) in the cluster.
    tanzu feature list --activated
    tanzu feature list --deactivated

Flags:
  -a, --activated                             List only activated features
  -d, --deactivated                           List only deactivated features
  -e, --extended                              Include extended output
  -f, --featuregate string                    List features gated by a particular FeatureGate
  -h, --help                                  Help for list
  -x, --include-experimental                  Allows displaying experimental features
  -o, --output string                         Output format (yaml|json|table)
```

### activate command

```sh
>>> tanzu feature activate --help
Activate Features

Usage:
  tanzu feature activate <feature> [flags]

Examples:
  
    # Activate a cluster Feature
    tanzu feature activate myfeature

Flags:
  -f, --featuregate string   Activate a Feature gated by a particular FeatureGate (default "tkg-system")
  -h, --help                 help for activate
```

### deactivate command

```sh
>>> tanzu feature deactivate --help
Deactivate Features

Usage:
  tanzu feature deactivate <feature> [flags]

Examples:
  
    # Deactivate a cluster Feature
    tanzu feature deactivate myfeature

Flags:
  -f, --featuregate string   Deactivate Feature gated by a particular FeatureGate (default "tkg-system")
  -h, --help                 help for deactivate
```
