# Using YUM/DNF to install the Tanzu CLI

YUM and DNF (the replacement for YUM) use RPM packages for installation.  This document describes how to
build such packages for the Tanzu CLI, how to push them to a public repository and how to install the CLI
from that repository.

## Building the RPM package

Executing the `hack/rpm/build_package.sh` script will build the RPM packages under `cli/core/hack/rpm/_output`.
The `hack/rpm/build_package.sh` script is meant to be run on a Linux machine that has `dnf` or `yum` installed.
This can be done in docker using the `fedora` image.  To facilitate this operation, the new `rpm-package`
Makefile target has been added to `cli/core/Makefile`; this Makefile target will first start a docker
container and then run the `hack/rpm/build_package.sh` script.

```bash
cd tanzu-framework/cli/core
make rpm-package
```

Note that two packages will be built, one for AMD64 and one for ARM64.
Also, a repository will be generated as a directory called `_output/rpm` which will contain the two
built packages as well as some metadata.  Please see the section on publishing the repository for more details.

## Testing the installation of the Tanzu CLI locally

We can install the Tanzu CLI using the newly built RPM repository locally on a Linux machine with `yum` or `dnf`
installed or using a docker container.  For example, using `yum`:

```bash
cd tanzu-framework
docker run --rm -it -v $(pwd)/cli/core/hack/rpm/_output/rpm:/tmp/rpm fedora
cat << EOF | sudo tee /etc/yum.repos.d/tanzu-cli.repo
[tanzu-cli]
name=Tanzu CLI
baseurl=file:///tmp/rpm
enabled=1
gpgcheck=0
EOF
yum install -y tanzu-cli
tanzu
```

Note that the repository isn't signed at the moment, so you may see warnings during installation.

## Publishing the package to GCloud

We have a GCloud bucket dedicated to hosting the Tanzu CLI OS packages.  That bucket can be controlled from:
`https://console.cloud.google.com/storage/browser/tanzu-cli-os-packages`.

To publish the repository containing the new rpm packages for the Tanzu CLI, we must upload the entire `rpm`
directory to the root of the bucket.  You can do this manually.  Once uploaded, the Tanzu CLI can be installed
publicly as described in the next section.

## Installing the Tanzu CLI

Currently, the repo is not signed but will be in the future; you may get warnings during installation.
To install from an insecure repo:

```bash
docker run --rm -it fedora
cat << EOF | sudo tee /etc/yum.repos.d/tanzu-cli.repo
[tanzu-cli]
name=Tanzu CLI
baseurl=https://storage.googleapis.com/tanzu-cli-os-packages/rpm
enabled=1
gpgcheck=0
EOF
yum install -y tanzu-cli
```
