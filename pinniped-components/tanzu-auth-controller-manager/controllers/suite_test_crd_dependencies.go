//go:build tests
// +build tests

// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Ensure CRDs used for tests are present

package controllers

import (
	_ "sigs.k8s.io/cluster-api/api/v1beta1"
)
