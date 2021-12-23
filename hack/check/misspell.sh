#!/usr/bin/env bash

# Copyright 2021 VMware, Inc. All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0


set -o errexit
set -o nounset
set -o pipefail


TF_REPO_PATH="$(git rev-parse --show-toplevel)"

MISSPELL_LOC="${TF_REPO_PATH}/hack/check/tools/bin"

# Install tools we need if it is not present
if [[ ! -f "${MISSPELL_LOC}/misspell" ]]; then
  curl -L https://git.io/misspell | bash
  mkdir -p "${MISSPELL_LOC}"
  mv ./bin/misspell "${MISSPELL_LOC}/misspell"
fi

# Spell checking
# misspell check Project - https://github.com/client9/misspell
misspellignore_files="${TF_REPO_PATH}/hack/check/.misspellignore"
ignore_files=$(cat "${misspellignore_files}")
git ls-files | grep -v "${ignore_files}" | xargs "${MISSPELL_LOC}/misspell" | grep "misspelling" && echo "Please fix the listed misspell errors and verify using 'make misspell'" && exit 1 || echo "misspell check passed!"
