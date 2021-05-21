#!/bin/bash

# Usage:
# ci-validate-clustergen.sh branch_name target-branch-name


SCRIPT_DIR="$(cd -P "$(dirname "$0")" && pwd)"

GOOS=$(go env GOOS)
GOARCH=$(go env GOARCH)
CLI_REPO=${CLI_REPO:-${PWD}/../..}
GIT_BRANCH_PROVIDERS=$1
GIT_BRANCH_PROVIDERS_BASE=$2

cd pkg/v1/providers

# TODO: As we are using providers as library, clustergen should not rely on building binary
# each time with new providers, instead we should write some go testing tool  which runs
# clustergen tests via the tkg cli library interface
make generate-testcases

git log --pretty=oneline -5
make generate-bindata
CLI_REPO=${CLI_REPO} ${SCRIPT_DIR}/rebuild-cli.sh
CLUSTERGEN_OUTPUT_DIR=new make GOOS=${GOOS} GOARCH=${GOARCH} CLI_REPO=${CLI_REPO} cluster-generation-tests

echo git checkout -B old origin/${GIT_BRANCH_PROVIDERS_BASE}
git checkout -B old origin/${GIT_BRANCH_PROVIDERS_BASE}
git log --pretty=oneline -5
make generate-bindata
CLI_REPO=${CLI_REPO} ${SCRIPT_DIR}/rebuild-cli.sh ${PWD}
CLUSTERGEN_OUTPUT_DIR=old make GOOS=${GOOS} GOARCH=${GOARCH} CLI_REPO=${CLI_REPO} cluster-generation-tests
git checkout .
git checkout -

pushd tests/clustergen/testdata
diff -r -U15 old new > clustergen.diff.txt
cat clustergen.diff.txt
docker run -i --rm -v $PWD:$PWD -w $PWD gcr.io/eminent-nation-87317/diff2html diff2html -i file -F clustergen.html -- clustergen.diff.txt
popd

exit 0
