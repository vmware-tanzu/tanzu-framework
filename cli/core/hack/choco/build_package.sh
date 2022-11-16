#!/usr/bin/env bash

# Copyright 2022 VMware, Inc. All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0

set -e

if [ "$(command -v choco)" = "" ]; then
   echo "This script must be run on a system that has 'choco' installed"
   exit 1
fi

# VERSION should be set when calling this script
if [ -z "${VERSION}" ]; then
   echo "\$VERSION must be set before calling this script"
   exit 1
fi

# Strip 'v' prefix to be consistent with our other package names
VERSION=${VERSION#v}

BASE_DIR=$(cd $(dirname "${BASH_SOURCE[0]}"); pwd)
OUTPUT_DIR=${BASE_DIR}/_output/choco

mkdir -p ${OUTPUT_DIR}/
# Remove the nupkg file made by `choco pack` in the working dir
rm -f ${OUTPUT_DIR}/*.nupkg

# Obtain SHA
SHA=$(curl -sL https://github.com/vmware-tanzu/tanzu-framework/releases/download/v${VERSION}/tanzu-framework-executables-checksums.txt | grep tanzu-cli-windows-amd64.zip |cut -f1 -d" ")
if [ -z "$SHA" ]; then
   echo "Unable to determine SHA for package of version $VERSION"
   exit 1
fi

# Prepare install script
sed -e s,__CLI_VERSION__,v${VERSION}, -e s,__CLI_SHA__,${SHA}, \
   ${BASE_DIR}/chocolateyInstall.ps1.tmpl > ${OUTPUT_DIR}/chocolateyInstall.ps1
chmod a+x ${OUTPUT_DIR}/chocolateyInstall.ps1

# Bundle the powershell scripts and nuspec into a nupkg file
choco pack ${BASE_DIR}/tanzu-cli-release.nuspec --out ${OUTPUT_DIR} "cliVersion=${VERSION}"

# Upload the nupkg file to the registry
# DON'T DO THIS YET
# choco push --source https://push.chocolatey.org/ --api-key .......
