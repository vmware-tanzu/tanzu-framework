// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package fakegen provides a fake CtrClient  for testing
package fakegen

import (
	crtclient "sigs.k8s.io/controller-runtime/pkg/client"
)

//go:generate counterfeiter -o ../pkg/fakeclusterclient/crtclusterclient.go --fake-name CRTClusterClient . CrtClient

// CrtClient clientset interface
type CrtClient interface {
	crtclient.Client
}
