# Using APT to install the Tanzu CLI

APT uses Debian packages for installation.  This document describes how to build such packages
for the Tanzu CLI, how to push them to a public repository and how to install the CLI from that repository.

## Building the Debian package

Executing the `hack/apt/build_package.sh` script will build the Debian packages under `cli/core/hack/apt/_output`.
The `hack/apt/build_package.sh` script is meant to be run on a Linux machine that has `apt` installed.
This can be done in docker.  To facilitate this operation, the new `apt-package` Makefile target has been added
to `cli/core/Makefile`; this Makefile target will first start a docker container and then run the `hack/apt/build_package.sh` script.

```bash
cd tanzu-framework/cli/core
make apt-package
```

Note that two packages will be built, one for AMD64 and one for ARM64.
Also, a repository will be generated as a directory called `_output/apt` which will contain the two
built packages.  Please see the section on publishing the repository for more details.

## Testing the installation of the Tanzu CLI locally

We can install the Tanzu CLI using the newly built Debian repository locally on a Linux machine with `apt` installed
or using a docker container.  For example:

```bash
$ cd tanzu-framework
$ docker run --rm -it -v $(pwd)/cli/core/hack/apt/_output/apt:/tmp/apt ubuntu
echo "deb file:///tmp/apt jessie main" | tee /etc/apt/sources.list.d/tanzu.list
apt-get update --allow-insecure-repositories
apt install -y tanzu-cli --allow-unauthenticated
tanzu
```

Note that the repository isn't signed at the moment, so you may see warnings during installation.

## Publishing the package to GCloud

We have a GCloud bucket dedicated to hosting the Tanzu CLI OS packages.  That bucket can be controlled from:
`https://console.cloud.google.com/storage/browser/tanzu-cli-os-packages`.

To publish the repository containing the new debian packages for the Tanzu CLI, we must upload the entire `apt`
directory to the root of the bucket.  You can do this manually.  Once uploaded, the Tanzu CLI can be installed
publicly as described in the next section.

## Installing the Tanzu CLI

Currently, the repo is not signed but will be in the future; you may get warnings.
To install from an insecure repo:

```bash
$ docker run --rm -it ubuntu
apt update
apt install -y ca-certificates
echo "deb https://storage.googleapis.com/tanzu-cli-os-packages/apt jessie main" | tee /etc/apt/sources.list.d/tanzu.list
apt update --allow-insecure-repositories
apt install -y tanzu-cli --allow-unauthenticated
```
