#!/usr/bin/env bash

set -e

if [ $(uname) != "Linux" ] || [ -z "$(command -v dnf)" ]; then
   echo "This script must be run on a Linux system that uses the DNF package manager"
   exit 1
fi

# VERSION should be set when calling this script
if [ -z "${VERSION}" ]; then
   echo "\$VERSION must be set before calling this script"
   exit 1
fi

# Strip 'v' prefix as an rpm package version must start with an integer
VERSION=${VERSION#v}

BASE_DIR=$(cd $(dirname "${BASH_SOURCE[0]}"); pwd)
OUTPUT_DIR=${BASE_DIR}/_output

# Install build dependencies
dnf install -y rpmdevtools rpmlint

rpmlint ${BASE_DIR}/tanzu-cli.spec

# We must reate the directory structure under $HOME
cd ${HOME}
rpmdev-setuptree
cd rpmbuild
# Copy the tanzu binary to the SOURCES directory
curl -o SOURCES/tanzu-cli-linux-amd64.tar.gz https://github.com/vmware-tanzu/tanzu-framework/releases/download/v${VERSION}/tanzu-cli-linux-amd64.tar.gz

# Copy the spec file
cp ${BASE_DIR}/tanzu-cli.spec SPECS

# Create the .rpm package
rpmbuild -bb ~/rpmbuild/SPECS/tanzu-cli.spec