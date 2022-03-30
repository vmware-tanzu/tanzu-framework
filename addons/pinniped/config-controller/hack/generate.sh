#!/usr/bin/env bash

# Copyright 2022 VMware, Inc. All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0

set -euo pipefail

MY_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
TF_ROOT="${MY_DIR}/../../../.."

# Always run from config-controller directory for reproducibility
cd "${MY_DIR}/.."

# Install controller-gen
make controller-gen -C "${TF_ROOT}/hack/tools"

# generate RBAC rules
"${TF_ROOT}/hack/tools/bin/controller-gen" paths=./controllers +rbac:roleName=pinniped-config-controller-manager output:rbac:dir=./manifests
