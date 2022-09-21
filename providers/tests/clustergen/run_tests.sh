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

SUPPORTED_INFRAS="AWS AZURE VSPHERE"
TESTED_INFRAS=${TESTED_INFRAS:-${SUPPORTED_INFRAS}}
MAX_CASES_PER_INFRA=${MAX_CASES_PER_INFRA:-20}

TKG_CONFIG_DIR="/tmp/test_tkg_config_dir"
rm -rf $TKG_CONFIG_DIR
mkdir -p $TKG_CONFIG_DIR

# shellcheck source=tests/clustergen/diffcluster/helpers.sh
. "${TESTROOT}"/diffcluster/helpers.sh

# initialize the CLI, set up providers, boms, etc
$TKG get mc --configdir ${TKG_CONFIG_DIR}

generate_cluster_configurations() {
  local infra=$1
  local outputdir=$2
  local outputdircc=$3
  pushd "${TESTDATA}"

  matched_cases=$(for i in `grep -H EXE: ${CASES} | grep -i ${infra}:v | cut -d: -f1 | uniq | cut -d/ -f1 | head -${MAX_CASES_PER_INFRA}`; do echo -n "$i "; done; echo)
  echo "MATCHED CASES = ($matched_cases)"

  $TKG get mc --configdir ${TKG_CONFIG_DIR}
  docker run -t --rm -v ${TKG_CONFIG_DIR}:${TKG_CONFIG_DIR} -v ${TESTROOT}:/clustergen -w /clustergen -e TKG_CONFIG_DIR=${TKG_CONFIG_DIR} ${BUILDER_IMAGE} /bin/bash -c "./gen_duplicate_bom_azure.py $TKG_CONFIG_DIR"
  RESULT=$?
  if [[ ! $RESULT -eq 0 ]]; then
    exit 1
  fi

  echo "# failed cases" >${outputdir}/failed.txt
  echo "Running $TKG config cluster ..."
  for t in $matched_cases; do
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
      if [ ! -z "$outputdircc" ]; then
        ${CLUSTERCTL} alpha generate-normalized-topology -r -f /tmp/expected.yaml > ${outputdircc}/"$t".norm.output
      fi
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

    if [ ! -z "$outputdircc" ]; then
      if [[ $RESULT -eq 0 ]]; then
        # XXX fixup plan, hard code cluster class
        cat "$t" | perl -pe 's/--plan (\S+)/--plan $1cc/; s/_PLAN: (\S+)/_PLAN: $1cc/' > /tmp/test_tkg_config_cc
        echo "CLUSTER_CLASS: tkg-${infra}-default" >> /tmp/test_tkg_config_cc
        read -r -a cmdargs < <(grep EXE: /tmp/test_tkg_config_cc | cut -d: -f2-)

        if [[ "${infra}" == "aws" ]]; then
          cat <<- EOF >> /tmp/test_tkg_config_cc
TKR_DATA: |-
  v1.23.5+vmware.1:
    kubernetesSpec:
      version: v1.23.5+vmware.1
      imageRepository: projects-stg.registry.vmware.com
      etcd:
        imageTag: v1.0.0-test
      coredns:
        imageTag: v1.1.0-test
      kube-vip:
        imageTag: v2.0.0-test
    labels:
      os-name: ubuntu
      os-type: linux
      os-arch: amd64
    osImageRef:
      id: test-ami-id
      region: test-region
EOF
        fi
        if [[ "${infra}" == "vsphere" ]]; then
          cat <<- EOF >> /tmp/test_tkg_config_cc
TKR_DATA: |-
  v1.21.2:
    kubernetesSpec:
      version: v1.21.2
      imageRepository: projects-stg.registry.vmware.com
      etcd:
        imageTag: v1.0.0-test
      coredns:
        imageTag: v1.1.0-test
      kube-vip:
        imageTag: v2.0.0-test
    labels:
      os-name: ubuntu
      os-type: linux
      os-arch: amd64
EOF
        fi
        if [[ "${infra}" == "azure" ]]; then
          if grep -q "AZURE_IMAGE_GALLERY" /tmp/test_tkg_config_cc; then
            cat <<- EOF >> /tmp/test_tkg_config_cc
TKR_DATA: |-
  v1.23.5+vmware.1:
    kubernetesSpec:
      version: v1.23.5+vmware.1
      imageRepository: projects-stg.registry.vmware.com
      etcd:
        imageTag: v1.0.0-test
      coredns:
        imageTag: v1.1.0-test
    labels:
      os-name: ubuntu
      os-type: linux
      os-arch: amd64
    osImageRef:
      version: test-version
      gallery: test-gallery
      name: test-name
      resourceGroup: test-resource-group
      subscriptionID: test-subscription-id
EOF
          else
            cat <<- EOF >> /tmp/test_tkg_config_cc
TKR_DATA: |-
  v1.23.5+vmware.1:
    kubernetesSpec:
      version: v1.23.5+vmware.1
      imageRepository: projects-stg.registry.vmware.com
      etcd:
        imageTag: v1.0.0-test
      coredns:
        imageTag: v1.1.0-test
    labels:
      os-name: ubuntu
      os-type: linux
      os-arch: amd64
    osImageRef:
      sku: test-sku
      publisher: test-publisher
      offer: test-offer
      version: test-version
      thirdPartyImage: test-third-party-image
EOF
          fi
        fi

        echo $TKG --file /tmp/test_tkg_config_cc --configdir ${TKG_CONFIG_DIR} --log_file /tmp/"$t"_cc.log config cluster "${cmdargs[@]}"
        $TKG --file /tmp/test_tkg_config_cc --configdir ${TKG_CONFIG_DIR} --log_file /tmp/"$t"_cc.log config cluster "${cmdargs[@]}" 2>/tmp/err_cc.txt 1>/tmp/expected_cc.yaml
        #normalize_cc /tmp/expected_cc.yaml ${outputdir}/"$t".cc.output
        cp /tmp/expected_cc.yaml ${outputdir}/"$t".cc.output
        ${CLUSTERCTL} alpha generate-normalized-topology -p -f ${outputdir}/"$t".cc.output > /tmp/"$t".cc.gnt.yaml
        denoise_dryrun /tmp/"$t".cc.gnt.yaml ${outputdircc}/"$t".cc.norm.output

        echo generate_diff_summary "${outputdircc}","$t"
        generate_diff_summary ${outputdircc} $t
      fi
    fi
  done
  popd
  rm -rf $HOME/.tkg/bom/bom-clustergen-*
}

generate_diff_summary() {
   local dir=$1
   local t=$2

   ${YQ4} e '. | select(.kind != "Secret")' ${dir}/"$t".norm.output > ${dir}/legacy/"$t".norm.yaml
   ${YQ4} e '. | select(.kind != "Secret")' ${dir}/"$t".cc.norm.output > ${dir}/cclass/"$t".norm.yaml

   diff -U10 ${dir}/legacy/"$t".norm.yaml ${dir}/cclass/"$t".norm.yaml > ${dir}/"$t".u10.diff
   wdiff -s ${dir}/legacy/"$t".norm.yaml ${dir}/cclass/"$t".norm.yaml | tail -2 | head -1 > ${dir}/"$t".diff_stats
   cat ${dir}/"$t".diff_stats
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
    #echo "$untracked"
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

export SUPPRESS_PROVIDERS_UPDATE=1
export CLUSTERCTL_SKIP_UNIQUE_NAMESPACE=true
export CLUSTERCTL_SKIP_FETCH_CC=true
export UNRANDOMIZE=true

outputdir=$1
mkdir -p ${TESTDATA}/${outputdir} || true
rm -rf ${TESTDATA}/${outputdir}/*

outputdircc=
if [ ! -z "$2" ]; then
outputdircc=$2
mkdir -p ${TESTDATA}/${outputdircc} || true
rm -rf ${TESTDATA}/${outputdircc}/*
mkdir -p ${TESTDATA}/${outputdircc}/cclass
mkdir -p ${TESTDATA}/${outputdircc}/legacy
fi

for infra in ${TESTED_INFRAS}; do
  infra_lc=`echo "$infra" | awk '{ print tolower($1) }'`
  generate_cluster_configurations $infra_lc $outputdir $outputdircc
done

set -e
check_generated
