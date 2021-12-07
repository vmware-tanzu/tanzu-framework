#!/bin/bash

# Copyright 2021 VMware Tanzu Community Edition contributors. All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0

# Script to test release artifacts on MacOS and Linux OS
# Inspired by - https://github.com/vmware-tanzu/community-edition/blob/main/test/release-build-test/check-release-build.sh

set -o errexit
set -o nounset
set -o pipefail
set -o xtrace

version="${1:?Tanzu-Framework version argument empty. Example usage: ./hack/test-release-artifacts.sh v0.10.0}"
: "${GITHUB_TOKEN:?GITHUB_TOKEN is not set}"

TF_REPO_URL="https://github.com/vmware-tanzu/tanzu-framework"
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH="amd64"

temp_dir=$(mktemp -d)

TF_TAR_BALL="${temp_dir}/tanzu-framework-${OS}-${ARCH}.tar.gz"
TF_INSTALLATION_DIR="${temp_dir}/tanzu-framework-${OS}-${ARCH}"

gh release download "${version}" --repo ${TF_REPO_URL} --pattern "tanzu-framework-${OS}-${ARCH}.tar.gz" --dir "${temp_dir}"

mkdir "${temp_dir}"/tanzu && tar xvzf "${TF_TAR_BALL}" --directory "${temp_dir}" -C tanzu

if [ "${OS}" == 'darwin' ]; then
    for binary in "${temp_dir}"/tanzu/cli/*; do
        if [[ -d "${binary}" ]]; then
            spctl -vv --type install --asses "${binary}"/"${version}"/*
        fi
    done
fi

./hack/install.sh "${temp_dir}/tanzu/cli/core/${version}/tanzu-core-${OS}_${ARCH}" "${version}"

tanzu version

tanzu management-cluster version

tanzu package version

tanzu secret version

tanzu login version

tanzu pinniped-auth version
