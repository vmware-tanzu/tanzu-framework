#!/bin/bash

# Copyright 2021 VMware, Inc. All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0

tanzu config set features.cluster.auto-apply-generated-clusterclass-based-configuration true
tanzu config get
