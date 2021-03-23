#!/bin/bash
# Copyright 2021 VMware, Inc. All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0

set -o nounset

if ! (git diff --quiet HEAD -- .); then
git diff HEAD -- . \
   echo "you haven’t committed files you’re supposed to commit"; exit 1; \
fi

