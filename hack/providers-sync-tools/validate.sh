#!/bin/bash

# Copyright 2022 VMware, Inc. All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0

set -e
set -o pipefail

function validate_rendered_files() {
	local source_package_name="${1}"
	local provider_name="${2}"
	local rendered_dir="${source_package_name}/build/rendered"

	local version=$(cat ../../packages/${source_package_name}/vendir.yml | grep tag | sed 's/^.*tag: //')

	local rendered_files="$(find "${rendered_dir}/upstream" -type f)"
	set +e
	for rendered_file in ${rendered_files}; do
	  diff -s "../../providers/${provider_name}/${version}/$(basename ${rendered_file})" "${rendered_file}"
	  local exit_code="${?}"
	  if [[ ${exit_code} -ne 0 ]]; then
		echo "[Error] Files ${rendered_file} and ../../providers/${provider_name}/${version}/$(basename ${rendered_file}) are different."
	    echo "[Error] providers is out of sync with packages/cluster-api*. See 'hack/providers-sync-tools/README.md'."
	    exit ${exit_code}
	  fi
	done
	set -e
}

validate_rendered_files "cluster-api" "cluster-api"
validate_rendered_files "cluster-api-control-plane-kubeadm" "control-plane-kubeadm"
validate_rendered_files "cluster-api-bootstrap-kubeadm" "bootstrap-kubeadm"
validate_rendered_files "cluster-api-provider-aws" "infrastructure-aws"
validate_rendered_files "cluster-api-provider-azure" "infrastructure-azure"
validate_rendered_files "cluster-api-provider-docker" "infrastructure-docker"
validate_rendered_files "cluster-api-provider-vsphere" "infrastructure-vsphere"
