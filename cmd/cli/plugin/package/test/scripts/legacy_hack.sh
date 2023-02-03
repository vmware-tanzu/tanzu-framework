#!/bin/bash

# Copyright 2021 VMware, Inc. All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0

tanzu config set features.management-cluster.package-based-cc false
tanzu config set features.cluster.allow-legacy-cluster true
tanzu config get
