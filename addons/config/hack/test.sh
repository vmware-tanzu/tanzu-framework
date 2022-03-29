#!/usr/bin/env bash

# Copyright 2022 VMware, Inc. All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0

set -euo pipefail

MY_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
TF_ROOT="${MY_DIR}/../../.."
TF_TOOL_DIR="${TF_ROOT}/hack/tools/bin"

# Always run from addons/config directory
cd "${MY_DIR}/.."

# Default test inputs
TKR_INPUT="v1.23.3---vmware.1-tkg.1"
NAMESPACE_INPUT="tkg-system"

# test with default inputs
TEST_RESULT=$(${TF_TOOL_DIR}/ytt --ignore-unknown-comments -f templates/${1}/${2}/${3}.yaml -v TKR_VERSION=${TKR_INPUT} -v GLOBAL_NAMESPACE=${NAMESPACE_INPUT})
EXPECTED="$(cat expected/${1}/${2}/${3}.yaml)"

if [[ "${TEST_RESULT}" != "${EXPECTED}" ]]
then
  echo -e "$(tput setaf 1)Failed to run template sanity test.\nDefault config generation does not match expected output\n$(tput sgr 0)"
  echo -e "result: \n${TEST_RESULT}\n"
  echo -e "expected: \n${EXPECTED}\n"
  diff <(echo "${TEST_RESULT}") <(echo "${EXPECTED}")
  exit 1
fi
