#!/bin/bash

# Copyright 2022 VMware, Inc. All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0

set -euo pipefail

MY_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
TF_ROOT="${MY_DIR}/../../../.."

# location for the package install
MANAGEMENT_DIR="${TF_ROOT}/packages/management"
TANZU_AUTH_DIR="${MANAGEMENT_DIR}/tanzu-auth"
TANZU_AUTH_BUNDLE_DIR="${TANZU_AUTH_DIR}/bundle"
TANZU_AUTH_CONFIG_DIR="${TANZU_AUTH_BUNDLE_DIR}/config"
TANZU_AUTH_PACKAGE="${TANZU_AUTH_CONFIG_DIR}/package.yaml"

# location to deploy the image
REGISTRY_NAME="harbor-repo.vmware.com"
REGISTRY_PROJECT="tkgiam"
REGISTRY="${REGISTRY_NAME}/${REGISTRY_PROJECT}"

# name of image to match that defined in /packages/management/tanzu-auth
DEPLOYMENT_NAME="tanzu-auth-controller-manager"
NAMESPACE_NAME="tanzu-auth"

# Always run from config-controller directory for reproducibility
cd "${MY_DIR}/.."

tag="dev"
# tag="$(uuidgen)" # Uncomment to create random image every time
image="${REGISTRY}/$(whoami)/${DEPLOYMENT_NAME}:$tag"

# generate new RBAC into /packages/management/tanzu-auth/bundle/config
./hack/generate.sh

# the namespace must exist before the package can be deployed
kubectl apply -f "${TANZU_AUTH_CONFIG_DIR}/namespace.yaml"

docker build -t "$image" .
docker push "$image"

# deploy via the packages/management/tanzu-auth package
ytt --data-value "image=$image" -f "${TANZU_AUTH_CONFIG_DIR}" | kbld -f - --imgpkg-lock-output "${TANZU_AUTH_BUNDLE_DIR}/.imgpkg/images.yaml" | kapp deploy -a "${DEPLOYMENT_NAME}" -f - -y

kapp inspect --app "${DEPLOYMENT_NAME}" --tree
kapp logs --app "${DEPLOYMENT_NAME}" -f

