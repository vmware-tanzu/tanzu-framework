/*
Copyright 2020 The TKG Contributors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Package providersupgradeclient provides wrapper for clusterctl upgrade functionalities
package providersupgradeclient

import (
	clusterctl "sigs.k8s.io/cluster-api/cmd/clusterctl/client"
)

// Client provides interface which is a proxy for clusterctl client's interface ApplyUpgrade method, created to help unit tests
// TODO: This can be extended as a proxy for all operations of clusterctl client
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
