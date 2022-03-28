#!/usr/bin/env bash

# Copyright 2022 VMware, Inc. All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0

set -euo pipefail

MY_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
TF_ROOT="${MY_DIR}/../../../.."

# Always run from config-controller directory for reproducibility
cd "${MY_DIR}/.."

# test with no args
TEST_RESULT_1=$(./hack/generate-package-secret.sh)
EXPECTED_1="apiVersion: v1
kind: Secret
metadata:
  name: default-pinniped-config-v0.0.0
stringData:
  values.yaml: |
    infrastructure_provider: vsphere
    identity_management_type: none
    tkg_cluster_role: workload"

if [[ "${TEST_RESULT_1}" != "${EXPECTED_1}" ]]
then
  echo "default secret generation does not match expected output"
  echo -e "result: \n${TEST_RESULT_1}"
  echo -e "expected: \n${EXPECTED_1}"
  exit 1
fi

# test with provided args
TEST_RESULT_2=$(./hack/generate-package-secret.sh -v tkr=foo -v infrastructure_provider=ALTERNATIVE_IAAS_TO_VSPHERE)
EXPECTED_2="apiVersion: v1
kind: Secret
metadata:
  name: default-pinniped-config-foo
stringData:
  values.yaml: |
    infrastructure_provider: ALTERNATIVE_IAAS_TO_VSPHERE
    identity_management_type: none
    tkg_cluster_role: workload"

if [[ "${TEST_RESULT_2}" != "${EXPECTED_2}" ]]
then
  echo "secret generation with parameters does not match expected output"
  echo -e "result: \n${TEST_RESULT_2}"
  echo -e "expected: \n${EXPECTED_2}"
  exit 1
fi
