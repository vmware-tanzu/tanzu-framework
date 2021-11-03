#!/usr/bin/env bash

# Copyright 2021 VMware, Inc. All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0

set -eoux pipefail

PROJECT_ROOT=$(git rev-parse --show-toplevel)
TOOLS_DIR="${PROJECT_ROOT}/hack/tools"
TOOLS_BIN_DIR="${TOOLS_DIR}/bin"
MANAGEMENT_PACKAGES_BUILD_ARTIFACTS_DIR="${PROJECT_ROOT}/build"

if [[ ${MANAGEMENT_PACKAGE_REPO_VERSION:0:1} == "v" ]] ; then
  REPO_VERSION=${MANAGEMENT_PACKAGE_REPO_VERSION:1}
else
  REPO_VERSION=${MANAGEMENT_PACKAGE_REPO_VERSION}
fi

function generate_imgpkg_lock_output() {
  while IFS='|' read -r name path version; do
    mkdir -p "$path/bundle/.imgpkg"
    yttCmd="${TOOLS_BIN_DIR}/ytt --ignore-unknown-comments -f $path/bundle/config/"
    ${yttCmd} | "${TOOLS_BIN_DIR}"/kbld -f - -f "${PROJECT_ROOT}/management-packages/kbld-config.yml" --imgpkg-lock-output "$path/bundle/.imgpkg/images.yml" > /dev/null
  done < <("${TOOLS_BIN_DIR}"/yq e ".${PACKAGE_REPOSITORY}PackageRepository.packages[] | .name + \"|\" + .path + \"|\" + .version" "${PROJECT_ROOT}/management-packages/package-values.yaml")
}

function create_package_bundles() {
  cp ${PROJECT_ROOT}/management-packages/package-values.yaml "${PROJECT_ROOT}/management-packages/package-values-sha256.yaml"
  while IFS='|' read -r name version packageSubVersion path; do
    if [ -z "$packageSubVersion" ]; then
      packageVersion="${REPO_VERSION}"
      imagePackageVersion="v${REPO_VERSION}"
    else
      packageVersion="$REPO_VERSION+$packageSubVersion"
      imagePackageVersion="v${REPO_VERSION}_${packageSubVersion}"
    fi
    "${TOOLS_BIN_DIR}"/imgpkg push -b "${1}/$name:$imagePackageVersion" --file "$path/bundle" --lock-output "$name-$packageVersion-lock-output.yaml"
    "${TOOLS_BIN_DIR}"/yq e '.bundle | .image' "$name-$packageVersion-lock-output.yaml" | sed "s,${1}/$name@sha256:, ,g" | xargs -I '{}' sed -ie "s,${name}:${version},{},g" "${PROJECT_ROOT}/management-packages/package-values-sha256.yaml"
    [ -z "$packageSubVersion" ] && echo "${REPO_VERSION}" | sed "s,${1}/$name@version:, ,g" | xargs -I '{}' sed -ie "s,${version},{},g" "${PROJECT_ROOT}/management-packages/package-values-sha256.yaml"
    mkdir -p "${MANAGEMENT_PACKAGES_BUILD_ARTIFACTS_DIR}/package-bundles/${PACKAGE_REPOSITORY}"
    tar -czvf "${MANAGEMENT_PACKAGES_BUILD_ARTIFACTS_DIR}/package-bundles/${PACKAGE_REPOSITORY}/$name-$imagePackageVersion.tar.gz" -C "$path/bundle" .
    rm -f "$name-$packageVersion-lock-output.yaml" "${PROJECT_ROOT}/management-packages/package-values-sha256.yamle"
  done < <("${TOOLS_BIN_DIR}"/yq e ".${PACKAGE_REPOSITORY}PackageRepository.packages[] | .name + \"|\" + .version + \"|\" + .packageSubVersion + \"|\" + .path" "${PROJECT_ROOT}/management-packages/package-values.yaml")
}

function create_package_repo_bundles() {
  mkdir -p "${MANAGEMENT_PACKAGES_BUILD_ARTIFACTS_DIR}/package-repo-bundles/${PACKAGE_REPOSITORY}/.imgpkg" "${MANAGEMENT_PACKAGES_BUILD_ARTIFACTS_DIR}/package-repo-bundles/${PACKAGE_REPOSITORY}/packages"

  "${TOOLS_BIN_DIR}"/ytt -f hack/packages/templates/repo-utils/images-tmpl.yaml -f hack/packages/templates/repo-utils/package-helpers.lib.yaml -f ${PROJECT_ROOT}/management-packages/package-values-sha256.yaml -v packageRepository="${PACKAGE_REPOSITORY}" -v "${PACKAGE_REPOSITORY}PackageRepository.registry=${REGISTRY}" > "${MANAGEMENT_PACKAGES_BUILD_ARTIFACTS_DIR}/package-repo-bundles/${PACKAGE_REPOSITORY}/.imgpkg/images.yml"

  domain=$("${TOOLS_BIN_DIR}"/yq e ".${PACKAGE_REPOSITORY}PackageRepository.domain" "${PROJECT_ROOT}/management-packages/package-values.yaml")

  timestamp=$(date +"%Y-%m-%dT%H:%M:%SZ")

  while IFS='|' read -r name version packageSubVersion path; do
    mkdir -p "${MANAGEMENT_PACKAGES_BUILD_ARTIFACTS_DIR}/package-repo-bundles/${PACKAGE_REPOSITORY}/packages/${name}.${domain}"

    packageYttCmd="${TOOLS_BIN_DIR}/ytt -f $path/package.yaml -f hack/packages/templates/repo-utils/package-cr-overlay.yaml -f hack/packages/templates/repo-utils/package-helpers.lib.yaml -f ${PROJECT_ROOT}/management-packages/package-values-sha256.yaml -v packageRepository=${PACKAGE_REPOSITORY} -v packageName=${name} -v ${PACKAGE_REPOSITORY}PackageRepository.registry=${REGISTRY} -v timestamp=${timestamp}"
    if [ -z "$packageSubVersion" ]; then
      packageFileName="${REPO_VERSION}.yml"
    else
      packageFileName="${REPO_VERSION}+${packageSubVersion}.yml"
    fi
    ${packageYttCmd} > "${MANAGEMENT_PACKAGES_BUILD_ARTIFACTS_DIR}/package-repo-bundles/${PACKAGE_REPOSITORY}/packages/${name}.${domain}/${packageFileName}"

    packageMetadataYttCmd="${TOOLS_BIN_DIR}/ytt -f $path/metadata.yaml -f hack/packages/templates/repo-utils/package-metadata-cr-overlay.yaml -f hack/packages/templates/repo-utils/package-helpers.lib.yaml -f ${PROJECT_ROOT}/management-packages/package-values-sha256.yaml -v packageRepository=${PACKAGE_REPOSITORY} -v packageName=${name} -v ${PACKAGE_REPOSITORY}PackageRepository.registry=${REGISTRY}"
    ${packageMetadataYttCmd} > "${MANAGEMENT_PACKAGES_BUILD_ARTIFACTS_DIR}/package-repo-bundles/${PACKAGE_REPOSITORY}/packages/${name}.${domain}/metadata.yml"

  done < <("${TOOLS_BIN_DIR}"/yq e ".${PACKAGE_REPOSITORY}PackageRepository.packages[] | .name + \"|\" + .version + \"|\" + .packageSubVersion + \"|\" + .path" "${PROJECT_ROOT}/management-packages/package-values.yaml")

  pushd "${MANAGEMENT_PACKAGES_BUILD_ARTIFACTS_DIR}/package-repo-bundles"
  tar -czvf "tanzu-framework-${PACKAGE_REPOSITORY}-repo-${REPO_VERSION}.tar.gz" -C "${PACKAGE_REPOSITORY}" .
  mv "tanzu-framework-${PACKAGE_REPOSITORY}-repo-${REPO_VERSION}.tar.gz" "${PACKAGE_REPOSITORY}"/
  popd
}

function trivy_scan() {
  tmp_dir=$(mktemp -d)
  while IFS='|' read -r image; do
    "${TOOLS_BIN_DIR}"/trivy --cache-dir "$tmp_dir" image --exit-code 1 --severity CRITICAL --ignore-unfixed "$image"
  done < <("${TOOLS_BIN_DIR}"/yq e ".overrides[] | .newImage" "${PROJECT_ROOT}/management-packages/kbld-config.yml")
}

function push_package_bundles() {
  while IFS='|' read -r name version packageSubVersion; do
    if [ "$packageSubVersion" == null -o "$packageSubVersion" == "" ]; then
      packageVersion="${REPO_VERSION}"
      imagePackageVersion="v${REPO_VERSION}"
    else
      packageVersion="$REPO_VERSION+$packageSubVersion"
      imagePackageVersion="v${REPO_VERSION}_${packageSubVersion}"
    fi
    mkdir -p "${MANAGEMENT_PACKAGES_BUILD_ARTIFACTS_DIR}/package-bundles/${PACKAGE_REPOSITORY}/${name}-${packageVersion}"
    tar -xvf "${MANAGEMENT_PACKAGES_BUILD_ARTIFACTS_DIR}/package-bundles/${PACKAGE_REPOSITORY}/${name}-${imagePackageVersion}.tar.gz" -C "${MANAGEMENT_PACKAGES_BUILD_ARTIFACTS_DIR}/package-bundles/${PACKAGE_REPOSITORY}/${name}-${packageVersion}"
    "${TOOLS_BIN_DIR}"/imgpkg push -b "${REGISTRY}/${name}:${imagePackageVersion}" --file "${MANAGEMENT_PACKAGES_BUILD_ARTIFACTS_DIR}/package-bundles/${PACKAGE_REPOSITORY}/${name}-${packageVersion}"
    rm -rf "${MANAGEMENT_PACKAGES_BUILD_ARTIFACTS_DIR}/package-bundles/${PACKAGE_REPOSITORY}/${name}-${packageVersion}"
  done < <("${TOOLS_BIN_DIR}"/yq e ".${PACKAGE_REPOSITORY}PackageRepository.packages[] | .name + \"|\" + .version + \"|\" + .packageSubVersion" "${PROJECT_ROOT}/management-packages/package-values.yaml")
}

function push_package_repo_bundles() {
  mkdir -p "${MANAGEMENT_PACKAGES_BUILD_ARTIFACTS_DIR}/package-repo-bundles/${PACKAGE_REPOSITORY}/tanzu-framework-${PACKAGE_REPOSITORY}-repo-${REPO_VERSION}"
  tar -xvf "${MANAGEMENT_PACKAGES_BUILD_ARTIFACTS_DIR}/package-repo-bundles/${PACKAGE_REPOSITORY}/tanzu-framework-${PACKAGE_REPOSITORY}-repo-${REPO_VERSION}.tar.gz" -C "${MANAGEMENT_PACKAGES_BUILD_ARTIFACTS_DIR}/package-repo-bundles/${PACKAGE_REPOSITORY}/tanzu-framework-${PACKAGE_REPOSITORY}-repo-${REPO_VERSION}"
  "${TOOLS_BIN_DIR}"/imgpkg push -b "${REGISTRY}/${PACKAGE_REPOSITORY}:${REPO_VERSION}" --file "${MANAGEMENT_PACKAGES_BUILD_ARTIFACTS_DIR}/package-repo-bundles/${PACKAGE_REPOSITORY}/tanzu-framework-${PACKAGE_REPOSITORY}-repo-${REPO_VERSION}" --lock-output "${MANAGEMENT_PACKAGES_BUILD_ARTIFACTS_DIR}/package-repo-bundles/${PACKAGE_REPOSITORY}/tanzu-framework-${PACKAGE_REPOSITORY}-repo-${REPO_VERSION}/tanzu-framework-${PACKAGE_REPOSITORY}-repo-${REPO_VERSION}-lock-output.yaml"

  REPO_URL=$("${TOOLS_BIN_DIR}"/yq e ".bundle.image" "${MANAGEMENT_PACKAGES_BUILD_ARTIFACTS_DIR}/package-repo-bundles/${PACKAGE_REPOSITORY}/tanzu-framework-${PACKAGE_REPOSITORY}-repo-${REPO_VERSION}/tanzu-framework-${PACKAGE_REPOSITORY}-repo-${REPO_VERSION}-lock-output.yaml")
  REPO_VERSION_SHA=$(echo "${REPO_URL}" | cut -d ':' -f '2')
  "${TOOLS_BIN_DIR}"/ytt -f hack/packages/templates/repo-utils/packagerepo-tmpl.yaml -f hack/packages/templates/repo-utils/package-helpers.lib.yaml -f "${PROJECT_ROOT}/management-packages/package-values-sha256.yaml" -v packageRepository="${PACKAGE_REPOSITORY}" -v "${PACKAGE_REPOSITORY}PackageRepository.registry=${REGISTRY}" -v "${PACKAGE_REPOSITORY}PackageRepository.sha256=${REPO_VERSION_SHA}" > "${MANAGEMENT_PACKAGES_BUILD_ARTIFACTS_DIR}/package-repo-bundles/${PACKAGE_REPOSITORY}/tanzu-framework-${PACKAGE_REPOSITORY}-repo-${REPO_VERSION}.yaml"
  rm -rf "${MANAGEMENT_PACKAGES_BUILD_ARTIFACTS_DIR}/package-repo-bundles/${PACKAGE_REPOSITORY}/tanzu-framework-${PACKAGE_REPOSITORY}-repo-${REPO_VERSION}"
}

"$@"
