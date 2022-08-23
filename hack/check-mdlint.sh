#!/bin/bash

# Copyright 2021 VMware, Inc. All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0

set -o errexit
set -o nounset
set -o pipefail

# Change directories to the parent directory of the one in which this
# script is located.
cd "$(dirname "${BASH_SOURCE[0]}")/.."

# mdlint rules with common errors and possible fixes can be found here:
# https://github.com/DavidAnson/markdownlint/blob/main/doc/Rules.md
# Additional configuration can be found in the .markdownlintrc file at
# the root of the repo.
docker run --rm -v "$(pwd)":/build \
  gcr.io/cluster-api-provider-vsphere/extra/mdlint:0.23.2 /md/lint \
  -i **/CHANGELOG.md \
  -i pkg/v1/tkg/web/node_modules \
  -i docs/cli/commands \
  -i test/cli/mock \
  -i providers/ytt/vendir \
  -i providers/provider-bundle/providers/ytt/vendir \
  -i pinniped .
