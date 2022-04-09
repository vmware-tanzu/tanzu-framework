#!/usr/bin/env bash

# Copyright 2021 VMware, Inc. All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0

set -euo pipefail

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
PROVIDERS_DIR=${SCRIPT_DIR}/../../../../pkg/v1/providers
PKG_MGMT_DIR=${SCRIPT_DIR}/../..
BINNAME="$( basename "${BASH_SOURCE[0]}" )"

infras="aws vsphere azure" # TODO: add docker

for infra in $infras; do
   ls ${PROVIDERS_DIR}/infrastructure-${infra}/v*/cconly/*
   rm ${PKG_MGMT_DIR}/tkg-clusterclass-${infra}/bundle/config/upstream/*
   cp ${PROVIDERS_DIR}/infrastructure-${infra}/v*/cconly/* ${PKG_MGMT_DIR}/tkg-clusterclass-${infra}/bundle/config/upstream/
done

