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

  #TODO: Automate configure-bom as part of the build process instead
  exit 0
else
  echo "OK"
fi
