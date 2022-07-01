#!/bin/bash
# Copyright 2021 VMware, Inc. All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0

set -eo pipefail

TANZU_BOM_DIR=${HOME}/.config/tanzu/tkg/bom
INSTALL_INSTRUCTIONS='See https://github.com/mikefarah/yq#install for installation instructions'
TKG_CUSTOM_IMAGE_REPOSITORY=${TKG_CUSTOM_IMAGE_REPOSITORY:-''}
TKG_IMAGE_REPO=${TKG_IMAGE_REPO:-''}
TKG_BOM_IMAGE_TAG=${TKG_BOM_IMAGE_TAG:-''}


echodual() {
  echo "$@" 1>&2
  echo "#" "$@"
}

if [ -z "$TKG_CUSTOM_IMAGE_REPOSITORY" ]; then
  echo "TKG_CUSTOM_IMAGE_REPOSITORY variable is required but is not defined" >&2
  exit 1
fi

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
    src=$1
    dst=$2
    echo ""
    echo "imgpkg copy --tar $src.tar --to-repo $dst"
}

if [ -n "$TKG_CUSTOM_IMAGE_REPOSITORY_CA_CERTIFICATE" ]; then
  function imgpkg_copy() {
      src=$1
      dst=$2
      echo ""
      echo "imgpkg copy --tar $src.tar --to-repo $dst --registry-ca-cert-path /tmp/cacrtbase64d.crt"
  }
fi

echo "set -eo pipefail"
echo 'if [ -n "$TKG_CUSTOM_IMAGE_REPOSITORY_CA_CERTIFICATE" ]; then
  echo $TKG_CUSTOM_IMAGE_REPOSITORY_CA_CERTIFICATE > /tmp/cacrtbase64
  base64 -d /tmp/cacrtbase64 > /tmp/cacrtbase64d.crt
fi'
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

    actualTKGImage=${TKG_CUSTOM_IMAGE_REPOSITORY}/tkg-bom
    customTKGImage=tkg-bom-${imageTag}
    imgpkg_copy $customTKGImage $actualTKGImage

    # Get components in the tkg-bom.
    # Remove the leading '[' and trailing ']' in the output of yq.
    components=(`yq e '.components | keys | .. style="flow"' "tmp/$TKG_BOM_FILE" | sed 's/^.//;s/.$//'`)
    for comp in "${components[@]}"
    do
    # remove: leading and trailing whitespace, and trailing comma
    comp=`echo $comp | sed -e 's/^[[:space:]]*//' | sed 's/,*$//g'`
    get_comp_images="yq e '.components[\"${comp}\"][]  | select(has(\"images\"))|.images[] | .imagePath + \":\" + .tag' "\"tmp/\"$TKG_BOM_FILE""

    eval $get_comp_images | while read -r image; do        
        image2=$(echo "$image" | tr ':' '-' | tr '/' '-')
        image3=$(echo "$image" | cut -f1 -d":")
        actualImage=${TKG_CUSTOM_IMAGE_REPOSITORY}/${image3}
        customImage=${image2}
        imgpkg_copy $customImage $actualImage 
      done
    done

    rm -rf tmp
    echodual "Finished processing TKG BOM file ${TKG_BOM_FILE}"
    echo ""
  fi
done

# Iterate through TKR BoM file to create the complete Image name
# and then pull, retag and push image to custom registry.
# DOWNLOAD_TKRS is a space separated list of TKR names, as stored in registry
# when set, the tkr-bom will be ignored.
list=${DOWNLOAD_TKRS:-$(imgpkg  tag  list -i ${actualImageRepository}/tkr-bom)}
for imageTag in ${list}; do
  if [[ ${imageTag} == v* ]]; then
    TKR_BOM_FILE="tkr-bom-${imageTag//_/+}.yaml"
    echodual "Processing TKR BOM file ${TKR_BOM_FILE}"

    actualTKRImage=${TKG_CUSTOM_IMAGE_REPOSITORY}/tkr-bom
    customTKRImage=tkr-bom-${imageTag}
    imgpkg_copy $customTKRImage $actualTKRImage
    imgpkg pull --image ${actualImageRepository}/tkr-bom:${imageTag} --output "tmp" > /dev/null 2>&1

    # Get components in the tkr-bom.
    # Remove the leading '[' and trailing ']' in the output of yq.
    components=(`yq e '.components | keys | .. style="flow"' "tmp/$TKR_BOM_FILE" | sed 's/^.//;s/.$//'`)
    for comp in "${components[@]}"
    do
    # remove: leading and trailing whitespace, and trailing comma
    comp=`echo $comp | sed -e 's/^[[:space:]]*//' | sed 's/,*$//g'`
    get_comp_images="yq e '.components[\"${comp}\"][]  | select(has(\"images\"))|.images[] | .imagePath + \":\" + .tag' "\"tmp/\"$TKR_BOM_FILE""

    eval $get_comp_images | while read -r image; do
        image2=$(echo "$image" | cut -f1 -d":")
        actualImage=${TKG_CUSTOM_IMAGE_REPOSITORY}/${image2}
        image3=$(echo "$image" | tr ':' '-' | tr '/' '-')
        customImage=${image3}
        imgpkg_copy $customImage $actualImage 
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
    actualImage=${TKG_CUSTOM_IMAGE_REPOSITORY}/tkr-compatibility
    customImage=tkr-compatibility-${imageTag}
    imgpkg_copy $customImage $actualImage 
    echo ""
    echodual "Finished processing TKR compatibility image"
  fi
done

list=$(imgpkg  tag  list -i ${actualImageRepository}/tkg-compatibility)
for imageTag in ${list}; do
  if [[ ${imageTag} == v* ]]; then
    echodual "Processing TKG compatibility image"
    actualImage=${TKG_CUSTOM_IMAGE_REPOSITORY}/tkg-compatibility
    customImage=tkg-compatibility-${imageTag}
    imgpkg_copy $customImage $actualImage 
    echo ""
    echodual "Finished processing TKG compatibility image"
  fi
done
