#!/bin/bash

# Copyright 2022 VMware, Inc. All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0

set -e
set -o pipefail

function place_rendered_files() {
	local source_package_name="${1}"
	local provider_name="${2}"
	local rendered_dir="${source_package_name}/build/rendered"

	local version=$(cat ../../packages/${source_package_name}/vendir.yml | grep tag | sed 's/^.*tag: //')

	cp -r "${rendered_dir}/upstream/." "../../providers/${provider_name}/${version}/"
}

place_rendered_files "cluster-api" "cluster-api"
place_rendered_files "cluster-api-control-plane-kubeadm" "control-plane-kubeadm"
place_rendered_files "cluster-api-bootstrap-kubeadm" "bootstrap-kubeadm"
place_rendered_files "cluster-api-provider-aws" "infrastructure-aws"
place_rendered_files "cluster-api-provider-azure" "infrastructure-azure"
place_rendered_files "cluster-api-provider-docker" "infrastructure-docker"
place_rendered_files "cluster-api-provider-vsphere" "infrastructure-vsphere"
