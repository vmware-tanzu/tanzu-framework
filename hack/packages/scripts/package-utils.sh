#!/usr/bin/env bash

# Copyright 2021 VMware, Inc. All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0

set -eoux pipefail

function trivy_scan() {
  tmp_dir=$(mktemp -d)
  while IFS='|' read -r image; do
    "${TOOLS_BIN_DIR}"/trivy --cache-dir "$tmp_dir" image --exit-code 1 --severity CRITICAL --ignore-unfixed "$image"
  done < <("${TOOLS_BIN_DIR}"/yq e ".overrides[] | .newImage" "${PROJECT_ROOT}/packages/kbld-config.yaml")
}

"$@"
