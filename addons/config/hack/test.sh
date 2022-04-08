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
IAAS_INPUT="vsphere"

function verifyAddonConfigTemplateForGVR() {
  # test with default inputs
  TEST_RESULT=$(${TF_TOOL_DIR}/ytt --ignore-unknown-comments -f templates/${1}/${2}/${3}.yaml -v TKR_VERSION=${TKR_INPUT} -v GLOBAL_NAMESPACE=${NAMESPACE_INPUT} -v IAAS=${IAAS_INPUT})
  EXPECTED="$(cat expected/${1}/${2}/${3}.yaml)"

  if [[ "${TEST_RESULT}" != "${EXPECTED}" ]]
  then
    echo -e "$(tput setaf 1)Failed to run template sanity test.\nDefault config generation does not match expected output\n$(tput sgr 0)"
    echo -e "result: \n${TEST_RESULT}\n"
    echo -e "expected: \n${EXPECTED}\n"
    diff <(echo "${TEST_RESULT}") <(echo "${EXPECTED}")
    exit 1
  fi
}

function verifyAddonConfigTemplateForGVRWithIaas() {
    TEST_RESULT=$(${TF_TOOL_DIR}/ytt --ignore-unknown-comments -f templates/${1}/${2}/${3}.yaml -v TKR_VERSION=${TKR_INPUT} -v GLOBAL_NAMESPACE=${NAMESPACE_INPUT} -v IAAS=${4})
    EXPECTED="$(cat expected/${1}/${2}/${3}-${4}.yaml)"

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
        verifyAddonConfigTemplateForGVR "${array[1]}" "${array[2]}" "${array[3]%.yaml}"
        echo "-- Successfully did sanity check on ${kindPath}"
      done
		done
	done

  # additionally, add test coverage for VSphereCPIConfig when iaas is tkgs, i.e. paravirtual mode
	verifyAddonConfigTemplateForGVRWithIaas "cpi.tanzu.vmware.com" "v1alpha1" "vspherecpiconfig" "tkgs"
  echo "-- Successfully did sanity check on cpi.tanzu.vmware.com/v1alpha1/vspherecpiconfig.yaml with iaas tkgs"
}

"$@"
