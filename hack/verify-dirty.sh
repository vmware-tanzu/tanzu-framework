#!/bin/bash
# Copyright 2021 VMware, Inc. All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0

set -euo pipefail

if ! (git diff --quiet HEAD -- .); then
   echo -e "\nThe following files are uncommitted. Please commit them or add them to .gitignore:";
   git diff --name-only HEAD -- . | awk '{print "- " $0}'
   echo -e "\nDiff:"
   git diff HEAD -- . ; exit 1;
fi

