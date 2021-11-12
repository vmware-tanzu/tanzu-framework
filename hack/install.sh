#!/bin/bash

# Copyright 2021 VMware Tanzu Community Edition contributors. All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0

# Inspired by - https://github.com/vmware-tanzu/community-edition/blob/main/hack/install.sh
# Script to install tanzu framework
# Usage: ./hack/install.sh /path/to/tanzu-framework-binary v0.10.0

# set -o errexit
set -o nounset
set -o pipefail
set -o xtrace

TF_INSTALL_PATH="${1:?Tanzu-Framework path argument empty. Example usage: ./hack/install.sh /path/to/tanzu-framework-binary v0.10.0}"
VERSION="${2:?Version argument empty. Example usage: ./hack/install.sh /path/to/tanzu-framework-binary v0.10.0}"

ALLOW_INSTALL_AS_ROOT="${ALLOW_INSTALL_AS_ROOT:-""}"
if [[ "$EUID" -eq 0 && "${ALLOW_INSTALL_AS_ROOT}" != "true" ]]; then
  echo "Do not run this script as root"
  exit 1
fi

MY_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"

OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH="amd64"

case "${OS}" in
  linux)
    XDG_DATA_HOME="${HOME}/.local/share"
    ;;
  darwin)
    XDG_DATA_HOME="${HOME}/Library/Application Support"
    ;;
  *)
    echo "${OS} is unsupported"
    exit 1
    ;;
esac
echo "${XDG_DATA_HOME}"

# check if the tanzu CLI already exists and remove it to avoid conflicts
TANZU_BIN_PATH=$(command -v tanzu)
if [[ -n "${TANZU_BIN_PATH}" ]]; then
  # best effort, so just ignore errors
  sudo rm -f "${TANZU_BIN_PATH}" > /dev/null
fi

# set install dir to /usr/local/bin
TANZU_BIN_PATH="/usr/local/bin"
echo Installing tanzu cli to "${TANZU_BIN_PATH}"

# if plugin cache pre-exists, remove it so new plugins are detected
TANZU_PLUGIN_CACHE="${HOME}/.cache/tanzu/catalog.yaml"
if [[ -n "${TANZU_PLUGIN_CACHE}" ]]; then
  echo "Removing old plugin cache from ${TANZU_PLUGIN_CACHE}"
  rm -f "${TANZU_PLUGIN_CACHE}" > /dev/null
fi

# install tanzu cli
sudo install "${TF_INSTALL_PATH}/cli/core/${VERSION}/tanzu-core-${OS}_${ARCH}" "${TANZU_BIN_PATH}/tanzu"

# install plugins
tanzu plugin sync

tanzu plugin list

echo "Installation complete!"
