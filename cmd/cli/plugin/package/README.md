# Package

Manage package lifecycle operations.

## Usage

Package Operations ( Subject to Change ):

```
>>> tanzu package --help
Tanzu package management

Usage:
  tanzu package [command]

Available Commands:
  install       Install a package
  uninstall     Uninstall a package

Flags:
  -h, --help              help for package
      --log-file string   Log file path
  -v, --verbose int32     Number for the log level verbosity(0-9)

Use "tanzu package [command] --help" for more information about a command.
```

## Test

1. Install a package

Example 1: Install the specified version for package name "fluent-bit.tkg-standard.tanzu.vmware" and while providing the values.yaml file.
```sh

>>> tanzu package install fluentbit --package-name fluent-bit.tkg-standard.tanzu.vmware --namespace test-ns --create-namespace --version 1.7.5-vmware1 --values-file values.yaml
Added installed package 'fluentbit' in namespace 'test-ns'
```

An example values.yaml is as follows:
```sh
#@data/values
#@overlay/match-child-defaults missing_ok=True
---
fluent_bit:
  outputs: |
    [OUTPUT]
      Name     stdout
      Match    *
```

Example 2: Install the latest version for package name "simple-app.corp.com". If the namespace does not exist beforehand, it gets created.
```sh
>>> tanzu package install simple-app --package-name simple-app.corp.com --namespace test-ns-2 --create-namespace
Added installed package 'simple-app' in namespace 'test-ns-2'
```

2. Uninstall a package

```sh
>>> tanzu package uninstall fluentbit --namespace test-ns
Uninstalled package 'fluentbit' in namespace 'test-ns'
```
