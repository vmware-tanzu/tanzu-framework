#!/bin/bash

# Copyright 2023 VMware, Inc. All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0

# Use clusterclass based management cluster until we need legacy mc again
#tanzu config set features.management-cluster.package-based-cc false
tanzu config set features.cluster.allow-legacy-cluster true
tanzu config get
