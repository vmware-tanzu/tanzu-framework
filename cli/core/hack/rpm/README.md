# Using YUM/DNF to install the Tanzu CLI

YUM and DNF (the replacement for YUM) use RPM packages for installation.  This document describes how to build
such a package for the Tanzu CLI and how to install it.

## Building the RPM package

Executing the `hack/rpm/build_package.sh` script will build the RPM package under `cli/core/hack/rpm/_output`.
The `hack/rpm/build_package.sh` script is meant to be run on a Linux machine that has `rpm` installed.
This can be done in docker.  To facilitate this operation, the new `rpm-package` Makefile target has been added
to `cli/core/Makefile`; this Makefile target will first start a docker container and then run the `hack/rpm/build_package.sh` script.

```bash
$ cd tanzu-framework
$ cd cli/core
$ make rpm-package
```

## Installing the Tanzu CLI using the built RPM package

Installing the Tanzu CLI using the newly built RPM package can be done on a Linux machine with `yum` or `dnf` installed
or a docker container.  For example:

```bash
$ cd tanzu-framework
$ docker run -it -v $(pwd):/tmp/tanzu-framework ubuntu
root$ apt-get install -f /tmp/tanzu-framework/cli/core/hack/apt/_output/tanzu-cli_0.28.0-dev_linux_amd64.deb 
```
