#!/usr/bin/env bash

# Copyright 2022 VMware, Inc. All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0

set -xeuo pipefail

MY_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"

# Always run from pinniped directory for reproducibility.
cd "${MY_DIR}/.."

# Test post-deploy job
make -C ./post-deploy test

# Test config-controller
./config-controller/hack/check.sh
