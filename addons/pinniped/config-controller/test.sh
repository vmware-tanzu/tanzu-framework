#!/usr/bin/env bash

MY_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
TF_ROOT="${MY_DIR}/../../.."

# Always run from config-controller directory for reproducibility
cd "${MY_DIR}"

# Install kubebuilder
make kubebuilder -C "${TF_ROOT}/hack/tools"

# Run tests
KUBEBUILDER_ASSETS="${TF_ROOT}/hack/tools/bin/kubebuilder/bin" go test ./... -coverprofile cover.out -v 2