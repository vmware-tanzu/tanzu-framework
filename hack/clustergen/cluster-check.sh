#!/bin/bash

# Copyright 2021 VMware, Inc. All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0

# Usage:
# clustergen-check.sh base-commit current-commit

if ! git diff-index --quiet HEAD --; then
	echo >&2 "Please commit or stash uncommitted changes before running again."
	exit 1
fi

BASE_DIR="$(dirname "$0")"

GOOS=$(go env GOOS)
GOARCH=$(go env GOARCH)
CLI_REPO=${CLI_REPO:-${PWD}/../..}
# Note: if GIT_BRANCH_PROVIDERS_BASE is not set, default to HEAD^, which will not
# produce the behavior unless the provider changes is consolidate into a single
# commit at HEAD
GIT_BRANCH_PROVIDERS_BASE=${1:-HEAD^}
GIT_BRANCH_PROVIDERS=${2:-HEAD}
LAST_BASE_BRANCH_COMMIT_FILE="tests/clustergen/testdata/base_branch_commit"

BASE_BRANCH_COMMIT=$(git rev-parse ${GIT_BRANCH_PROVIDERS_BASE})
LAST_BASE_BRANCH_COMMIT=$(cat ${LAST_BASE_BRANCH_COMMIT_FILE})

CURR_COMMIT=$(git rev-parse ${GIT_BRANCH_PROVIDERS})

echo "=== Diffing clustergen output between: ${BASE_BRANCH_COMMIT} (${GIT_BRANCH_PROVIDERS_BASE})"
echo "                                  and: ${CURR_COMMIT} (${GIT_BRANCH_PROVIDERS})"


# TODO: As we are using providers as library, clustergen should not rely on building binary
# each time with new providers, instead we should write some go testing tool  which runs
# clustergen tests via the tkg cli library interface
make generate-testcases

# TODO: Handle current commit that is not HEAD
rm -rf tests/clustergen/testdata/new || true
git log --pretty=oneline -5 | cat
make CLUSTERGEN_CC_OUTPUT_DIR=newcc CLUSTERGEN_OUTPUT_DIR=new GOOS=${GOOS} GOARCH=${GOARCH} CLI_REPO=${CLI_REPO} cluster-generation-tests

if [ "${LAST_BASE_BRANCH_COMMIT}" != "${BASE_BRANCH_COMMIT}" ]; then
  echo "Base branch differs, regenerating output for ${GIT_BRANCH_PROVIDERS_BASE}...."
  echo -n "${BASE_BRANCH_COMMIT}" > "${LAST_BASE_BRANCH_COMMIT_FILE}"
  rm -rf tests/clustergen/testdata/old || true
  git checkout -B clustergen_test_base ${GIT_BRANCH_PROVIDERS_BASE}
  git log --pretty=oneline -5 | cat
  make CLUSTERGEN_CC_OUTPUT_DIR=oldcc CLUSTERGEN_OUTPUT_DIR=old GOOS=${GOOS} GOARCH=${GOARCH} CLI_REPO=${CLI_REPO} cluster-generation-tests
  git checkout .
  git checkout -
else
  echo "Base branch commit for ${GIT_BRANCH_PROVIDERS_BASE} unchanged, skipping generation of base set...."
fi

pushd tests/clustergen/testdata
mkdir output || true
rm -f output/*
diff -r -U15 oldcc/cclass newcc/cclass > output/clustergen_cc.diff.txt
diff -r -U15 old new > output/clustergen.diff.txt
cat output/clustergen.diff.txt
docker run -i --rm -v $PWD:$PWD -w $PWD gcr.io/eminent-nation-87317/tkg-go-ci diff2html -i file -F output/clustergen.html -- output/clustergen.diff.txt
popd

exit 0
