#!/bin/bash

# Copyright 2022 VMware, Inc. All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0

set -e
set -o pipefail
ROOT="$(cd "$(dirname ${0})/../.." &> /dev/null && pwd)"

function render_upstream_package_file() {
	local source_provider_folder="${1}"
	local provider_bundle_folder="${2}"
	local component_file_prefix="${3}"
	local infrastructure_components_file="${source_provider_folder}/${component_file_prefix}.yaml"
	local overlay_file="${source_provider_folder}/change-image-url.yaml"
	local output_file="${provider_bundle_folder}/${source_provider_folder}"
	
	${ROOT}/hack/tools/bin/ytt -f "${infrastructure_components_file}" \
		-f "${overlay_file}" --output-files "${output_file}"
}

render_upstream_package_file "cluster-api/v1.2.4" ${1} "core-components"
render_upstream_package_file "bootstrap-kubeadm/v1.2.4" ${1} "bootstrap-components"
render_upstream_package_file "infrastructure-aws/v2.0.0-beta.1" ${1} "infrastructure-components"
render_upstream_package_file "infrastructure-azure/v1.5.3" ${1} "infrastructure-components"
render_upstream_package_file "infrastructure-vsphere/v1.4.1" ${1} "infrastructure-components"
