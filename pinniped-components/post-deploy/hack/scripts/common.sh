#!/usr/bin/env bash

# Copyright 2021 VMware, Inc. All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0

set -o errexit
set -o nounset
set -o pipefail

################################################################################
##                               FUNCTIONS                                    ##
################################################################################
function sha256 {
  local cmd

  if command -v sha256sum &> /dev/null; then
    cmd=(sha256sum)
  elif command -v shasum &> /dev/null; then
    cmd=(shasum -a 256)
  else
    echo "ERROR: could not find shasum or sha256sum."
    return 1
  fi

  "${cmd[@]}" "$@"
}

function sha1 {
  local cmd

  if command -v sha1sum &> /dev/null; then
    cmd=(sha1sum)
  elif command -v shasum &> /dev/null; then
    cmd=(shasum -a 1)
  else
    echo "ERROR: could not find shasum or sha1sum."
    return 1
  fi

  "${cmd[@]}" "$@"
}
