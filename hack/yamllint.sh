#!/bin/bash -eu

set -o pipefail

BASE_DIR="$(dirname "$0")"

TARGET_DIR=${BASE_DIR}/../pkg/v1/providers

pushd ${TARGET_DIR}

for i in $(find . -path ./ytt/vendir -prune -false -o -name "*yaml"); do
  echo linting $i
  cat $i | yamllint -
done

popd
