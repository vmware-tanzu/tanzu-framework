#!/bin/bash
# Copyright 2021 VMware, Inc. All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0

set -euo pipefail

echo "#############################"
echo "Verify uncommitted files..."
echo "#############################"

# Option to ignore changes to compatibility image path in
# pkg/v1/tkg/tkgconfigpaths/zz_bundled_default_bom_files_configdata.go. This is to support `usebom` directive
# in tests.
ignore_file=':!pkg/v1/tkg/tkgconfigpaths/zz_bundled_default_bom_files_configdata.go'

if ! (git diff --quiet HEAD -- . "${ignore_file}"); then
   echo -e "\nThe following files are uncommitted. Please commit them or add them to .gitignore:";
   git diff --name-only HEAD -- . "${ignore_file}" | awk '{print "- " $0}'
   echo -e "\nDiff:"
   git --no-pager diff  HEAD -- . "${ignore_file}"
   exit 1
else
   echo "OK"
fi


echo
echo "#############################"
echo "Verify make configure-bom..."
echo "#############################"
make configure-bom
if ! (git diff --quiet HEAD -- . "${ignore_file}"); then
  echo "FAIL"
  echo "'make configure-bom' generated diffs!"
  echo "Please verify if default BOM variable changes are intended and commit the diffs if so."
  #TODO: Automate configure-bom as part of the build process instead
  exit 0
else
  echo "OK"
fi
