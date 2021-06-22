#!/bin/bash
# Copyright 2021 VMware, Inc. All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0

set -euo pipefail

TANZU_BOM_DIR=${HOME}/.config/tanzu/tkg/bom
INSTALL_INSTRUCTIONS='See https://github.com/mikefarah/yq#install for installation instructions'

echodual() {
  echo "$@" 1>&2
  echo "#" "$@"
}

if [ -z "$TKG_CUSTOM_IMAGE_REPOSITORY" ]; then
  echo "TKG_CUSTOM_IMAGE_REPOSITORY variable is not defined" >&2
  exit 1
fi

if [[ -d "$TANZU_BOM_DIR" ]]; then
  BOM_DIR="${TANZU_BOM_DIR}"
else
  echo "Tanzu Kubernetes Grid directories not found. Run CLI once to initialise." >&2
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

echo "set -euo pipefail"
echodual "Note that yq must be version above or equal to version 4.5 and below version 5."

actualImageRepository=""
# Iterate through BoM file to create the complete Image name
# and then pull, retag and push image to custom registry.
for TKG_BOM_FILE in "$BOM_DIR"/*.yaml; do
  echodual "Processing BOM file ${TKG_BOM_FILE}"
  # Get actual image repository from BoM file
  actualImageRepository=$(yq e '.imageConfig.imageRepository' "$TKG_BOM_FILE")
  yq e '.. | select(has("images"))|.images[] | .imagePath + ":" + .tag ' "$TKG_BOM_FILE" |
    while read -r image; do
      actualImage=${actualImageRepository}/${image}
      customImage=$TKG_CUSTOM_IMAGE_REPOSITORY/${image}
      echo "docker pull $actualImage"
      echo "docker tag  $actualImage $customImage"
      echo "docker push $customImage"
      echo ""
    done
  echodual "Finished processing BOM file ${TKG_BOM_FILE}"
  echo ""
done

# Iterate through TKR BoM file to create the complete Image name
# and then pull, retag and push image to custom registry.
list=$(imgpkg  tag  list -i "${actualImageRepository}"/tkr-bom)
for imageTag in ${list}; do
  if [[ ${imageTag} == v* ]]; then 
    TKR_BOM_FILE="tkr-bom-${imageTag//_/+}.yaml"
    echodual "Processing TKR BOM file ${TKR_BOM_FILE}"

    actualTKRImage=${actualImageRepository}/tkr-bom:${imageTag}
    customTKRImage=${TKG_CUSTOM_IMAGE_REPOSITORY}/tkr-bom:${imageTag}
    echo ""
    echo "docker pull $actualTKRImage"
    echo "docker tag  $actualTKRImage $customTKRImage"
    echo "docker push $customTKRImage"
    imgpkg pull --image "${actualImageRepository}/tkr-bom:${imageTag}" --output "tmp" > /dev/null 2>&1
    yq e '.. | select(has("images"))|.images[] | .imagePath + ":" + .tag ' "tmp/$TKR_BOM_FILE" |
    while read -r image; do
      actualImage=${actualImageRepository}/${image}
      customImage=$TKG_CUSTOM_IMAGE_REPOSITORY/${image}
      echo "docker pull $actualImage"
      echo "docker tag  $actualImage $customImage"
      echo "docker push $customImage"
      echo ""
    done
    rm -rf tmp
    echodual "Finished processing TKR BOM file ${TKR_BOM_FILE}"
    echo ""
  fi
done

list=$(imgpkg  tag  list -i "${actualImageRepository}"/tkr-compatibility)
for imageTag in ${list}; do
  if [[ ${imageTag} == v* ]]; then
    echodual "Processing TKR compatibility image"
    actualImage=${actualImageRepository}/tkr-compatibility:${imageTag}
    customImage=$TKG_CUSTOM_IMAGE_REPOSITORY/tkr-compatibility:${imageTag}
    echo ""
    echo "docker pull $actualImageRepository/tkr-compatibility:$imageTag"
    echo "docker tag  $actualImage $customImage"
    echo "docker push $customImage"
    echo ""
    echodual "Finished processing TKR compatibility image"
  fi
done
