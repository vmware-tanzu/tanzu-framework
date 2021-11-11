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
shift
GOOS=$1
shift
GOARCH=$1
shift

while (( "$#" )); do
  pinniped_version="$1"
  pinniped_binary="cmd/cli/plugin/pinniped-auth/asset/pinniped-${pinniped_version}"
  echo "embed-pinniped-binary.sh: building pinniped version '$pinniped_version' to '$pinniped_binary'"

  pushd pinniped >/dev/null
    git checkout "$pinniped_version"
    GOARCH=${GOARCH} GOOS=${GOOS} ${GO} build -o "../${pinniped_binary}" ./cmd/pinniped
  popd >/dev/null

  git update-index --assume-unchanged "$pinniped_binary"

  shift
done
