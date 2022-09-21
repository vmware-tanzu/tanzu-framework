#!/bin/bash

# Copyright 2022 VMware, Inc. All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0

set -e
set -o pipefail
ROOT="$(cd "$(dirname ${0})/../.." &> /dev/null && pwd)"

function render_upstream_package_file() {
	local source_package_name="${1}"
	local build_dir="${source_package_name}/build"
	local pkg_v1_providers_specific_overlays_dir="${source_package_name}/overlay"
	local syncd_upstream_dir="${build_dir}/upstream"
	local rendered_dir="${build_dir}/rendered"

	mkdir -p "${rendered_dir}"
	mkdir -p "${pkg_v1_providers_specific_overlays_dir}"

	${ROOT}/hack/tools/bin/ytt -f "${syncd_upstream_dir}" \
		-f "${pkg_v1_providers_specific_overlays_dir}" \
		--output-files "${rendered_dir}"
	find "${rendered_dir}" -type f -exec chmod 664 {} \;
}

render_upstream_package_file "cluster-api-control-plane-kubeadm"
render_upstream_package_file "cluster-api"
render_upstream_package_file "cluster-api-bootstrap-kubeadm"
render_upstream_package_file "cluster-api-provider-aws"
render_upstream_package_file "cluster-api-provider-azure"
render_upstream_package_file "cluster-api-provider-docker"
render_upstream_package_file "cluster-api-provider-vsphere"

# infrastructure-docker: The infrastructure-components.yaml is named
# infrastructure-components-development.yaml upstream. Renaming it to
# infrastructure-components.yaml here to match providers naming.
mv cluster-api-provider-docker/build/rendered/upstream/infrastructure-components{-development,}.yaml
