#!/bin/sh

# Copyright 2021 VMware, Inc. All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0

# This shell script creates the prescribed directory structure for a Tanzu management package.

set -eoux pipefail

# set this value to your package repository
PACKAGE_REPOSITORY=$1

# set this value to your package name
PACKAGE_NAME=$2

if [ -z "$PACKAGE_NAME" ]
then
  echo "create package failed. must set PACKAGE_REPOSITORY"
  exit 2
fi

if [ -z "$PACKAGE_NAME" ]
then
  echo "create package failed. must set PACKAGE_NAME"
  exit 2
fi

# Handle differences in MacOS sed
SEDARGS=""
if [ "$(uname -s)" = "Darwin" ]; then
    SEDARGS="-e"
fi

ROOT_DIR="packages"
BUNDLE_DIR="bundle"
CONFIG_DIR="config"
OVERLAY_DIR="overlay"
UPSTREAM_DIR="upstream"
IMGPKG_DIR=".imgpkg"
PACKAGE_DIR="${ROOT_DIR}/${PACKAGE_REPOSITORY}/${PACKAGE_NAME}"

# create directory structure for package
mkdir -vp "${PACKAGE_DIR}/${BUNDLE_DIR}/${CONFIG_DIR}"
mkdir -v "${PACKAGE_DIR}/${BUNDLE_DIR}/${IMGPKG_DIR}"
mkdir -v "${PACKAGE_DIR}/${BUNDLE_DIR}/${CONFIG_DIR}/${OVERLAY_DIR}"
mkdir -v "${PACKAGE_DIR}/${BUNDLE_DIR}/${CONFIG_DIR}/${UPSTREAM_DIR}"

# create README and fill with name of package
sed $SEDARGS "s/PACKAGE_NAME/${PACKAGE_NAME}/g" hack/packages/templates/new-package/readme.md > "${PACKAGE_DIR}/README.md"

# create manifests and fill with name of package
sed $SEDARGS "s/PACKAGE_NAME/${PACKAGE_NAME}/g" hack/packages/templates/new-package/metadata.yaml > "${PACKAGE_DIR}/metadata.yaml"
sed $SEDARGS "s/PACKAGE_NAME/${PACKAGE_NAME}/g" hack/packages/templates/new-package/package.yaml > "${PACKAGE_DIR}/package.yaml"
sed $SEDARGS "s/PACKAGE_NAME/${PACKAGE_NAME}/g" hack/packages/templates/new-package/values.yaml > "${PACKAGE_DIR}/bundle/config/values.yaml"
sed $SEDARGS "s/PACKAGE_NAME/${PACKAGE_NAME}/g" hack/packages/templates/new-package/vendir.yml > "${PACKAGE_DIR}/vendir.yml"
cp hack/packages/templates/new-package/Makefile ${PACKAGE_DIR}/Makefile

echo
echo "${PACKAGE_NAME} package bootstrapped at ${PACKAGE_DIR}"
echo
