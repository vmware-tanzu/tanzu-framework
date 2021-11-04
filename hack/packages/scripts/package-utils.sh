#!/usr/bin/env bash

# Copyright 2021 VMware, Inc. All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0

set -eoux pipefail

PROJECT_ROOT=$(git rev-parse --show-toplevel)
TOOLS_DIR="${PROJECT_ROOT}/hack/tools"
TOOLS_BIN_DIR="${TOOLS_DIR}/bin"
PACKAGES_BUILD_ARTIFACTS_DIR="${PROJECT_ROOT}/build"

if [[ ${REPO_VERSION:0:1} == "v" ]] ; then
  REPO_VERSION=${REPO_VERSION:1}
fi

function generate_single_imgpkg_lock_output() {
	path="${PROJECT_ROOT}/${PACKAGE_REPOSITORY}-packages/${PACKAGE_NAME}"
	mkdir -p "$path/bundle/.imgpkg"
	yttCmd="${TOOLS_BIN_DIR}/ytt --ignore-unknown-comments -f $path/bundle/config/"
	${yttCmd} | "${TOOLS_BIN_DIR}"/kbld -f - -f "${PROJECT_ROOT}/${PACKAGE_REPOSITORY}-packages/kbld-config.yml" --imgpkg-lock-output "$path/bundle/.imgpkg/images.yml" > /dev/null
}

function generate_imgpkg_lock_output() {
  while IFS='|' read -r name path version; do
    mkdir -p "$path/bundle/.imgpkg"
    yttCmd="${TOOLS_BIN_DIR}/ytt --ignore-unknown-comments -f $path/bundle/config/"
    ${yttCmd} | "${TOOLS_BIN_DIR}"/kbld -f - -f "${PROJECT_ROOT}/${PACKAGE_REPOSITORY}-packages/kbld-config.yml" --imgpkg-lock-output "$path/bundle/.imgpkg/images.yml" > /dev/null
  done < <("${TOOLS_BIN_DIR}"/yq e ".${PACKAGE_REPOSITORY}PackageRepository.packages[] | .name + \"|\" + .path + \"|\" + .version" "${PROJECT_ROOT}/${PACKAGE_REPOSITORY}-packages/package-values.yaml")
}

function create_single_package_bundle() {
	path="${PROJECT_ROOT}/${PACKAGE_REPOSITORY}-packages/${PACKAGE_NAME}"
	if [ -z "$PACKAGE_SUB_VERSION" ]; then
      imagePackageVersion="v${REPO_VERSION}"
  else
      imagePackageVersion="v${REPO_VERSION}_${PACKAGE_SUB_VERSION}"
  fi
	mkdir -p "build/package-bundles/${PACKAGE_REPOSITORY}"
	tar -czvf "build/package-bundles/${PACKAGE_REPOSITORY}/${PACKAGE_NAME}-$imagePackageVersion.tar.gz" -C "$path/bundle" .
}

function create_package_bundles() {
  cp "${PROJECT_ROOT}/${PACKAGE_REPOSITORY}-packages/package-values.yaml" "${PROJECT_ROOT}/${PACKAGE_REPOSITORY}-packages/package-values-sha256.yaml"
  while IFS='|' read -r name version packageSubVersion path; do
    if [ "$packageSubVersion" == null -o "$packageSubVersion" == "" ]; then
      packageVersion="${REPO_VERSION}"
      imagePackageVersion="v${REPO_VERSION}"
    else
      packageVersion="$REPO_VERSION+$packageSubVersion"
      imagePackageVersion="v${REPO_VERSION}_${packageSubVersion}"
    fi
    "${TOOLS_BIN_DIR}"/imgpkg push -b "${1}/$name:$imagePackageVersion" --file "$path/bundle" --lock-output "$name-$packageVersion-lock-output.yaml"
    "${TOOLS_BIN_DIR}"/yq e '.bundle | .image' "$name-$packageVersion-lock-output.yaml" | sed "s,${1}/$name@sha256:, ,g" | xargs -I '{}' sed -ie "s,${name}:${version},{},g" "${PROJECT_ROOT}/${PACKAGE_REPOSITORY}-packages/package-values-sha256.yaml"
    [ -z "$packageSubVersion" ] && echo "${REPO_VERSION}" | sed "s,${1}/$name@version:, ,g" | xargs -I '{}' sed -ie "s,${version},{},g" "${PROJECT_ROOT}/${PACKAGE_REPOSITORY}-packages/package-values-sha256.yaml"
    mkdir -p "${PACKAGES_BUILD_ARTIFACTS_DIR}/package-bundles/${PACKAGE_REPOSITORY}"
    tar -czvf "${PACKAGES_BUILD_ARTIFACTS_DIR}/package-bundles/${PACKAGE_REPOSITORY}/$name-$imagePackageVersion.tar.gz" -C "$path/bundle" .
    rm -f "$name-$packageVersion-lock-output.yaml" "${PROJECT_ROOT}/${PACKAGE_REPOSITORY}-packages/package-values-sha256.yamle"
  done < <("${TOOLS_BIN_DIR}"/yq e ".${PACKAGE_REPOSITORY}PackageRepository.packages[] | .name + \"|\" + .version + \"|\" + .packageSubVersion + \"|\" + .path" "${PROJECT_ROOT}/${PACKAGE_REPOSITORY}-packages/package-values.yaml")
}

function generate_package_bundles_sha256() {
  cp "${PROJECT_ROOT}/${PACKAGE_REPOSITORY}-packages/package-values.yaml" "${PROJECT_ROOT}/${PACKAGE_REPOSITORY}-packages/package-values-sha256.yaml"
	while IFS='|' read -r name version packageSubVersion; do
		path="${PROJECT_ROOT}/${PACKAGE_REPOSITORY}-packages/${name}"
		if [ "$packageSubVersion" == null -o "$packageSubVersion" == "" ]; then
      packageVersion="${REPO_VERSION}"
      imagePackageVersion="v${REPO_VERSION}"
    else
      packageVersion="$REPO_VERSION+$packageSubVersion"
      imagePackageVersion="v${REPO_VERSION}_${packageSubVersion}"
    fi
		"${TOOLS_BIN_DIR}"/imgpkg push -b "${1}/$name:$imagePackageVersion" --file "$path/bundle" --lock-output "$name-$packageVersion-lock-output.yaml"
    "${TOOLS_BIN_DIR}"/yq e '.bundle | .image' "$name-$packageVersion-lock-output.yaml" | sed "s,${1}/$name@sha256:, ,g" | xargs -I '{}' sed -ie "s,${name}:${version},{},g" "${PROJECT_ROOT}/${PACKAGE_REPOSITORY}-packages/package-values-sha256.yaml"
    [ "$packageSubVersion" == null -o "$packageSubVersion" == "" ] && echo "${REPO_VERSION}" | sed "s,${1}/$name@version:, ,g" | xargs -I '{}' sed -ie "s,${version},{},g" "${PROJECT_ROOT}/${PACKAGE_REPOSITORY}-packages/package-values-sha256.yaml"
    rm -f "$name-$packageVersion-lock-output.yaml" "${PROJECT_ROOT}/${PACKAGE_REPOSITORY}-packages/package-values-sha256.yamle"
  done < <("${TOOLS_BIN_DIR}"/yq e ".${PACKAGE_REPOSITORY}PackageRepository.packages[] | .name + \"|\" + .version + \"|\" + .packageSubVersion" "${PROJECT_ROOT}/${PACKAGE_REPOSITORY}-packages/package-values.yaml")
}

function create_package_repo_bundles() {
  if [ -z "$PACKAGE_VALUES_FILE" ];
	then
	  generate_package_bundles_sha256 localhost:5000
		PACKAGE_VALUES_FILE="${PROJECT_ROOT}/${PACKAGE_REPOSITORY}-packages/package-values-sha256.yaml"
	fi
  mkdir -p "${PACKAGES_BUILD_ARTIFACTS_DIR}/package-repo-bundles/${PACKAGE_REPOSITORY}/.imgpkg" "${PACKAGES_BUILD_ARTIFACTS_DIR}/package-repo-bundles/${PACKAGE_REPOSITORY}/packages"

  "${TOOLS_BIN_DIR}"/ytt -f hack/packages/templates/repo-utils/images-tmpl.yaml -f hack/packages/templates/repo-utils/package-helpers.lib.yaml -f "${PACKAGE_VALUES_FILE}" -v packageRepository="${PACKAGE_REPOSITORY}" -v "${PACKAGE_REPOSITORY}PackageRepository.registry=${REGISTRY}" > "${PACKAGES_BUILD_ARTIFACTS_DIR}/package-repo-bundles/${PACKAGE_REPOSITORY}/.imgpkg/images.yml"

  domain=$("${TOOLS_BIN_DIR}"/yq e ".${PACKAGE_REPOSITORY}PackageRepository.domain" "${PACKAGE_VALUES_FILE}")

  timestamp=$(date +"%Y-%m-%dT%H:%M:%SZ")

  while IFS='|' read -r name version packageSubVersion path; do
    mkdir -p "${PACKAGES_BUILD_ARTIFACTS_DIR}/package-repo-bundles/${PACKAGE_REPOSITORY}/packages/${name}.${domain}"

    packageYttCmd="${TOOLS_BIN_DIR}/ytt -f $path/package.yaml -f hack/packages/templates/repo-utils/package-cr-overlay.yaml -f hack/packages/templates/repo-utils/package-helpers.lib.yaml -f ${PACKAGE_VALUES_FILE} -v packageRepository=${PACKAGE_REPOSITORY} -v packageName=${name} -v ${PACKAGE_REPOSITORY}PackageRepository.registry=${REGISTRY} -v timestamp=${timestamp}"
    if [ "$packageSubVersion" == null -o "$packageSubVersion" == "" ]; then
      packageFileName="${REPO_VERSION}.yml"
    else
      packageFileName="${REPO_VERSION}+${packageSubVersion}.yml"
    fi
    ${packageYttCmd} > "${PACKAGES_BUILD_ARTIFACTS_DIR}/package-repo-bundles/${PACKAGE_REPOSITORY}/packages/${name}.${domain}/${packageFileName}"

    packageMetadataYttCmd="${TOOLS_BIN_DIR}/ytt -f $path/metadata.yaml -f hack/packages/templates/repo-utils/package-metadata-cr-overlay.yaml -f hack/packages/templates/repo-utils/package-helpers.lib.yaml -f ${PACKAGE_VALUES_FILE} -v packageRepository=${PACKAGE_REPOSITORY} -v packageName=${name} -v ${PACKAGE_REPOSITORY}PackageRepository.registry=${REGISTRY}"
    ${packageMetadataYttCmd} > "${PACKAGES_BUILD_ARTIFACTS_DIR}/package-repo-bundles/${PACKAGE_REPOSITORY}/packages/${name}.${domain}/metadata.yml"

  done < <("${TOOLS_BIN_DIR}"/yq e ".${PACKAGE_REPOSITORY}PackageRepository.packages[] | .name + \"|\" + .version + \"|\" + .packageSubVersion + \"|\" + .path" "${PACKAGE_VALUES_FILE}")

  pushd "${PACKAGES_BUILD_ARTIFACTS_DIR}/package-repo-bundles"
  tar -czvf "tanzu-framework-${PACKAGE_REPOSITORY}-repo-${REPO_VERSION}.tar.gz" -C "${PACKAGE_REPOSITORY}" .
  mv "tanzu-framework-${PACKAGE_REPOSITORY}-repo-${REPO_VERSION}.tar.gz" "${PACKAGE_REPOSITORY}"/
  popd
}

function trivy_scan() {
  tmp_dir=$(mktemp -d)
  while IFS='|' read -r image; do
    "${TOOLS_BIN_DIR}"/trivy --cache-dir "$tmp_dir" image --exit-code 1 --severity CRITICAL --ignore-unfixed "$image"
  done < <("${TOOLS_BIN_DIR}"/yq e ".overrides[] | .newImage" "${PROJECT_ROOT}/${PACKAGE_REPOSITORY}-packages/kbld-config.yml")
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
    mkdir -p "${PACKAGES_BUILD_ARTIFACTS_DIR}/package-bundles/${PACKAGE_REPOSITORY}/${name}-${packageVersion}"
    tar -xvf "${PACKAGES_BUILD_ARTIFACTS_DIR}/package-bundles/${PACKAGE_REPOSITORY}/${name}-${imagePackageVersion}.tar.gz" -C "${PACKAGES_BUILD_ARTIFACTS_DIR}/package-bundles/${PACKAGE_REPOSITORY}/${name}-${packageVersion}"
    "${TOOLS_BIN_DIR}"/imgpkg push -b "${REGISTRY}/${name}:${imagePackageVersion}" --file "${PACKAGES_BUILD_ARTIFACTS_DIR}/package-bundles/${PACKAGE_REPOSITORY}/${name}-${packageVersion}"
    rm -rf "${PACKAGES_BUILD_ARTIFACTS_DIR}/package-bundles/${PACKAGE_REPOSITORY}/${name}-${packageVersion}"
  done < <("${TOOLS_BIN_DIR}"/yq e ".${PACKAGE_REPOSITORY}PackageRepository.packages[] | .name + \"|\" + .version + \"|\" + .packageSubVersion" "${PROJECT_ROOT}/${PACKAGE_REPOSITORY}-packages/package-values.yaml")
}

function push_package_repo_bundles() {
  mkdir -p "${PACKAGES_BUILD_ARTIFACTS_DIR}/package-repo-bundles/${PACKAGE_REPOSITORY}/tanzu-framework-${PACKAGE_REPOSITORY}-repo-${REPO_VERSION}"
  tar -xvf "${PACKAGES_BUILD_ARTIFACTS_DIR}/package-repo-bundles/${PACKAGE_REPOSITORY}/tanzu-framework-${PACKAGE_REPOSITORY}-repo-${REPO_VERSION}.tar.gz" -C "${PACKAGES_BUILD_ARTIFACTS_DIR}/package-repo-bundles/${PACKAGE_REPOSITORY}/tanzu-framework-${PACKAGE_REPOSITORY}-repo-${REPO_VERSION}"
  "${TOOLS_BIN_DIR}"/imgpkg push -b "${REGISTRY}/${PACKAGE_REPOSITORY}:v${REPO_VERSION}" --file "${PACKAGES_BUILD_ARTIFACTS_DIR}/package-repo-bundles/${PACKAGE_REPOSITORY}/tanzu-framework-${PACKAGE_REPOSITORY}-repo-${REPO_VERSION}" --lock-output "${PACKAGES_BUILD_ARTIFACTS_DIR}/package-repo-bundles/${PACKAGE_REPOSITORY}/tanzu-framework-${PACKAGE_REPOSITORY}-repo-${REPO_VERSION}/tanzu-framework-${PACKAGE_REPOSITORY}-repo-${REPO_VERSION}-lock-output.yaml"

  REPO_URL=$("${TOOLS_BIN_DIR}"/yq e ".bundle.image" "${PACKAGES_BUILD_ARTIFACTS_DIR}/package-repo-bundles/${PACKAGE_REPOSITORY}/tanzu-framework-${PACKAGE_REPOSITORY}-repo-${REPO_VERSION}/tanzu-framework-${PACKAGE_REPOSITORY}-repo-${REPO_VERSION}-lock-output.yaml")
  REPO_VERSION_SHA=$(echo "${REPO_URL}" | cut -d ':' -f '2')
  "${TOOLS_BIN_DIR}"/ytt -f hack/packages/templates/repo-utils/packagerepo-tmpl.yaml -f hack/packages/templates/repo-utils/package-helpers.lib.yaml -f "${PROJECT_ROOT}/${PACKAGE_REPOSITORY}-packages/package-values-sha256.yaml" -v packageRepository="${PACKAGE_REPOSITORY}" -v "${PACKAGE_REPOSITORY}PackageRepository.registry=${REGISTRY}" -v "${PACKAGE_REPOSITORY}PackageRepository.sha256=${REPO_VERSION_SHA}" > "${PACKAGES_BUILD_ARTIFACTS_DIR}/package-repo-bundles/${PACKAGE_REPOSITORY}/tanzu-framework-${PACKAGE_REPOSITORY}-repo-${REPO_VERSION}.yaml"
  rm -rf "${PACKAGES_BUILD_ARTIFACTS_DIR}/package-repo-bundles/${PACKAGE_REPOSITORY}/tanzu-framework-${PACKAGE_REPOSITORY}-repo-${REPO_VERSION}"
}

"$@"
