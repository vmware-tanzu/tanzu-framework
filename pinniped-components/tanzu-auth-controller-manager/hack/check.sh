#!/usr/bin/env bash

# Copyright 2022 VMware, Inc. All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0

set -xeuo pipefail

MY_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"

# Always run from tanzu-auth-controller-manager directory for reproducibility
cd "${MY_DIR}/.."

# Make sure all the files are formatted
test -z "$(go fmt ./...)" || (echo "files were not properly formatted per 'go fmt'" && exit 1)

# Make sure all the files pass 'go vet'
go vet ./... || error "'go vet' failed"

# Make sure all the files pass t-f linting config
./hack/lint.sh

# Make sure our tests pass.
./hack/test.sh

# Make sure our default secret generation script works as expected
./hack/generate-package-secret.test.sh
