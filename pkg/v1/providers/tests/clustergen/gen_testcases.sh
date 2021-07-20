#!/usr/bin/env bash
# Copyright 2020 The TKG Contributors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# Generates pair-wise combinatorial test cases. http://pairwise.org/

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
