#!/usr/bin/env bash

set -xeuo pipefail

MY_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
TF_ROOT="${MY_DIR}/../../../.."

# Always run from config-controller directory for reproducibility
cd "${MY_DIR}/.."

# Make sure all the files are formatted
test -z "$(go fmt ./...)" || (echo "files were not properly formatted per 'go fmt'" && exit 1)

# Make sure all the files pass 'go vet'
go vet ./... || error "'go vet' failed"

# Make sure all the files pass t-f linting config
make golangci-lint -C "${TF_ROOT}/hack/tools"
"${TF_ROOT}/hack/tools/bin/golangci-lint" run -v

# Make sure our tests pass.
./hack/test.sh