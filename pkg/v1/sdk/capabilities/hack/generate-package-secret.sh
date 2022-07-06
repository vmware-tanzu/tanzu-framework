#!/usr/bin/env bash

# Copyright 2022 VMware, Inc. All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0

set -euo pipefail

MY_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
TF_ROOT="${MY_DIR}/../../../../../"

# Always run from capabilities directory for reproducibility
cd "${MY_DIR}/.."

# Run YTT, passing the arguments to this script straight through to YTT to ease future extensions template
ytt -f ./hack/ytt/schema.yaml -f ./hack/ytt/default-package-secret.yaml -f ./hack/ytt/values.yaml $@
