#!/bin/bash

# Copyright 2021 VMware, Inc. All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0

# Inspired by - https://github.com/vmware-tanzu/community-edition/blob/main/hack/install.sh
# Script to install tanzu framework
# Usage: ./hack/install.sh /path/to/tanzu-framework/core/binary

set -o errexit
set -o nounset
set -o pipefail
set -o xtrace

#Identify the OS specific binary to be downloaded
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH="amd64"
BASE_CLI_PACKAGE_NAME="tanzu-cli"
COMPRESSION="tar.gz"
BASE_CLI="tanzu-core"

case "${OS}" in
  linux|darwin)
    BINARY="${BASE_CLI_PACKAGE_NAME}-${OS}-${ARCH}.${COMPRESSION}"
    ;;
  *)
    echo "${OS} is unsupported"
    exit 1
    ;;
esac

#Validate if required tools are installed
TOOLS=("curl" "jq")
MISSING_TOOLS=()
for tool in "${TOOLS[@]}";do
  if ! type "${tool}" > /dev/null; then
    MISSING_TOOLS+=("${tool}")
  fi
done
if [ ${#MISSING_TOOLS[@]} -gt 0 ]; then
    echo "Missing installation of required tool(s)- ${MISSING_TOOLS[*]}"
    exit 1
fi


GH_REPO_ORG="vmware-tanzu"
GH_REPO="tanzu-framework"
GH_API="https://api.github.com"
GH_REPO_API_BASE="${GH_API}/repos/${GH_REPO_ORG}/${GH_REPO}"
GH_TAGS="${GH_REPO_API_BASE}/releases/latest"
AUTH=""
# Validate GitHub access token
if [ -z "${GITHUB_TOKEN:-}" ]; then
    echo "Warning: No GITHUB_TOKEN variable defined - requests to the GitHub API may be rate limited"
else
    AUTH="Authorization: token ${GITHUB_TOKEN}"
    curl -o /dev/null -sLH "${AUTH}" ${GH_REPO_API_BASE} || { echo "Error: Unauthenticated token or network issue!";  exit 1; }
fi


echo "Fetching information about the latest Tanzu Framework release."
LATEST_RELEASE_RESPONSE=$(curl -sLH "${AUTH}" ${GH_TAGS} )
BIN_DOWNLOAD_URL=$(echo "${LATEST_RELEASE_RESPONSE}" | jq -r ".assets[].browser_download_url" | grep "${BINARY}")
LATEST_VERSION=$(echo "${LATEST_RELEASE_RESPONSE}" | jq -r ".tag_name")

echo "The latest version of Tanzu Framework CLI '${LATEST_VERSION}' will be downloaded and installed."

#Generate temp directory where cli will be downloaded
TF_INSTALL_PATH=$(mktemp -d 2>/dev/null || mktemp -d -t 'tf-install-tmp')

function cleanup {
  echo "Cleanup temporary binary download path."
  rm -rf "${TF_INSTALL_PATH}"
}
trap cleanup EXIT

echo "Downloading '${BINARY}' from the latest Tanzu Framework Release."
curl -sLH "${AUTH}" "${BIN_DOWNLOAD_URL}" > "${TF_INSTALL_PATH}/${BINARY}"
mkdir "${TF_INSTALL_PATH}/${LATEST_VERSION}"
tar -zxf "${TF_INSTALL_PATH}/${BINARY}" -C "${TF_INSTALL_PATH}" "${LATEST_VERSION}/${BASE_CLI}-${OS}_${ARCH}"

GH_RAW_CONTENT="https://raw.githubusercontent.com"
INSTALL_SCRIPT_URL="${GH_RAW_CONTENT}/${GH_REPO_ORG}/${GH_REPO}/${LATEST_VERSION}/hack/install.sh"
curl -sLH "${AUTH}" "${INSTALL_SCRIPT_URL}" | bash -se - "${TF_INSTALL_PATH}/${LATEST_VERSION}/${BASE_CLI}-${OS}_${ARCH}"
INSTALL_STATUS=$?
if [ "${INSTALL_STATUS}" -eq 0 ]; then
  echo "Successfully installed Tanzu CLI. You can verify the same by trying to check the version using 'tanzu version' command."
else
  echo "Failed to install the downloaded Tanzu CLI."
  exit 1
fi
