#!/usr/bin/env bash

# Copyright 2021 VMware, Inc. All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0

# shellcheck disable=SC1091
. ../helpers.sh

normalize original.yaml src
normalize generated.yaml dst

#vimdiff src dst
kapp tools diff -c --file2 dst --file src
