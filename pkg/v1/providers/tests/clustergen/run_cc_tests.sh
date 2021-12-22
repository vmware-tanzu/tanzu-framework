#!/usr/bin/env bash

# Copyright 2021 VMware, Inc. All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0

SCRIPT=$(realpath "${BASH_SOURCE[0]}")
TESTROOT=$(dirname "$SCRIPT")
TKG=${TKG:-${TESTROOT}/../../bin/tkg-darwin-amd64}
CLUSTERCTL=${CLUSTERCTL:-~/cluster-api/bin/clusterctl}
TESTDATA=${TESTDATA:-testdata}
CASES=${CASES:-*.case}
BUILDER_IMAGE=gcr.io/eminent-nation-87317/tkg-go-ci:latest
YQ4=${YQ4:-yq}

TESTED_INFRAS=${TESTED_INFRAS:-"AWS AZURE VSPHERE"}
AZURE_CASES=${AZURE_CASES:-""}
AWS_CASES=${AWS_CASES:-""}
VSPHERE_CASES=${VSPHERE_CASES:-""}
MAX_CASES_PER_INFRA=${MAX_CASES_PER_INFRA:-20}

TKG_CONFIG_DIR="/tmp/test_tkg_config_dir"
rm -rf $TKG_CONFIG_DIR
mkdir -p $TKG_CONFIG_DIR

# shellcheck source=tests/clustergen/diffcluster/helpers.sh
. "${TESTROOT}"/diffcluster/helpers.sh

# initialize the CLI, set up providers, boms, etc
$TKG get mc --configdir ${TKG_CONFIG_DIR}

generate_cluster_configurations() {
  local outputdir=$1
  local infra=$2
  pushd "${TESTDATA}"

  VNAME=${infra}_CASES
  CASES=${!VNAME}

  if [ -z "$CASES" ]; then
    CASES=$(for i in `grep EXE: *.case | grep -i ${infra}:v | cut -d: -f1 | uniq | cut -d/ -f1 | head -${MAX_CASES_PER_INFRA}`; do echo -n "$i "; done; echo)
  fi

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
    echo $TKG --file /tmp/test_tkg_config --configdir ${TKG_CONFIG_DIR} --log_file /tmp/"$t".log config cluster "${cmdargs[@]}"
    $TKG --file /tmp/test_tkg_config --configdir ${TKG_CONFIG_DIR} --log_file /tmp/"$t".log config cluster "${cmdargs[@]}" 2>/tmp/err.txt 1>/tmp/expected.yaml
    RESULT=$?
    if [[ $RESULT -eq 0 ]]; then
      echo "$t":POS >>${outputdir}/failed.txt
      # normalize should not modify the yaml node trees, so doing so before saving to expected to
      # reduce the chance of generating diffs due to template formatting differences in the future.
      normalize /tmp/expected.yaml ${outputdir}/"$t".output
      ${CLUSTERCTL} alpha generate-normalized-topology -r -f /tmp/expected.yaml > ${outputdir}/"$t".norm.output
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

    if [[ $RESULT -eq 0 ]]; then
      # XXX fixup plan, hard code cluster class
      cat "$t" | perl -pe 's/--plan (\S+)/--plan $1cc/; s/_PLAN: (\S+)/_PLAN: $1cc/' > /tmp/test_tkg_config_cc
      infra=`echo "$infra" | awk '{ print tolower($1) }'`
      echo "CLUSTER_CLASS: tkg-${infra}-default" >> /tmp/test_tkg_config_cc
      read -r -a cmdargs < <(grep EXE: /tmp/test_tkg_config_cc | cut -d: -f2-)
      echo $TKG --file /tmp/test_tkg_config_cc --configdir ${TKG_CONFIG_DIR} --log_file /tmp/"$t"_cc.log config cluster "${cmdargs[@]}"
      $TKG --file /tmp/test_tkg_config_cc --configdir ${TKG_CONFIG_DIR} --log_file /tmp/"$t"_cc.log config cluster "${cmdargs[@]}" 2>/tmp/err_cc.txt 1>/tmp/expected_cc.yaml
      #normalize_cc /tmp/expected_cc.yaml ${outputdir}/"$t".cc.output
      cp /tmp/expected_cc.yaml ${outputdir}/"$t".cc.output
      ${CLUSTERCTL} alpha generate-normalized-topology -p -f ${outputdir}/"$t".cc.output > ${outputdir}/"$t".cc.gnt.yaml
      denoise_dryrun ${outputdir}/"$t".cc.gnt.yaml ${outputdir}/"$t".cc.norm.output

      generate_diff_summary ${outputdir} $t
    fi
  done

  compile_diff_stats ${outputdir} ${infra}
  rm ${outputdir}/*.diff_stats
  popd

  rm -rf $HOME/.tkg/bom/bom-clustergen-*
}

generate_diff_summary() {
   local outputdir=$1
   local t=$2

   ${YQ4} e '. | select(.kind != "Secret")' ${outputdir}/"$t".norm.output > ${outputdir}/"$t".norm.for_diff.yaml
   ${YQ4} e '. | select(.kind != "Secret")' ${outputdir}/"$t".cc.norm.output > ${outputdir}/"$t".cc.norm.for_diff.yaml

   wdiff -s ${outputdir}/"$t".norm.for_diff.yaml ${outputdir}/"$t".cc.norm.for_diff.yaml | tail -2 | head -1 > ${outputdir}/"$t".diff_stats
   cat ${outputdir}/"$t".diff_stats
}

compile_diff_stats() {
   local outputdir=$1
   local infra=$2
   local outfile=${outputdir}/${infra}_diff_summary.csv

   echo doing ${outfile}
   echo "TEST,SAME,DELETED,CHANGED" > ${outfile}
   for f in ${outputdir}/*.diff_stats; do
      echo -n "$f," >> ${outfile}
      cat $f | perl -pe 's/^.*\D(\d+)%.*\D(\d+)%.*\D(\d+)%.*$/$1,$2,$3/' >> ${outfile}
   done
   cat ${outfile}
}

export SUPPRESS_PROVIDERS_UPDATE=1
export CLUSTERCTL_SKIP_UNIQUE_NAMESPACE=true
export CLUSTERCTL_SKIP_FETCH_CC=true

outputdir=$1
mkdir -p ${TESTDATA}/${outputdir} || true
rm -rf ${TESTDATA}/${outputdir}/*
for infra in ${TESTED_INFRAS}; do
   pushd "${TKG_CONFIG_DIR}/providers/infrastructure-${infra}"
   for i in `find . -name "cluster-template-definition*cc.yaml"`; do
      perl -pi -e 's@^(.*- path: providers/infrastructure-.*/v.*/)(yttcc)@$1cconly\n$1$2@g' $i
   done
   popd

   generate_cluster_configurations $outputdir $infra
done

"${TESTROOT}"/summarize_diff_scores.py "${TESTDATA}/${outputdir}" "${TESTDATA}/${outputdir}/output.csv"
