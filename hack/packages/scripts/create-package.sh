#!/bin/sh

# Copyright 2021 VMware, Inc. All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0
q
# this shell script creates the prescribed directory structure for a Tanzu package.

# set this value to your package name
NAME=$1

if [ -z "$NAME" ]
then
  echo "create package failed. must set NAME"
  exit 2
fi

# Handle differences in MacOS sed
SEDARGS=""
if [ "$(uname -s)" = "Darwin" ]; then
    SEDARGS="-e"
fi

ROOT_DIR="management-packages"
BUNDLE_DIR="bundle"
CONFIG_DIR="config"
OVERLAY_DIR="overlay"
UPSTREAM_DIR="upstream"
IMGPKG_DIR=".imgpkg"
PACKAGE_DIR="${ROOT_DIR}/${NAME}"

# create directory structure for package
mkdir -vp "${PACKAGE_DIR}/${BUNDLE_DIR}/${CONFIG_DIR}"
mkdir -v "${PACKAGE_DIR}/${BUNDLE_DIR}/${IMGPKG_DIR}"
mkdir -v "${PACKAGE_DIR}/${BUNDLE_DIR}/${CONFIG_DIR}/${OVERLAY_DIR}"
mkdir -v "${PACKAGE_DIR}/${BUNDLE_DIR}/${CONFIG_DIR}/${UPSTREAM_DIR}"

# create README and fill with name of package
sed $SEDARGS "s/PACKAGE_NAME/${NAME}/g" hack/packages/templates/new-package/readme.md > "${PACKAGE_DIR}/README.md"

# create manifests and fill with name of package
sed $SEDARGS "s/PACKAGE_NAME/${NAME}/g" hack/packages/templates/new-package/metadata.yaml > "${PACKAGE_DIR}/metadata.yaml"
sed $SEDARGS "s/PACKAGE_NAME/${NAME}/g" hack/packages/templates/new-package/package.yaml > "${PACKAGE_DIR}/package_a.yaml"
sed $SEDARGS "s/VERSION/${VERSION}/g" "${PACKAGE_DIR}/package_a.yaml" > "${PACKAGE_DIR}/package.yaml"
sed $SEDARGS "s/PACKAGE_NAME/${NAME}/g" hack/packages/templates/new-package/values.yaml > "${PACKAGE_DIR}/bundle/config/values.yaml"
sed $SEDARGS "s/PACKAGE_NAME/${NAME}/g" hack/packages/templates/new-package/vendir.yml > "${PACKAGE_DIR}/bundle/vendir.yml"

rm "${PACKAGE_DIR}/package_a.yaml"

echo
echo "package bootstrapped at ${PACKAGE_DIR}"
echo
