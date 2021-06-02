#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

################################################################################
##                               GLOBAL VARIABLES                             ##
################################################################################
ROOT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )/.." >/dev/null 2>&1 && pwd )"
PAR_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )/../.." >/dev/null 2>&1 && pwd )"
TOOLS_BIN_DIR="${PAR_DIR}/hack/tools/bin"
YTT="${TOOLS_BIN_DIR}"/ytt
OUTPUT_DIR="${1:-"${ROOT_DIR}"/template_tests/.rendered-output}"

################################################################################
##                               FUNCTIONS                                    ##
################################################################################
function render_manifest() {
  local component=$1
  output_dir="${OUTPUT_DIR}/$component"
  mkdir -p ${OUTPUT_DIR}/$component
  ytt="${YTT} --ignore-unknown-comments --dangerous-allow-all-symlink-destinations -f ${ROOT_DIR}/ytt-common-libs/ -f ${ROOT_DIR}/$component/templates/"
  if [[ -d ${ROOT_DIR}/template_tests/testcases/$component/ ]]; then
    for f in ${ROOT_DIR}/template_tests/testcases/$component/*; do
      sub_component="$(basename "$f" | cut -f 1 -d '.')"
      ytt_input=" -f ${ROOT_DIR}/template_tests/testcases/$component/$sub_component"
      eval "${ytt}${ytt_input}.yaml" > "${output_dir}/${component}-${sub_component}.yaml"
    done
  else
    eval "${ytt}" > "${output_dir}/${component}.yaml"
  fi
}
################################################################################
##                               RENDER MANIFESTS                             ##
################################################################################
render_manifest pinniped
render_manifest calico
render_manifest antrea
render_manifest metrics-server
render_manifest kapp-controller
render_manifest addons-manager
render_manifest vsphere_csi
render_manifest vsphere_cpi
render_manifest ako-operator
