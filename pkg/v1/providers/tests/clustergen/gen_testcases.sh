#!/usr/bin/env bash

# Copyright 2021 VMware, Inc. All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0

# alternate location of the pict (https://github.com/microsoft/pict) binary if not in PATH
PICT=${PICT:=pict}

SCRIPT=$(realpath "${BASH_SOURCE[0]}")
TESTROOT=$(dirname "$SCRIPT")
PARAMS_CSV=${TESTROOT}/gencluster_params.csv
TESTDATA=${TESTDATA:-testdata}
RUNPICT=${RUNPICT:-true}

#shellcheck disable=SC2164,SC2086,SC2035,SC2012
infras=$(cd ${TESTROOT}/param_models; ls -1 *.model | cut -d. -f1 | cut -d_ -f2-)

rm -rf "${TESTDATA:?}"/*.case

for infra in ${infras}; do
  PARAMS_CSV=${TESTROOT}/gencluster_params_${infra}.csv

  if [ "${RUNPICT}" = true ]; then
    echo -n >"${PARAMS_CSV}"
    model=${TESTROOT}/param_models/cluster_"$infra".model
    echo processing "$model" ...
    ${PICT} "$model" /s
    echo ${PICT} "$model" to /tmp/pict.gen."$infra"
    ${PICT} "$model" >/tmp/pict.gen."$infra"
    echo "${PARAMS_CSV}"
    # pict's output is padded to tabular form. Cleanse it into csv with no spaces between values
    perl -pe 's/"\s+"/","/g; s/([-A-Z0-9_])\s+([-A-Z0-9_])/$1,$2/g; s/@@/,/g' /tmp/pict.gen."${infra}" >>"${PARAMS_CSV}"
  else
    echo "(skipped) parameter generation with pict"
  fi

  echo "Generating test cases for cluster creation from ${PARAMS_CSV} ..."
  "${TESTROOT}"/gencluster_params.py "${PARAMS_CSV}" "${TESTDATA}"
done
