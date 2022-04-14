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

function verifyAddonConfigTemplateForGVR() {
    TEST_RESULT=$(${TF_TOOL_DIR}/ytt --ignore-unknown-comments -f templates/${1}/${2}/${3}.yaml \
      -v TKR_VERSION=${TKR_INPUT} -v GLOBAL_NAMESPACE=${NAMESPACE_INPUT} \
      -f "testcases/${1}/${2}/${3}/${4}.yaml")
    EXPECTED="$(cat "expected/${1}/${2}/${3}/${4}.yaml")"

    if [[ "${TEST_RESULT}" != "${EXPECTED}" ]]
    then
      echo -e "$(tput setaf 1)Failed to run template sanity test.\nDefault config generation does not match expected output\n$(tput sgr 0)"
      echo -e "result: \n${TEST_RESULT}\n"
      echo -e "expected: \n${EXPECTED}\n"
      diff <(echo "${TEST_RESULT}") <(echo "${EXPECTED}")
      exit 1
    fi
}

function verifyAllAddonConfigTemplates() {
  # scan for config CRs with all the versions
  echo "Checking all addon Config CR templates..."
	for groupPath in "templates"/*; do
		for versionPath in "${groupPath}"/*; do
		  for kindPath in "${versionPath}"/*; do
		    IFS='/' read -r -a array <<< "${kindPath}"

        testcasePath="testcases/${array[1]}/${array[2]}/${array[3]%.yaml}"
		    if [ ! -d "${testcasePath}" ]; then
          echo "-- Test cases are not provided for ${kindPath}"
        else
          for testcase in "${testcasePath}"/*; do
            IFS='/' read -r -a array <<< "${testcase}"
            verifyAddonConfigTemplateForGVR "${array[1]}" "${array[2]}" "${array[3]}" "${array[4]%.yaml}"
            echo "-- Successfully did sanity check on ${kindPath} with data values ${testcase}"
          done
        fi
      done
		done
	done
}

"$@"
