// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package providersupgradeclient provides wrapper for clusterctl upgrade functionalities
package providersupgradeclient

import (
	clusterctl "sigs.k8s.io/cluster-api/cmd/clusterctl/client"
)

// Client provides interface which is a proxy for clusterctl client's interface ApplyUpgrade method, created to help unit tests
// TODO: This can be extended as a proxy for all operations of clusterctl client
//
//go:generate counterfeiter -o ../fakes/providersupgradeclient.go --fake-name ProvidersUpgradeClient . Client
type Client interface {
	ApplyUpgrade(*clusterctl.ApplyUpgradeOptions) error
}

// ProvidersUpgradeClient implements providers upgrade functionality
type ProvidersUpgradeClient struct {
	clusterctlClient clusterctl.Client
}

// New returns the ProviderUpgradeClient
func New(cctlClient clusterctl.Client) Client {
	return &ProvidersUpgradeClient{
		clusterctlClient: cctlClient,
	}
}

// ApplyUpgrade acts as proxy for ClusterCtl ApplyUpgrade method
func (puc *ProvidersUpgradeClient) ApplyUpgrade(pUpgradeApplyOptions *clusterctl.ApplyUpgradeOptions) error {
	return puc.clusterctlClient.ApplyUpgrade(*pUpgradeApplyOptions)
}
