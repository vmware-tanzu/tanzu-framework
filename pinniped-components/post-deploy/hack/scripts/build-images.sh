#!/usr/bin/env bash

# Copyright 2021 VMware, Inc. All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0

set -o errexit
set -o nounset
set -o pipefail

ROOT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )/../.." >/dev/null 2>&1 && pwd )"
# shellcheck source=./hack/scripts/common.sh
source "${ROOT_DIR}/hack/scripts/common.sh"

REPO_NAME="${REPO_NAME:-vmware.io}"
IMAGE_NAME="${IMAGE_NAME:-tkg-pinniped-post-deploy}"
IMAGE_TAG="${VERSION//+/_}"
FULL_IMAGE_NAME="${REPO_NAME}/${IMAGE_NAME}:${IMAGE_TAG}"
FULL_IMAGE_TAR_NAME="${IMAGE_NAME}-${IMAGE_TAG}"

# Build from publicly reachable source by default, but allow people to re-build images on
# top of their own trusted images.
BUILDER_BASE_IMAGE="${BUILDER_BASE_IMAGE:-}"
if [[ -z "${BUILDER_BASE_IMAGE}" ]];
then
  docker build \
    -t "${FULL_IMAGE_NAME}" \
    -f "${ROOT_DIR}"/Dockerfile ..
else
  docker build \
    --build-arg BUILDER_BASE_IMAGE="${BUILDER_BASE_IMAGE}" \
    -t "${FULL_IMAGE_NAME}" \
    -f "${ROOT_DIR}"/Dockerfile ..
fi

mkdir -p "${ROOT_DIR}"/artifacts/images
cd "${ROOT_DIR}"/artifacts/images
docker save "${FULL_IMAGE_NAME}" | gzip -c > "${FULL_IMAGE_TAR_NAME}.tar.gz"

IMAGE_ID=$(docker inspect -f '{{.ID}}' "${FULL_IMAGE_NAME}")
echo "${REPO_NAME}/${IMAGE_NAME}@${IMAGE_ID}" > "${FULL_IMAGE_TAR_NAME}-image-digests.txt"

sha256 "${FULL_IMAGE_TAR_NAME}-image-digests.txt" "${FULL_IMAGE_TAR_NAME}.tar.gz" > "${FULL_IMAGE_TAR_NAME}-image-checksums.txt"
