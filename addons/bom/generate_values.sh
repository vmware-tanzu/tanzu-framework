#!/bin/bash

ROOT_DIR=$2
ADDON_FOLDER=$1
IMG_TAG=$3

# get variables from the Makefile of each add-on
filename="$ROOT_DIR$ADDON_FOLDER/Makefile"
img_name=$(grep -m 1 IMG_NAME $filename | sed 's/^.*= //g')
addon_name=$(grep -m 1 ADDON_NAME $filename | sed 's/^.*= //g')
category=$(grep -m 1 IMG_CATEGORY $filename | sed 's/^.*= //g')
cluster_type=$(grep -m 1 IMG_CLUSTER_TYPE $filename | sed 's/^.*= //g')
img_tag=$IMG_TAG

[ -z "$img_name" ] || [ -z "$addon_name" ] || [ -z "$category" ] || [ -z "$cluster_type" ] || [ -z "$img_tag" ] && echo "Missing variable in $ADDON_FOLDER" && exit 1

target_file="bom.yaml"
# generate the target bom file
echo "  $addon_name:" >> $target_file
echo "    category: $category" >> $target_file
echo "    clusterTypes:" >> $target_file
for type in $cluster_type
do
    echo "      - $type" >> $target_file
done
echo "    templatesImagePath: tanzu_core/addons/$img_name" >> $target_file
echo "    templatesImageTag: $img_tag" >> $target_file