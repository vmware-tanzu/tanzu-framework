#!/bin/bash

# Copyright 2021 VMware, Inc. All Rights Reserved.
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

#Tanzu-framework
temp_dir=$(mktemp -d)
TF_TAR_BALL="${temp_dir}/tanzu-framework-${OS}-${ARCH}.tar.gz"
gh release download "${version}" --repo "${TF_REPO_URL}" --pattern "tanzu-framework-${OS}-${ARCH}.tar.gz" --dir "${temp_dir}"
mkdir "${temp_dir}"/tanzu && tar xvzf ${TF_TAR_BALL} --directory "${temp_dir}" -C tanzu
if [[ ${OS} == 'darwin' ]]; then
	for binary in "${temp_dir}"/tanzu/cli/*; do
		if [[ -d "${binary}" ]]; then
			spctl -vv --type install --assess "${binary}"/"${version}"/*
		fi
	done
fi
./install.sh "${temp_dir}/tanzu/cli/core/${version}/tanzu-core-${OS}_${ARCH}" "${version}"
rm -rf $temp_dir

#Tanzu cli
temp_dir=$(mktemp -d)
TF_TAR_BALL="${temp_dir}/tanzu-cli-${OS}-${ARCH}.tar.gz"
gh release download "${version}" --repo ${TF_REPO_URL} --pattern "tanzu-cli-${OS}-${ARCH}.tar.gz" --dir "${temp_dir}"
mkdir "${temp_dir}"/tanzu && tar xvzf "${TF_TAR_BALL}" --directory "${temp_dir}" -C tanzu
if [ "${OS}" == 'darwin' ]; then
    spctl -vv --type install --asses "${temp_dir}/tanzu/${version}"/*
fi
./install.sh "${temp_dir}/tanzu/${version}/tanzu-core-${OS}_${ARCH}" "${version}"
rm -rf $temp_dir

#Context aware plugin
temp_dir=$(mktemp -d)
TF_TAR_BALL="${temp_dir}/tanzu-framework-plugins-context-${OS}-${ARCH}.tar.gz"
gh release download "${version}" --repo "${TF_REPO_URL}" --pattern "tanzu-framework-plugins-context-${OS}-${ARCH}.tar.gz" --dir "${temp_dir}"
mkdir "${temp_dir}"/tanzu && tar xvzf "${TF_TAR_BALL}" --directory "${temp_dir}" -C tanzu
if [[ "${OS}" == 'darwin' ]]; then
	for binary in "${temp_dir}"/tanzu/context-plugins/distribution/"${OS}"/"${ARCH}"/cli/*; do
		if [[ -d "${binary}" ]]; then
			spctl -vv --type install --assess "${binary}"/"${version}"/*
		fi
	done
fi
rm -rf $temp_dir

#Standalone plugin
temp_dir=$(mktemp -d)
TF_TAR_BALL="${temp_dir}/tanzu-framework-plugins-standalone-${OS}-${ARCH}.tar.gz"
gh release download "${version}" --repo "${TF_REPO_URL}" --pattern "tanzu-framework-plugins-standalone-${OS}-${ARCH}.tar.gz" --dir "${temp_dir}"
mkdir "${temp_dir}"/tanzu && tar xvzf "${TF_TAR_BALL}" --directory "${temp_dir}" -C tanzu
if [[ "${OS}" == 'darwin' ]]; then
	for binary in "${temp_dir}"/tanzu/standalone-plugins/distribution/"${OS}"/"${ARCH}"/cli/*; do
		if [[ -d "${binary}" ]]; then
			spctl -vv --type install --assess "${binary}"/"${version}"/*
		fi
	done
fi
rm -rf $temp_dir

#Admin Plugins
temp_dir=$(mktemp -d)
TF_TAR_BALL="${temp_dir}/tanzu-framework-plugins-admin-${OS}-${ARCH}.tar.gz"
gh release download "${version}" --repo "${TF_REPO_URL}" --pattern "tanzu-framework-plugins-admin-${OS}-${ARCH}.tar.gz" --dir "${temp_dir}"
mkdir "${temp_dir}"/tanzu && tar xvzf "${TF_TAR_BALL}" --directory "${temp_dir}" -C tanzu
if [[ "${OS}" == 'darwin' ]]; then
	for binary in "${temp_dir}"/tanzu/admin-plugins/distribution/"${OS}"/"${ARCH}"/cli/*; do
		if [[ -d "${binary}" ]]; then
			spctl -vv --type install --assess "${binary}"/"${version}"/*
		fi
	done
fi
rm -rf $temp_dir

tanzu version

tanzu management-cluster version

tanzu package version

tanzu secret version

tanzu login version

tanzu pinniped-auth version