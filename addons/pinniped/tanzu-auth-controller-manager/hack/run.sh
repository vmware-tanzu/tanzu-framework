#!/bin/bash

# Copyright 2022 VMware, Inc. All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0

set -euo pipefail

YELLOW='\033[0;33m'
DEFAULT='\033[0m'

MY_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
TF_ROOT="${MY_DIR}/../../../.."

APP_NAME="tanzu-auth"
KAPP_CONTROLLER_NAME="${APP_NAME}-ctrl"

# location for the package install
PACKAGES_DIR="${TF_ROOT}/packages"
TANZU_AUTH_DIR="${PACKAGES_DIR}/${APP_NAME}"
TANZU_AUTH_BUNDLE_DIR="${TANZU_AUTH_DIR}/bundle"
TANZU_AUTH_CONFIG_DIR="${TANZU_AUTH_BUNDLE_DIR}/config"

TANZU_AUTH_PACKAGE_FILE="${TANZU_AUTH_DIR}/package.yaml"

# location to deploy the image
REGISTRY_NAME="harbor-repo.vmware.com"
REGISTRY_PROJECT="tkgiam"
REGISTRY="${REGISTRY_NAME}/${REGISTRY_PROJECT}"

# name of image to match that defined in /packages/tanzu-auth
CONTROLLER_MANAGER_NAME="${APP_NAME}-controller-manager"
CONTROLLER_MANAGER_PACKAGE_NAME="${CONTROLLER_MANAGER_NAME}-package"
CONTROLLER_NAMESPACE_NAME="${APP_NAME}"

# Always run from tanzu-auth-controller-manager directory for reproducibility
cd "${MY_DIR}/.."

TAG="dev"
# TAG="$(uuidgen)" # Uncomment to create random image every time

# build the tanzu-auth-controller-manager image
# --------------------------------------------------------
CONTROLLER_IMAGE="${REGISTRY}/$(whoami)/${CONTROLLER_MANAGER_NAME}:${TAG}"
echo -e "${YELLOW}building ${CONTROLLER_MANAGER_NAME} image and pushing to ${CONTROLLER_IMAGE}...${DEFAULT}"

# push the tanzu-auth-controller-manager to the registry
docker build -t "${CONTROLLER_IMAGE}" .
docker push "${CONTROLLER_IMAGE}"


# build the tanzu-auth-controller-manager-package image
# --------------------------------------------------------
PACKAGE_IMAGE="${REGISTRY}/$(whoami)/${CONTROLLER_MANAGER_PACKAGE_NAME}:${TAG}"
echo -e "${YELLOW}building ${CONTROLLER_MANAGER_PACKAGE_NAME} image and pushing to ${PACKAGE_IMAGE}...${DEFAULT}"

# generate new RBAC into /packages/tanzu-auth/bundle/config
echo -e "${YELLOW}generating RBAC...${DEFAULT}"
./hack/generate.sh

# the namespace must exist before the package can be deployed
echo -e "${YELLOW}creating namespace ${APP_NAME}...${DEFAULT}"
ytt --data-value "namespace=${APP_NAME}" --file "${TANZU_AUTH_CONFIG_DIR}/namespace.yaml" | kubectl apply -f -

echo -e "${YELLOW}rebuilding image lock file with ytt and kbld...${DEFAULT}"
mkdir -p "${TANZU_AUTH_BUNDLE_DIR}/.imgpkg"
ytt --data-value "controller.image=${CONTROLLER_IMAGE}" --file "${TANZU_AUTH_CONFIG_DIR}" \
  | kbld --file - --imgpkg-lock-output "${TANZU_AUTH_BUNDLE_DIR}/.imgpkg/images.yml"

# push the tanzu-auth-controller-manager-package to the registry
echo -e "${YELLOW}pushing ${PACKAGE_IMAGE} to registry with imgpkg...${DEFAULT}"
imgpkg push --bundle "${PACKAGE_IMAGE}" --file "${TANZU_AUTH_BUNDLE_DIR}"

# inject the package name into the package file & deploy
echo -e "${YELLOW}creating package ${TANZU_AUTH_PACKAGE_FILE} with image ${PACKAGE_IMAGE}...${DEFAULT}"
yq e ".spec.template.spec.fetch[0].imgpkgBundle.image = \"${PACKAGE_IMAGE}\"" "${TANZU_AUTH_PACKAGE_FILE}" \
  | kubectl apply -f -
kubectl apply -f "${TANZU_AUTH_DIR}/metadata.yaml"


# Deploy the package on the cluster
echo -e "${YELLOW}creating package install...${DEFAULT}"
PACKAGE_SA_NAME="${CONTROLLER_MANAGER_PACKAGE_NAME}-sa"
PACKAGE_NAMESPACE="tkg-system"
PACKAGE_INSTALL="$(cat <<EOF
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: ${PACKAGE_SA_NAME}
  namespace: ${PACKAGE_NAMESPACE}
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: ${PACKAGE_SA_NAME}-cluster-admin
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: cluster-admin
subjects:
- kind: ServiceAccount
  name: ${PACKAGE_SA_NAME}
  namespace: ${PACKAGE_NAMESPACE}
---
apiVersion: v1
kind: Secret
metadata:
  name: ${CONTROLLER_MANAGER_PACKAGE_NAME}-config
  namespace: ${PACKAGE_NAMESPACE}
stringData:
  values.yaml: |
    namespace: ${CONTROLLER_NAMESPACE_NAME}
    controller:
      image: ${CONTROLLER_IMAGE}
---
apiVersion: packaging.carvel.dev/v1alpha1
kind: PackageInstall
metadata:
  name: ${APP_NAME}
  namespace: ${PACKAGE_NAMESPACE}
spec:
  packageRef:
    refName: ${APP_NAME}.tanzu.vmware.com
    versionSelection:
      constraints: "1.6.0"
      prereleases: {}
  syncPeriod: 30s
  serviceAccountName: ${PACKAGE_SA_NAME}
  values:
  - secretRef:
      name: ${CONTROLLER_MANAGER_PACKAGE_NAME}-config
EOF)"

echo -e "${YELLOW}package install contents...${DEFAULT}"
echo "${PACKAGE_INSTALL}"

echo -e "${YELLOW}applying package install contents...${DEFAULT}"
echo "${PACKAGE_INSTALL}" | kubectl apply -f -

echo -e "${YELLOW}waiting for reconciliation...${DEFAULT}"
kubectl wait --timeout=1m --for=condition=ReconcileSucceeded app "${APP_NAME}" -n "${PACKAGE_NAMESPACE}"

# Tail the logs
kapp inspect --app "${APP_NAME}-ctrl" -n "${PACKAGE_NAMESPACE}"
kapp logs --app "${APP_NAME}-ctrl" -n "${PACKAGE_NAMESPACE}" -f
