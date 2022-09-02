# feature

Feature plugin lets you operate on Features and FeatureGates

## Usage

Feature plugin has 3 commands

1. list - allows to list the features that are gated by a particular
   FeatureGate.
2. activate - allows to activate a feature.
3. deactivate - allows to deactivate a feature.

By default, Feature plugin operates on Features that are gated by `tkg-system`
FeatureGate, but that can be changed by specifying `featuregate` flag.

Ex:

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
List Features

Usage:
  tanzu feature list [flags]

Examples:
  
    # List a clusters Features
    tanzu feature list --activated
    tanzu feature list --unavailable
    tanzu feature list --deactivated

Flags:
  -a, --activated            List only activated Features
  -d, --deactivated          List only deactivated Features
  -e, --extended             Include extended output. Higher latency as it requires more API calls.
  -f, --featuregate string   List Features gated by a particular FeatureGate (default "tkg-system")
  -h, --help                 help for list
  -o, --output string        Output format (yaml|json|table)
  -u, --unavailable          List only Features specified in the gate but missing from cluster
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
