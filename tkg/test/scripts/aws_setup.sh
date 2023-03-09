#!/bin/bash

# Copyright 2021 VMware, Inc. All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0

# Turn off debugging (to prevent printing out AWS credentials)
set +x

CG_TOKEN=$(curl -s --location --request POST 'https://api.console.cloudgate.vmware.com/authn/token' \
  --user "$USER_PASS" \
  --header 'Content-Type: application/json' \
  --data-raw '{"grant_type": "client_credentials"}' | jq -r '.access_token')

aws_creds=$(curl -s --location --request POST 'https://api.console.cloudgate.vmware.com/access/access' \
  --header "Authorization: Bearer ${CG_TOKEN}" \
  --header 'Content-Type: application/json' \
  --data-raw '{"ouId":"ou-kw69-lqh1erao","orgAccountId":"942999320260","masterAccountId":"116462199383","role":"PowerUser"}')

echo "AWS_ACCESS_KEY_ID="$(echo $aws_creds | jq -r '.credentials.accessKeyId') >> $GITHUB_ENV
echo "AWS_SECRET_ACCESS_KEY="$(echo $aws_creds | jq -r '.credentials.secretAccessKey') >> $GITHUB_ENV
echo "AWS_SESSION_TOKEN="$(echo $aws_creds | jq -r '.credentials.sessionToken') >> $GITHUB_ENV
echo "AWS_REGION=us-west-2" >> $GITHUB_ENV
