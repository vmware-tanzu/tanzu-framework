#!/usr/bin/env bash

# Copyright 2022 VMware, Inc. All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0

set -e

if [ $(uname) != "Linux" ]; then
   echo "This script must be run on a Linux system"
   exit 1
fi

# Use DNF and if it is not installed fallback to YUM
DNF=$(command -v dnf || command -v yum || true)
if [ -z "$DNF" ]; then
   echo "This script requires the presence of either DNF or YUM package manager"
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
OUTPUT_DIR=${BASE_DIR}/_output/rpm

# Install build dependencies
$DNF install -y rpmdevtools rpmlint createrepo

rpmlint ${BASE_DIR}/tanzu-cli.spec

# We must create the sources directory ourselves in the below location
mkdir -p ${HOME}/rpmbuild/SOURCES

# Create the .rpm packages
rm -rf ${OUTPUT_DIR}
mkdir -p ${OUTPUT_DIR}
rpmbuild --define "cli_version ${VERSION}" -bb ${BASE_DIR}/tanzu-cli.spec --target amd64
mv ${HOME}/rpmbuild/RPMS/amd64/* ${OUTPUT_DIR}/
rpmbuild --define "cli_version ${VERSION}" -bb ${BASE_DIR}/tanzu-cli.spec --target aarch64
mv ${HOME}/rpmbuild/RPMS/aarch64/* ${OUTPUT_DIR}/

# Create the repository metadata
createrepo ${OUTPUT_DIR}
