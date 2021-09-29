#!/bin/bash

# Copyright 2021 VMware, Inc. All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0

# Usage:
# rebuild-cli.sh [provider-repo-path]

GOOS=$(go env GOOS)
GOARCH=$(go env GOARCH)
CLI_REPO=${CLI_REPO:-${PWD}/../..}

pushd ${CLI_REPO}

#go mod edit -replace github.com/vmware-tanzu/tkg-providers=$1
#cat go.mod
#
# Build tkg cli
mkdir -p ${CLI_REPO}/pkg/v1/tkg/web/dist # Add web/dist directory if missing for building cli
make tkg-cli

# revert updated go.mod file of tkg-cli repo which was updated before building cli
#go mod edit -replace github.com/vmware-tanzu/tkg-providers=./providers

popd
