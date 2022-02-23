#!/bin/bash
# Copyright 2021 VMware, Inc. All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0

set -euo pipefail

TANZU_BOM_DIR=${HOME}/.config/tanzu/tkg/bom
INSTALL_INSTRUCTIONS='See https://github.com/mikefarah/yq#install for installation instructions'
TKG_IMAGE_REPO=${TKG_IMAGE_REPO:-''}
TKG_BOM_IMAGE_TAG=${TKG_BOM_IMAGE_TAG:-''}


echodual() {
  echo "$@" 1>&2
  echo "#" "$@"
}

if [ -z "$TKG_IMAGE_REPO" ]; then
  echo "TKG_IMAGE_REPO variable is required but is not defined" >&2
  exit 2
fi

if ! [ -x "$(command -v imgpkg)" ]; then
  echo 'Error: imgpkg is not installed.' >&2
  exit 3
fi

if ! [ -x "$(command -v yq)" ]; then
  echo 'Error: yq is not installed.' >&2
  echo "${INSTALL_INSTRUCTIONS}" >&2
  exit 3
fi

function imgpkg_copy() {
    flags=$1
    src=$2
    dst=$3
    echo ""
    echo "imgpkg copy $flags $src --to-tar $dst.tar"
}

echo "set -euo pipefail"
echodual "Note that yq must be version above or equal to version 4.9.2 and below version 5."

actualImageRepository="$TKG_IMAGE_REPO"

# Iterate through TKG BoM file to create the complete Image name
# and then pull, retag and push image to custom registry.
list=$(imgpkg  tag  list -i "${actualImageRepository}"/tkg-bom)
for imageTag in ${list}; do
  tanzucliversion=$(tanzu version | head -n 1 | cut -c10-15)
  if [[ ${imageTag} == ${tanzucliversion}* ]] || [[ ${imageTag} == ${TKG_BOM_IMAGE_TAG} ]]; then
    TKG_BOM_FILE="tkg-bom-${imageTag//_/+}.yaml"
    imgpkg pull --image "${actualImageRepository}/tkg-bom:${imageTag}" --output "tmp" > /dev/null 2>&1
    echodual "Processing TKG BOM file ${TKG_BOM_FILE}"

    actualTKGImage=${actualImageRepository}/tkg-bom:${imageTag}
    customTKGImage=tkg-bom-${imageTag}
    imgpkg_copy "-i" $actualTKGImage $customTKGImage

    # Get components in the tkg-bom.
    # Remove the leading '[' and trailing ']' in the output of yq.
    components=(`yq e '.components | keys | .. style="flow"' "tmp/$TKG_BOM_FILE" | sed 's/^.//;s/.$//'`)
    for comp in "${components[@]}"
    do
    # remove: leading and trailing whitespace, and trailing comma
    comp=`echo $comp | sed -e 's/^[[:space:]]*//' | sed 's/,*$//g'`
    get_comp_images="yq e '.components[\"${comp}\"][]  | select(has(\"images\"))|.images[] | .imagePath + \":\" + .tag' "\"tmp/\"$TKG_BOM_FILE""

    flags="-i"
    if [ $comp = "tkg-standard-packages" ] || [ $comp = "standalone-plugins-package" ] || [ $comp = "tanzu-framework-management-packages" ]; then
      flags="-b"
    fi
    eval $get_comp_images | while read -r image; do
        actualImage=${actualImageRepository}/${image}
        image2=$(echo "$image" | tr ':' '-' | tr '/' '-')
        customImage=${image2}
        imgpkg_copy $flags $actualImage $customImage
      done
    done

    rm -rf tmp
    echodual "Finished processing TKG BOM file ${TKG_BOM_FILE}"
    echo ""
  fi
done

# Iterate through TKR BoM file to create the complete Image name
# and then pull, retag and push image to custom registry.
list=$(imgpkg  tag  list -i ${actualImageRepository}/tkr-bom)
for imageTag in ${list}; do
  if [[ ${imageTag} == v* ]]; then
    TKR_BOM_FILE="tkr-bom-${imageTag//_/+}.yaml"
    echodual "Processing TKR BOM file ${TKR_BOM_FILE}"

    actualTKRImage=${actualImageRepository}/tkr-bom:${imageTag}
    customTKRImage=tkr-bom-${imageTag}
    imgpkg_copy "-i" $actualTKRImage $customTKRImage
    imgpkg pull --image ${actualImageRepository}/tkr-bom:${imageTag} --output "tmp" > /dev/null 2>&1

    # Get components in the tkr-bom.
    # Remove the leading '[' and trailing ']' in the output of yq.
    components=(`yq e '.components | keys | .. style="flow"' "tmp/$TKR_BOM_FILE" | sed 's/^.//;s/.$//'`)
    for comp in "${components[@]}"
    do
    # remove: leading and trailing whitespace, and trailing comma
    comp=`echo $comp | sed -e 's/^[[:space:]]*//' | sed 's/,*$//g'`
    get_comp_images="yq e '.components[\"${comp}\"][]  | select(has(\"images\"))|.images[] | .imagePath + \":\" + .tag' "\"tmp/\"$TKR_BOM_FILE""

    flags="-i"
    if [ $comp = "tkg-core-packages" ]; then
      flags="-b"
    fi
    eval $get_comp_images | while read -r image; do
        actualImage=${actualImageRepository}/${image}
        image2=$(echo "$image" | tr ':' '-' | tr '/' '-')
        customImage=${image2}
        imgpkg_copy $flags $actualImage $customImage
      done
    done

    rm -rf tmp
    echodual "Finished processing TKR BOM file ${TKR_BOM_FILE}"
    echo ""
  fi
done

list=$(imgpkg  tag  list -i ${actualImageRepository}/tkr-compatibility)
for imageTag in ${list}; do
  if [[ ${imageTag} == v* ]]; then
    echodual "Processing TKR compatibility image"
    actualImage=${actualImageRepository}/tkr-compatibility:${imageTag}
    customImage=tkr-compatibility-${imageTag}
    imgpkg_copy "-i" $actualImage $customImage
    echo ""
    echodual "Finished processing TKR compatibility image"
  fi
done

list=$(imgpkg  tag  list -i ${actualImageRepository}/tkg-compatibility)
for imageTag in ${list}; do
  if [[ ${imageTag} == v* ]]; then
    echodual "Processing TKG compatibility image"
    actualImage=${actualImageRepository}/tkg-compatibility:${imageTag}
    customImage=tkg-compatibility-${imageTag}
    imgpkg_copy "-i" $actualImage $customImage
    echo ""
    echodual "Finished processing TKG compatibility image"
  fi
done
