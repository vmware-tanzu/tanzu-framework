#!/bin/bash
# Copyright 2018-2021 VMware, Inc. All Rights Reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#    http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
set -eu -o pipefail && [ -n "${DEBUG:-}" ] && set -x

DEFAULT_API_ENDPOINT="https://api.github.com/repos/"
DEFAULT_HEADERS=("Accept: application/vnd.github.symmetra-preview+json")
DEFAULT_CURL_ARGS=("-s")
DEFAULT_REPO="vmware-tanzu/tanzu-framework"
DEFAULT_MAX_LABELS=1000

API_ENDPOINT=${API_ENDPOINT:-${DEFAULT_API_ENDPOINT}}
HEADERS=("${HEADERS[@]:-${DEFAULT_HEADERS[@]}}")
CURL_ARGS=("${CURL_ARGS[@]:-${DEFAULT_CURL_ARGS[@]}}")
REPO=${REPO:-${DEFAULT_REPO}}
MAX_LABELS=${MAX_LABELS:-${DEFAULT_MAX_LABELS}}

HEADERS=("${HEADERS[@]}" "Authorization: token ${GITHUB_TOKEN?"GitHub API token must be supplied (create one at https://github.com/settings/tokens with the repo scope)"}")
HEADER_ARGS=("${HEADERS[@]/#/"-H"}")

# Colors from https://material.io/design/color/#tools-for-picking-colors
declare -A colors
colors=(
  [red]="FFCDD2"
  [pink]="F8BBD0"
  [purple]="E1BEE7"
  ["deep purple"]="D1C4E9"
  [indigo]="C5CAE9"
  [blue]="BBDEFB"
  ["light blue"]="B3E5FC"
  [cyan]="B2EBF2"
  [teal]="B2DFDB"
  [green]="C8E6C9"
  ["light green"]="DCEDC8"
  [lime]="F0F4C3"
  [yellow]="FFF9C4"
  [amber]="FFECB3"
  [orange]="FFE0B2"
  ["deep orange"]="FFCCBC"
  [brown]="D7CCC8"
  [gray]="F5F5F5"
  ["blue gray"]="CFD8DC"
)

declare -A assigned_colors
assigned_colors=(
  [status]="${colors[red]}"
  [area]="${colors[cyan]}"
  [commitment]="${colors[purple]}"
  [consumer]="${colors[amber]}"
  [impact]="${colors[brown]}"
  [kind]="${colors[indigo]}"
  [source]="${colors[green]}"
  [resolution]="${colors["blue gray"]}"
)


# Determines whether a label already exists
#
# Arguments:
# 1: the label name
#
# Returns:
# N/A
#
# Exits:
# 0: the label exists
# 1: the label does not exist
label-exists () {
    : "${1?"Usage: ${FUNCNAME[0]} LABEL"}"

    args=("-w%{http_code}\n" "${HEADER_ARGS[@]}" "${CURL_ARGS[@]}")
    code=$(curl "${args[@]}" "${API_ENDPOINT%/}/${REPO}/labels/${1// /%20}" | tail -n1)

    [ "$code" -eq 200 ]
}

# Updates the description and color associated with an existing label
#
# Arguments:
# 1: the label name
# 2: the label description
# 3: the label color
#
# Returns:
# N/A
#
# Exits:
# 0: the operation succeeded
# 1: the operation failed
label-update () {
    : "${2?"Usage: ${FUNCNAME[0]} LABEL DESCRIPTION [COLOR]"}"

    if [ -z "$3" ]
    then
        data="{\"description\": \"$2\"}"
    else
        data="{\"description\": \"$2\", \"color\": \"$3\"}"
    fi
    args=("--data" "${data}" "-XPATCH" "-w%{http_code}\n" "${HEADER_ARGS[@]}" "${CURL_ARGS[@]}")
    code=$(curl "${args[@]}" "${API_ENDPOINT%/}/${REPO}/labels/${1// /%20}" | tail -n1)

    [ "$code" -eq 200 ]
}

# Creates a label with the given description and color
#
# Arguments:
# 1: the label name
# 2: the label description
# 3: the label color
#
# Returns:
# N/A
#
# Exits:
# 0: the operation succeeded
# 1: the operation failed
label-create () {
    : "${2?"Usage: ${FUNCNAME[0]} LABEL DESCRIPTION [COLOR]"}"

    if [ -z "$3" ]
    then
        data="{\"name\":\"$1\", \"description\": \"$2\"}"
    else
        data="{\"name\":\"$1\", \"description\": \"$2\", \"color\": \"$3\"}"
    fi
    args=("--data" "${data}" "-w%{http_code}\n" "${HEADER_ARGS[@]}" "${CURL_ARGS[@]}")
    code=$(curl "${args[@]}" "${API_ENDPOINT%/}/${REPO}/labels" | tail -n1)

    [ "$code" -eq 201 ]
}

# Creates a label with the given description and color, or updates one that exists
#
# Arguments:
# 1: the label name
# 2: the label description
# 3: the label color
#
# Returns:
# N/A
#
# Exits:
# 0: the operation succeeded
# 1: the operation failed
label-merge () {
    : "${2?"Usage: ${FUNCNAME[0]} LABEL DESCRIPTION [COLOR]"}"

    if label-exists "$1"
    then
        label-update "$1" "$2" "$3"
    else
        label-create "$1" "$2" "$3"
    fi
}

# Creates a set of labels with a common prefix, updating the description and color of existing labels as necessary
#
# Arguments:
# 1: the label prefix
# 2: (pass-by-name) an associative array of label to description, with underscores instead of slashes
# 3: the color for labels with the supplied prefix
#
# Returns:
# Warning strings about any unexpected labels which already exist with a given prefix
merge () {
    : "${3?"Usage: ${FUNCNAME[0]} PREFIX {LABEL:DESCRIPTION} COLOR"}"

    prefix="$1"
    l="$( declare -p "$2" )"
    eval "declare -A labels=${l#*=}"
    color="$3"

    expected=()
    # The array is declared in the eval above
    # shellcheck disable=SC2154
    for label in "${!labels[@]}"; do
        name="${prefix}/${label//_/\/}"
        description="${labels[$label]}"

        label-merge "${name}" "${description}" "${color}"

        expected+=("${name}")
    done

    args=("${HEADER_ARGS[@]}" "${CURL_ARGS[@]}")
    existing=("$(curl "${args[@]}" "${API_ENDPOINT%/}/${REPO}/labels?per_page=${MAX_LABELS}" | \
               jq ".[] | .name | select(select(startswith(\"${prefix}/\")) | in({$(printf '"%s":0,' "${expected[@]}")}) != true)")")
    if [ -n "${existing[*]}" ]; then
        printf "WARNING: unexpected ${prefix} label %s\n" "${existing[@]}"
    fi
}

merge-oneoff () {
    label-merge "good first issue" " Good for newcomers " "${colors[lime]}"
    label-merge "help wanted" "A well-defined issue on which a pull request would be especially welcome" "${colors[orange]}"

    label-merge "ok-to-merge" "PRs should be labelled with this before merging" "${colors[teal]}"
    label-merge "do-not-merge/hold" "This PR should not be merged (specify reason when adding label)" "d73a4a"

    label-merge "cla-not-required" "" "ffffff"
    label-merge "cla-rejected" "" "fc2929"
}

merge-severity () {
    label-merge "severity/0-maximal" "Among the most severe issues imaginable. Use sparingly." "D32F2F" # (E.g., guaranteed data loss)
    label-merge "severity/1-critical" "Relates to a key use-case of the product. Often impacts many users." "E64A19" # (E.g., component crashes; missing feature blocking mass adoption)
    label-merge "severity/2-serious" "High usability or functional impact. Often has no workaround." "F57C00" # (E.g., advertised functionality does not work; core part of a new feature; refactoring code that is a recurring pain point)
    label-merge "severity/3-moderate" "Medium usability or functional impact. Potentially has an inconvenient workaround." "FFA000" # (E.g., API fails intermittently, but can be retried; optional part of a new feature; refactoring to improve maintainability)
    label-merge "severity/4-minor" "Low usability or functional impact. Often has an easy workaround." "FBC02D" # (E.g., short form of CLI argument causes failure, but long form works fine; nice-to-have part of a new feature; refactoring to improve readability)
    label-merge "severity/5-minimal" "Does not affect the ability the use the product in any way." "689F38" # (E.g., a typo which does not affect clarity; purely asthetic refactoring)
}

merge-areas () {
    declare -A areas
    # The array is passed by name at the end of this function
    # shellcheck disable=SC2034
    areas=(
            [api-machinery]=""
            [bootstrap]=""
            [cli-machinery]=""
            [cli-plugins]=""
            [cluster-lifecycle]=""
            [docs]=""
            [iam]=""
            [packages]=""
            [release]=""
            [repo-maintenance]=""
            [security]=""
            [testing]=""
    )

    merge "area" areas "${assigned_colors[area]}"
}

merge-commitments () {
    declare -A commitments
    # The array is passed by name at the end of this function
    # shellcheck disable=SC2034
    commitments=(
        [consumer]="Affects a commitment made to a consumer (please include consumer/* label)"
        [release]="Affects a release scope commitment (please specify a milestone)"
        [roadmap]="Affects a roadmap commitment"
        [other]="Affects another type of commitment"
    )

    merge "commitment" commitments "${assigned_colors[commitment]}"
}

merge-consumers () {
    declare -A consumers
    # The array is passed by name at the end of this function
    # shellcheck disable=SC2034
    consumers=(
            [TCE]="Related to Tanzu Community Edition"
            [TKG]="Related to Tanzu Kubernetes Grid"
            [TMC]="Related to Tanzu Mission Control"
    )

    merge "consumer" consumers "${assigned_colors[consumer]}"
}

merge-impacts () {
    declare -A impacts
    # The array is passed by name at the end of this function
    # shellcheck disable=SC2034
    impacts=(
        [doc_community]="Requires changes to documentation about contributing to the project"
        [doc_design]="Requires changes to documentation about the design of the project"
        [doc_note]="Requires creation of or changes to a release note"
        [doc_user]="Requires changes to user documentation"
    )

    merge "impact" impacts "${assigned_colors[impact]}"
}

merge-kinds () {
    declare -A kinds
    # The array is passed by name at the end of this function
    # shellcheck disable=SC2034
    kinds=(
        [debt]="Problems that increase the cost of other work"
        [defect]="Behavior that is inconsistent with what's intended"
        [defect_performance]="Behavior that is functionally correct, but performs worse than intended"
        [defect_regression]="Changed behavior that is inconsistent with what's intended"
        [defect_security]="A flaw or weakness that could lead to a violation of security policy"
        [enhancement]="Behavior that was intended, but we want to make better"
        [feature]="New functionality you could include in marketing material"
        [task]="Work not related to changing the functionality of the project"
        [question]="A request for information"
        [investigation]="A scoped effort to learn the answers to a set of questions which may include prototyping"
    )

    merge "kind" kinds "${assigned_colors[kind]}"
}

merge-resolutions () {
    declare -A resolutions
    # The array is passed by name at the end of this function
    # shellcheck disable=SC2034
    resolutions=(
        [duplicate]="Another issue exists for this issue"
        [incomplete]="Insufficint information is available to address this issue"
        [invalid]="The issue is intended behavior or otherwise invalid"
        [will-not-fix]="This issue is valid, but will not be fixed"
    )

    merge "resolution" resolutions "${assigned_colors[resolution]}"
}

merge-status () {
    declare -A statuses
    # The array is passed by name at the end of this function
    # shellcheck disable=SC2034
    statuses=(
        [need-info]="Additional information is needed to make progress"
        [needs-attention]="The issue needs to be discussed by the team"
        [needs-triage]="The issue needs to be evaluated and metadata updated"
    )

    merge "status" statuses "${assigned_colors[status]}"
}

merge-all () {
    merge-oneoff
    merge-severity
    merge-areas
    merge-commitments
    merge-consumers
    merge-impacts
    merge-kinds
    merge-resolutions
    merge-status
}

merge-all


