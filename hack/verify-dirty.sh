#!/bin/bash
# Copyright 2021 VMware, Inc. All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0

set -euo pipefail

echo "#############################"
echo "Verify uncommitted files..."
echo "#############################"

# Option to ignore changes to compatibility image path in
# tkg/tkgconfigpaths/zz_bundled_default_bom_files_configdata.go and plugin runtime version in
# cli/runtime/version/zz_generated_plugin_runtime_version.go. This is to support `usebom` directive
# in tests.
ignore_files=(':(exclude)tkg/tkgconfigpaths/zz_bundled_default_bom_files_configdata.go' ':(exclude)cli/runtime/version/zz_generated_plugin_runtime_version.go')
# Temporarily excluding UI generated bindata file from verify (zz_generated.bindata.go). Currently running into issues
# blocking the CI main build. CI generated bindata is different from bindata file generated on local
# developer machines, causing this failure. Need to root cause and then remove this exclusion.
ignore_file_ui_bindata=':!tkg/manifest/server/zz_generated.bindata.go'

if ! (git diff --quiet HEAD -- . "${ignore_files[@]}" "${ignore_file_ui_bindata}"); then
   echo -e "\nThe following files are uncommitted. Please commit them or add them to .gitignore:";
   git diff --name-only HEAD -- . "${ignore_files[@]}" "${ignore_file_ui_bindata}" | awk '{print "- " $0}'
   echo -e "\nDiff:"
   git --no-pager diff  HEAD -- . "${ignore_files[@]}" "${ignore_file_ui_bindata}"
   exit 1
else
   echo "OK"
fi


echo
echo "#############################"
echo "Verify make package-vendir-sync..."
echo "#############################"
make package-vendir-sync
if ! (git diff --quiet HEAD -- . "${ignore_files[@]}" "${ignore_file_ui_bindata}"); then
  echo "FAIL"
  echo "'make package-vendir-sync' generated diffs!"
  echo "Please verify if package CRD changes are intended and commit the diffs if so."
  exit 1 
else
  echo "OK"
fi


echo
echo "#############################"
echo "Verify make configure-bom..."
echo "#############################"
make configure-bom
if ! (git diff --quiet HEAD -- . "${ignore_files[@]}" "${ignore_file_ui_bindata}"); then
  echo "FAIL"
  echo "'make configure-bom' generated diffs!"
  echo "Please verify if default BOM variable changes are intended and commit the diffs if so."
  #TODO: Automate configure-bom as part of the build process instead
  exit 0
else
  echo "OK"
fi
