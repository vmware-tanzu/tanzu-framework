#!/bin/bash

# Copyright 2022 VMware, Inc. All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0

set -euo pipefail

MY_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
TF_ROOT="${MY_DIR}/../../../.."
PACKAGE_ROOT="${TF_ROOT}/packages/management/pinniped-config-controller-manager"

# Always run from config-controller directory for reproducibility
cd "${MY_DIR}/.."

# Perform some environment validations
if ! kubectl get pods; then
  echo "error: must run this script while targeting cluster"
  exit 1
fi
if ! kubectl get deploy -A -o name | grep -q kapp-controller; then
  echo "error: kapp-controller deployment must be running on targeted cluster"
  exit 1
fi

tag="dev"
# tag="$(uuidgen)" # Uncomment to create random image every time
controller_image="harbor-repo.vmware.com/tkgiam/$(whoami)/pinniped-config-controller-manager:$tag"
package_image="harbor-repo.vmware.com/tkgiam/$(whoami)/pinniped-config-controller-manager:$tag"

# Build pinniped-config-controller-manager image
docker build -t "$controller_image" .
docker push "$controller_image"

# Ensure generated deployment YAML (e.g., RBAC)
./hack/generate.sh

# Tell package to map default pinniped-config-controller-manager image to dev image
cat <<EOF | kbld -f - --imgpkg-lock-output "${PACKAGE_ROOT}/bundle/.imgpkg/images.yml"
---
apiVersion: kbld.k14s.io/v1alpha1
kind: Config
overrides:
- image: pinniped-config-controller-manager:latest  # This image is hardcoded in the package
  newImage: ${controller_image}                     # This image is the dev image built above
EOF

# Build the package
imgpkg push -b "$package_image" -f "${PACKAGE_ROOT}/bundle"

# Create the package on the cluster
kubectl apply -f "${PACKAGE_ROOT}"
cat <<EOF | kubectl apply -f -
apiVersion: packaging.carvel.dev/v1alpha1
kind: PackageInstall
metadata:
  name: pinniped-config-controller-manager
  namespace: tkg-system
spec:
  packageRef:
    refName: pinniped-config-controller-manager.tanzu.vmware.com
    versionSelection:
      prereleases: {}
EOF

# Wait for the app to be reconciled
kubectl wait --timeout 1m --for condition=ReconcileSucceeded -n tkg-system app pinniped-config-controller-manager

# Tail the logs
kubectl logs -n pinniped deploy/pinniped-config-controller-manager -f
