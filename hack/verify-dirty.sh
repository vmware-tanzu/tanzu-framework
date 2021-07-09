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
   git diff HEAD -- .
   exit 1
else
   echo "OK"
fi


echo
echo "#############################"
echo "Verify make configure-bom..."
echo "#############################"
make configure-bom
if ! (git diff --quiet HEAD -- .); then
  echo "FAIL"
  echo "'make configure-bom' generated diffs!"
  echo "Please verify if default BOM variable changes are intended and commit the diffs if so."
  exit 1
else
  echo "OK"
fi

echo
echo "#############################"
echo "Verify make providers..."
echo "#############################"
make providers > /dev/null
if ! (git diff --quiet HEAD -- .); then
  git diff --stat
  echo "FAIL"
  echo "'make providers' detected changes to provider files but checksum/bindata have not been updated."
  echo "Please verify if provider changes are intended and commit the generated files if so."
  exit 1
else
  echo "OK"
fi
