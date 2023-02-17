#!/bin/bash

# Copyright 2021 VMware, Inc. All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0

# Turn off debugging (to prevent printing out AWS credentials)
set +x

AWS_ACCESS_KEY_ID=$1
AWS_SECRET_ACCESS_KEY=$2
AWS_SESSION_TOKEN=$3
AWS_REGION=$4
FILTER=$5

if [ -z "$AWS_ACCESS_KEY_ID" ]; then
  exit 0
fi

# This step makes sure that the filter applied is never empty
if [ -z "$FILTER" ]; then
  echo "filter for leftovers is not set"
  exit 1
fi

set -x
wget -nv https://github.com/genevieve/leftovers/releases/download/v0.62.0/leftovers-v0.62.0-linux-amd64
mv leftovers-v0.62.0-linux-amd64 /usr/local/bin/leftovers
chmod +x /usr/local/bin/leftovers

wget -nv https://github.com/kubernetes-sigs/kind/releases/download/v0.11.0/kind-linux-amd64
mv kind-linux-amd64 /usr/local/bin/kind
chmod +x /usr/local/bin/kind
set +x

# Delete any kind cluster that are left behind
echo "Deleting any kind clusters that are left behind"
set -x
kind get clusters | xargs -n 1 kind delete cluster --name
set +x

# Run dry-run to see all resources
echo leftovers -d -i aws --aws-access-key-id="***" --aws-secret-access-key="***" --aws-session-token="***" --filter="${FILTER}" --aws-region="${AWS_REGION}" -n
leftovers -d -i aws --aws-access-key-id="${AWS_ACCESS_KEY_ID}" --aws-secret-access-key="${AWS_SECRET_ACCESS_KEY}" --aws-session-token="${AWS_SESSION_TOKEN}" --filter="${FILTER}" --aws-region="${AWS_REGION}" -n

echo leftovers -i aws --aws-access-key-id="***" --aws-secret-access-key="***" --aws-session-token="***" --filter="${FILTER}" --aws-region="${AWS_REGION}" -n || true
leftovers -i aws --aws-access-key-id="${AWS_ACCESS_KEY_ID}" --aws-secret-access-key="${AWS_SECRET_ACCESS_KEY}" --aws-session-token="${AWS_SESSION_TOKEN}" --filter="${FILTER}" --aws-region="${AWS_REGION}" -n || true

# Run dry-run to see any leftover resources
echo leftovers -d -i aws --aws-access-key-id="***" --aws-secret-access-key="***" --aws-session-token="***" --filter="${FILTER}" --aws-region="${AWS_REGION}" -n
leftovers -d -i aws --aws-access-key-id="${AWS_ACCESS_KEY_ID}" --aws-secret-access-key="${AWS_SECRET_ACCESS_KEY}" --aws-session-token="${AWS_SESSION_TOKEN}" --filter="${FILTER}" --aws-region="${AWS_REGION}" -n

# Retry the deletion incase the previous attempt failed
echo leftovers -i aws --aws-access-key-id="***" --aws-secret-access-key="***" --aws-session-token="***" --filter="${FILTER}" --aws-region="${AWS_REGION}" -n
leftovers -i aws --aws-access-key-id="${AWS_ACCESS_KEY_ID}" --aws-secret-access-key="${AWS_SECRET_ACCESS_KEY}" --aws-session-token="${AWS_SESSION_TOKEN}" --filter="${FILTER}" --aws-region="${AWS_REGION}" -n
