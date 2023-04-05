#!/usr/bin/env bash

# Copyright 2022 VMware, Inc. All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0

# This script is typically invoked by a make target in the top-level Makefile.
# For TKGm, that make target is called by the bolt CLI code during the build process.

set -euo pipefail

MY_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"

# Always run from tanzu-auth-controller-manager directory for reproducibility
cd "${MY_DIR}/.."

# Run YTT, passing the arguments to this script straight through to YTT to ease future extensions template
ytt -f ./hack/ytt/schema.yaml -f ./hack/ytt/default-package-secret.yaml -f ./hack/ytt/values.yaml $@
