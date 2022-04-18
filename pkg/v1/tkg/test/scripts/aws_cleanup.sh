#!/bin/bash

# Copyright 2021 VMware, Inc. All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0

set -x

AWS_ACCESS_KEY_ID=$1
AWS_SECRET_ACCESS_KEY=$2
AWS_SESSION_TOKEN=$3
AWS_REGION=$4
FILTER=$5

wget https://github.com/genevieve/leftovers/releases/download/v0.62.0/leftovers-v0.62.0-linux-amd64
mv leftovers-v0.62.0-linux-amd64 /usr/local/bin/leftovers
chmod +x /usr/local/bin/leftovers

leftovers -i aws --aws-access-key-id="${AWS_ACCESS_KEY_ID}" --aws-secret-access-key="${AWS_SECRET_ACCESS_KEY}" --aws-session-token="${AWS_SESSION_TOKEN}" --filter="${FILTER}" --aws-region="${AWS_REGION}" -n