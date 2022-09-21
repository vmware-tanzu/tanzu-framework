#!/bin/bash

# Copyright 2021 VMware, Inc. All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0

# Usage:
# ci-validate-clustergen.sh branch_name target-branch-name

SCRIPT_DIR="$(cd -P "$(dirname "$0")" && pwd)"

GOOS=$(go env GOOS)
GOARCH=$(go env GOARCH)
CLI_REPO=${CLI_REPO:-${PWD}/../..}
GIT_BRANCH_PROVIDERS=$1
GIT_BRANCH_PROVIDERS_BASE=$2

cd providers

# TODO: As we are using providers as library, clustergen should not rely on building binary
# each time with new providers, instead we should write some go testing tool  which runs
# clustergen tests via the tkg cli library interface
make generate-testcases

git log --pretty=oneline -5
make generate-bindata
make CLUSTERGEN_CC_OUTPUT_DIR=newcc CLUSTERGEN_OUTPUT_DIR=new GOOS=${GOOS} GOARCH=${GOARCH} CLI_REPO=${CLI_REPO} cluster-generation-tests
git checkout .

echo git checkout -B old origin/${GIT_BRANCH_PROVIDERS_BASE}
git checkout -B old origin/${GIT_BRANCH_PROVIDERS_BASE}
git log --pretty=oneline -5
make generate-bindata
make CLUSTERGEN_CC_OUTPUT_DIR=oldcc CLUSTERGEN_OUTPUT_DIR=old GOOS=${GOOS} GOARCH=${GOARCH} CLI_REPO=${CLI_REPO} cluster-generation-tests
git checkout .
git checkout -

pushd tests/clustergen/testdata

diff -r -U15 old new > clustergen.diff.txt
cat clustergen.diff.txt
echo "<html><body>no diff</body></html>" > clustergen.html
docker run -i --rm -v $PWD:$PWD -w $PWD gcr.io/eminent-nation-87317/diff2html diff2html -i file -F clustergen.html -- clustergen.diff.txt

diff -r -U10 oldcc/cclass  newcc/cclass > clustergen_cc.diff.txt

pushd newcc
diff -r -U10 legacy cclass > ../clustergen_noncc_vs_cc.diff.txt
popd

popd

exit 0
