#!/usr/bin/env bash

# Copyright 2021 VMware, Inc. All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0

SCRIPT=$(realpath "${BASH_SOURCE[0]}")
TESTROOT=$(dirname "$SCRIPT")
TKG=${TKG:-${TESTROOT}/../../bin/tkg-darwin-amd64}
TESTDATA=${TESTDATA:-testdata}
CASES=${CASES:-*.case}
BUILDER_IMAGE=gcr.io/eminent-nation-87317/tkg-go-ci:latest

TKG_CONFIG_DIR="/tmp/test_tkg_config_dir"
rm -rf $TKG_CONFIG_DIR
mkdir -p $TKG_CONFIG_DIR

# shellcheck source=tests/clustergen/diffcluster/helpers.sh
. "${TESTROOT}"/diffcluster/helpers.sh

generate_cluster_configurations() {
  outputdir=$1
  cd "${TESTDATA}"
  mkdir -p ${outputdir} || true
  rm -rf ${outputdir}/*

  $TKG get mc --configdir ${TKG_CONFIG_DIR}
  docker run -t --rm -v ${TKG_CONFIG_DIR}:${TKG_CONFIG_DIR} -v ${TESTROOT}:/clustergen -w /clustergen -e TKG_CONFIG_DIR=${TKG_CONFIG_DIR} ${BUILDER_IMAGE} /bin/bash -c "./gen_duplicate_bom_azure.py $TKG_CONFIG_DIR"
  RESULT=$?
  if [[ ! $RESULT -eq 0 ]]; then
    exit 1
  fi

  echo "# failed cases" >${outputdir}/failed.txt
  echo "Running $TKG config cluster ..."
  for t in $CASES; do
    cmdargs=()
    read -r -a cmdargs < <(grep EXE: "$t" | cut -d: -f2-)
    cp "$t" /tmp/test_tkg_config
    $TKG --file /tmp/test_tkg_config --configdir ${TKG_CONFIG_DIR} --log_file /tmp/"$t".log config cluster "${cmdargs[@]}" 2>/tmp/err.txt 1>/tmp/expected.yaml
    #shellcheck disable=SC2181
    if [ $? -eq 0 ]; then
      echo "$t":POS >>${outputdir}/failed.txt
      # normalize should not modify the yaml node trees, so doing so before saving to expected to
      # reduce the chance of generating diffs due to template formatting differences in the future.
      normalize /tmp/expected.yaml ${outputdir}/"$t".output
      echo -n "$t (POS) : "
    else
      # failure to generate a working configuration can be due to a variety of reasons. They are
      # represented as a NEGative test case. The output of the failed command is captured and is part
      # of the compliance dataset.
      cp "$t" /tmp/test_tkg_config
      $TKG --file /tmp/test_tkg_config --configdir ${TKG_CONFIG_DIR} --log_file /tmp/"$t".log config cluster "${cmdargs[@]}" &>${outputdir}/"$t".output
      echo "$t":NEG >>${outputdir}/failed.txt
      echo -n "$t (NEG) : "
    fi
    echo "${cmdargs[@]}"
  done
  rm -rf $HOME/.tkg/bom/bom-clustergen-*
}

diffcluster() {
  diff "$1" "$2"
  # TODO : update to use more yaml-aware diff
  # kapp tools diff -c --line-numbers=false --summary=false --file $1 --file2 $2
}

check_generated() {
  # flag new files generated
  untracked=$(git ls-files -o --directory --exclude-standard .)
  num_untracked=$(echo -n "${untracked}" | wc -l)
  if [ "$num_untracked" -ne 0 ]; then
    echo "New entries found:"
    echo "$untracked"
    echo ""
    echo "The above are new entries from the last test. If these changes are expected, commit them and retry."
    exit 1
  fi

  deleted=$(git status -s | grep ' D ' || true)
  num_deleted=$(echo -n "${deleted}" | wc -l)
  if [ "$num_deleted" -ne 0 ]; then
    echo "Deleted entries found:"
    echo "$deleted"
    echo ""
    echo "The above entries have been deleted. If the changes are expected, commit the removal (e.g.  git add -u) and retry."
    exit 1
  fi

  relpath=$(git rev-parse --show-prefix)
  modified=$(git status -s expected/*.yaml | grep ' M ' | cut -c4-)
  for m in $modified; do
    git show HEAD:"${relpath}""$m" >/tmp/orig.yaml
    normalize /tmp/orig.yaml /tmp/orig.normalized.yaml
    diffcluster /tmp/orig.normalized.yaml "$m"
  done
}

generate_cluster_configurations $1
set -e
check_generated
