#!/usr/bin/env bash

# Copyright 2022 VMware, Inc. All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0

set -euo pipefail

MY_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
TF_ROOT="${MY_DIR}/../../../.."

# Always run from capabilities directory for reproducibility
cd "${MY_DIR}/.."

# test with no args
TEST_RESULT_1=$(./hack/generate-package-secret.sh)
EXPECTED_1="apiVersion: v1
kind: Secret
metadata:
  name: default-capabilities-package-config-v0.0.0
stringData:
  values.yaml: |
    namespace: tkg-system
    deployment:
      hostNetwork: true
      nodeSelector: {}
      tolerations: []
    rbac:
      podSecurityPolicyNames: []"

if [[ "${TEST_RESULT_1}" != "${EXPECTED_1}" ]]
then
  echo "default secret generation does not match expected output"
  echo -e "result: \n${TEST_RESULT_1}"
  echo -e "expected: \n${EXPECTED_1}"
  exit 1
fi

# test with provided args
TEST_RESULT_2=$(./hack/generate-package-secret.sh -v tkr=foo --data-value-yaml 'rbac.podSecurityPolicyNames=[test-psp,test-psp-two]')
EXPECTED_2="apiVersion: v1
kind: Secret
metadata:
  name: default-capabilities-package-config-foo
stringData:
  values.yaml: |
    namespace: tkg-system
    deployment:
      hostNetwork: true
      nodeSelector: {}
      tolerations: []
    rbac:
      podSecurityPolicyNames:
      - test-psp
      - test-psp-two"

if [[ "${TEST_RESULT_2}" != "${EXPECTED_2}" ]]
then
  echo "secret generation with parameters does not match expected output"
  echo -e "result: \n${TEST_RESULT_2}"
  echo -e "expected: \n${EXPECTED_2}"
  exit 1
fi
