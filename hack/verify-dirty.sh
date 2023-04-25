#!/bin/bash
# Copyright 2021 VMware, Inc. All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0

set -euo pipefail

echo "#############################"
echo "Verify uncommitted files..."
echo "#############################"

if ! (git diff --quiet HEAD -- .); then
   echo -e "\nThe following files are uncommitted. Please commit them or add them to .gitignore:";
   git diff --name-only HEAD -- . | awk '{print "- " $0}'
   echo -e "\nDiff:"
   git --no-pager diff  HEAD -- .
   exit 1
else
   echo "OK"
fi

echo
echo "#############################"
echo "Verify make package-vendir-sync..."
echo "#############################"
make package-vendir-sync
echo "Done vendir sync!"

# This is required for github actions execution
# because the package-vendir-sync operation renders
# the pakcage directory contents inaccessible to the
# host machine user.
GITHUB_ACTIONS="${GITHUB_ACTIONS:-}"
if [[ -n "${GITHUB_ACTIONS}" ]]; then
  echo "Resetting directory ownership"
  sudo chown -R $(id -nu):$(id -ng) .;
fi

if ! (git diff --quiet HEAD -- .); then
  echo "FAIL"
  echo "'make package-vendir-sync' generated diffs!"
  echo "Please verify if package CRD changes are intended and commit the diffs if so."
  git diff --name-only HEAD -- . | awk '{print "- " $0}'
  echo -e "\nDiff:"
  git --no-pager diff  HEAD -- .
  exit 1 
else
  echo "OK"
fi
