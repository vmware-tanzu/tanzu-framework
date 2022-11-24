#!/bin/bash

# Copyright 2022 VMware, Inc. All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0

set -e
set -x
set -o pipefail
ROOT="$(cd "$(dirname ${0})/../.." &> /dev/null && pwd)"

function render_upstream_package_file_to_provider() {
	local source_package_name="${1}"
	local target_package_name="${2}"
	local version=$(cat ../../packages/${source_package_name}/vendir.yml | grep tag | sed 's/^.*tag: //')
	local build_dir="${source_package_name}/build"
	local pkg_v1_providers_specific_overlays_dir="${source_package_name}/overlay"
	local update_image_file="change-image-url.yaml"
	local syncd_upstream_dir="${build_dir}/upstream"
	local rendered_dir="../../providers/${target_package_name}/${version}/"

	mkdir -p "${rendered_dir}"
	mkdir -p "${pkg_v1_providers_specific_overlays_dir}"

	${ROOT}/hack/tools/bin/ytt -f "${syncd_upstream_dir}" \
		-f "${pkg_v1_providers_specific_overlays_dir}" \
		--file-mark 'change-image-url.yaml:exclude=true' \
		--output-files "${rendered_dir}"
	cp -r "${pkg_v1_providers_specific_overlays_dir}/${update_image_file}" "${rendered_dir}"
	mv "${rendered_dir}/upstream"/* "${rendered_dir}/upstream/../"
	find "${rendered_dir}" -type f -exec chmod 664 {} \;
}

render_upstream_package_file_to_provider "cluster-api-control-plane-kubeadm" "control-plane-kubeadm"
render_upstream_package_file_to_provider "cluster-api" "cluster-api"
render_upstream_package_file_to_provider "cluster-api-bootstrap-kubeadm" "bootstrap-kubeadm"
render_upstream_package_file_to_provider "cluster-api-provider-aws" "infrastructure-aws"
render_upstream_package_file_to_provider "cluster-api-provider-azure" "infrastructure-azure"
render_upstream_package_file_to_provider "cluster-api-provider-docker" "infrastructure-docker"
render_upstream_package_file_to_provider "cluster-api-provider-vsphere" "infrastructure-vsphere"

# infrastructure-docker: The infrastructure-components.yaml is named
# infrastructure-components-development.yaml upstream. Renaming it to
# infrastructure-components.yaml here to match providers naming.
mv ../../providers/infrastructure-docker/v1.2.4/infrastructure-components{-development,}.yaml
