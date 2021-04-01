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
KUBEVAL="${TOOLS_BIN_DIR}"/kubeval
OUTPUT_DIR="${1:-"${ROOT_DIR}"/template_tests/.rendered-output}"

################################################################################
##                               RENDER MANIFESTS                             ##
################################################################################
"${ROOT_DIR}"/scripts/render-manifests.sh "${OUTPUT_DIR}"

################################################################################
##                               KUBEVAL                                      ##
################################################################################
# TODO: Should remove the "--skip-kinds Namespace" after issue TKG-4221 was fixed.
#  The "skip-kinds" flag is added temporarily to deal with
# duplicate Namespace definitions across Pinniped template base-files
# TODO: Find a way to fix missing 'metadata' key without ignoring the whole file (--skip-kinds didn't work for Config kind in kapp-config.yaml)
${KUBEVAL} --skip-kinds Namespace --ignored-filename-patterns 'antrea.yaml' --ignore-missing-schemas --strict --force-color -d "${OUTPUT_DIR}"
