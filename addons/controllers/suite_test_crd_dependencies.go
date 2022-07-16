//go:build tests
// +build tests

// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Ensure CRDs used for tests are present

package controllers

import (
	_ "sigs.k8s.io/cluster-api-provider-vsphere/apis/v1beta1"
	_ "sigs.k8s.io/cluster-api/api/v1beta1"

	_ "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apis/packaging/v1alpha1"
)
