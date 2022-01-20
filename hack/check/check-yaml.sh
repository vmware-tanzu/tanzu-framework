#!/bin/bash

# Copyright 2021 VMware, Inc. All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0

set -o nounset
set -o pipefail

TF_REPO_PATH="$(git rev-parse --show-toplevel)"

docker run --rm -t cytopia/yamllint --version
CONTAINER_NAME="tf_yamllint_$RANDOM"
docker run --name ${CONTAINER_NAME} -t -v "${TF_REPO_PATH}":/tanzu-framework:ro cytopia/yamllint -s -c /tanzu-framework/hack/check/.yamllintconfig.yaml /tanzu-framework
EXIT_CODE=$(docker inspect ${CONTAINER_NAME} --format='{{.State.ExitCode}}')
docker rm -f ${CONTAINER_NAME} &> /dev/null

if [[ ${EXIT_CODE} == "0" ]]; then
  echo "yamllint passed!"
else
  echo "yamllint exit code ${EXIT_CODE}: YAML linting failed!"
  echo "Please fix the listed yamllint errors and verify using 'make yamllint'"
  exit "${EXIT_CODE}"
fi
