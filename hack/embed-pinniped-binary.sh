#!/bin/bash
# Copyright 2021 VMware, Inc. All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0

# A helper script to be invoked prior to building plugins for a particular target (os/arch)
# Specifically affects the correct building of the pinniped-auth plugin since it needs to embed the
# pinniped client built for the same os/arch that it is built for.
# TODO : remove when we no longer need to embed said binary and can directly link to the pinniped
# repo code.

set -o errexit
set -o nounset
set -o pipefail
set -o xtrace

GO=$1
GOOS=$2
GOARCH=$3

cd pinniped && GOARCH=${GOARCH} GOOS=${GOOS} ${GO} build -o ../cmd/cli/plugin/pinniped-auth/asset/pinniped ./cmd/pinniped && cd ..
git update-index --assume-unchanged cmd/cli/plugin/pinniped-auth/asset/pinniped
