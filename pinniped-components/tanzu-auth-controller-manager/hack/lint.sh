#!/usr/bin/env bash

# Copyright 2022 VMware, Inc. All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0

set -euo pipefail

MY_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
PINNIPED_DIR="${MY_DIR}/../.."

# Always run from tanzu-auth-controller-manager directory for reproducibility
cd "${MY_DIR}/.."

# Install golangci-lint
make golangci-lint -C "${PINNIPED_DIR}/hack/tools"

# Run linter
"${PINNIPED_DIR}/hack/tools/bin/golangci-lint" run -v
