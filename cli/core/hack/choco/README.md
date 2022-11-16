# Using Chocolatey to install the Tanzu CLI

This document describes how to build a Chocolatey package for the Tanzu CLI and how to install it.

## Building the Chocolatey package

Executing the `hack/choco/build_package.sh` script will build the Chocolatey package under `cli/core/hack/choco/_output/choco`.
The `hack/choco/build_package.sh` script is meant to be run on a Linux machine that has `choco` installed.
This is most easily done using docker.  Note that currently, the docker images for Chocolatey only support an
`amd64` architecture.  To facilitate building the package, the new `choco-package` Makefile target has been added
to `cli/core/Makefile`; this Makefile target will first start a docker container and then run the `hack/choco/build_package.sh`
script.  The `VERSION` environment variable must be set when running the make target.

```bash
cd tanzu-framework/cli/core
VERSION=v0.26.0 make choco-package
```

### Content of Chocolatey package

Currently, we build a Chocolatey package without including the actual Tanzu CLI binary.  Instead, when the
package is installed, Chocolatey will download the CLI binary from Github.  This has to do with distribution
rights as we will probably publish the Chocolatey package in the community package repository.

## Installing the Tanzu CLI using the built Chocolatey package

Installing the Tanzu CLI using the newly build Chocolatey package can be done on a Windows machine with `choco`
installed. First, the Chocolatey package must be uploaded to the Windows machine.

For example, if we upload the package to the Windows machine under `$HOME\tanzu-cli.0.26.0.nupkg`, we can then simply do:

```bash
choco install -f "$HOME\tanzu-cli.0.26.0.nupkg"
```

It is also possible to configure a local repository containing the local package:

```bash
choco source add -n=local -s="file://$HOME"
choco install tanzu-cli
```

## Uninstalling the Tanzu CLI

To uninstall the Tanzu CLI after it has been install with Chocolatey:

```bash
choco uninstall tanzu-cli
```

## Publishing the package

Once the Tanzu CLI is ready for full availability, we expect to publish our Chocolatey packages to the main Chocolatey community package repository.  This step remains to be properly defined.
