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
GOBINDATA=$2
GOOS=$3
GOARCH=$4

cd pinniped && GOARCH=${GOARCH} GOOS=${GOOS} ${GO} build -o pinniped ./cmd/pinniped && cd ..
${GOBINDATA} -mode=420 -modtime=1 -o=pkg/v1/auth/tkg/zz_generated.bindata.go -pkg=tkgauth pinniped/pinniped
git update-index --assume-unchanged pkg/v1/auth/tkg/zz_generated.bindata.go
