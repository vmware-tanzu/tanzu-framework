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
TESTS_DIR="${2:-"${ROOT_DIR}"/template_tests/tests}"
COMPONENTS="${3:-pinniped,calico,antrea,metrics-server,kapp-controller,addons-manager,vsphere_csi,vsphere_cpi}"

################################################################################
##                               FUNCTIONS                                    ##
################################################################################
function diff_manifests() {
  local component=$1
  diff="diff -qr"
  tests_prefix="${TESTS_DIR}/${component}"
  output_prefix=${OUTPUT_DIR}/${component}
  echo "GENERATE DIFF MANIFESTS FOR $component"
  if [[ "${COMPONENTS[@]}" =~ "${component}" ]]; then
    if [[ -d ${ROOT_DIR}/template_tests/testcases/$component/ ]]; then
      for f in ${ROOT_DIR}/template_tests/testcases/$component/*; do
        sub_component="$(basename "$f")"
        eval "${diff}" "${output_prefix}/${component}-${sub_component}" "${tests_prefix}/${sub_component}"
      done
    else
      eval "${diff}" "${output_prefix}/${component}.yaml" "${tests_prefix}/${component}.yaml"
    fi
  fi
}
################################################################################
##                               RENDER MANIFESTS                             ##
################################################################################
"${ROOT_DIR}"/scripts/render-manifests.sh

################################################################################
##                               DIFF TEST                                    ##
################################################################################
diff_manifests pinniped
diff_manifests calico
diff_manifests antrea
diff_manifests metrics-server
diff_manifests kapp-controller
diff_manifests addons-manager
diff_manifests vsphere_csi
diff_manifests vsphere_cpi

