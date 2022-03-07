#!/usr/bin/env bash

set -xeuo pipefail

MY_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"

# Always run from pinniped directory for reproducibility.
cd "${MY_DIR}/.."

# Test post-deploy job
make -C ./post-deploy test

# Test config-controller
./config-controller/hack/check.sh
