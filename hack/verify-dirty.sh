#!/bin/bash
# Copyright 2021 VMware, Inc. All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0

set -euo pipefail

#TODO: Remove package-lock.json from this list once dirty file issue is resolved.
if ! (git diff --quiet HEAD -- . ':(exclude)pkg/v1/tkg/web/package-lock.json'); then
   echo -e "\nThe following files are uncommitted. Please commit them or add them to .gitignore:";
   git diff --name-only HEAD -- . | awk '{print "- " $0}'
   echo -e "\nDiff:"
   git diff HEAD -- . ; exit 1;
fi

