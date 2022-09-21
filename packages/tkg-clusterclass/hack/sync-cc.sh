#!/usr/bin/env bash

# Copyright 2021 VMware, Inc. All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0

set -euo pipefail

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
PROVIDERS_DIR=${SCRIPT_DIR}/../../../providers
PKGS_DIR=${SCRIPT_DIR}/../..
BINNAME="$( basename "${BASH_SOURCE[0]}" )"

infras="aws vsphere azure docker"

for infra in $infras; do
   ls ${PROVIDERS_DIR}/infrastructure-${infra}/v*/cconly/*
   rm ${PKGS_DIR}/tkg-clusterclass-${infra}/bundle/config/upstream/*
   cp ${PROVIDERS_DIR}/infrastructure-${infra}/v*/cconly/* ${PKGS_DIR}/tkg-clusterclass-${infra}/bundle/config/upstream/
done

