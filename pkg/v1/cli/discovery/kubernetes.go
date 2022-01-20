// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package discovery

import (
	"time"

	"github.com/aunum/log"
	"github.com/pkg/errors"

	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/cli/common"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/cli/plugin"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/clusterclient"
)

// KubernetesDiscovery is an artifact discovery utilizing CLIPlugin API in kubernetes cluster
type KubernetesDiscovery struct {
	name           string
	kubeconfigPath string
	kubecontext    string
}

// NewKubernetesDiscovery returns a new kubernetes repository
func NewKubernetesDiscovery(name, kubeconfigPath, kubecontext string) Discovery {
	return &KubernetesDiscovery{
		name:           name,
		kubeconfigPath: kubeconfigPath,
		kubecontext:    kubecontext,
	}
}

// List available plugins.
func (k *KubernetesDiscovery) List() ([]plugin.Discovered, error) {
	return k.Manifest()
}

// Describe a plugin.
func (k *KubernetesDiscovery) Describe(name string) (p plugin.Discovered, err error) {
	plugins, err := k.Manifest()
	if err != nil {
		return
	}

	for i := range plugins {
		if plugins[i].Name == name {
			p = plugins[i]
			return
		}
	}
	err = errors.Errorf("cannot find plugin with name '%v'", name)
	return
}

// Name of the repository.
func (k *KubernetesDiscovery) Name() string {
	return k.name
}

// Manifest returns the manifest for a kubernetes repository.
func (k *KubernetesDiscovery) Manifest() ([]plugin.Discovered, error) {
	// Create cluster client
	clusterClientOptions := clusterclient.Options{GetClientInterval: 2 * time.Second, GetClientTimeout: 5 * time.Second}
	clusterClient, err := clusterclient.NewClient(k.kubeconfigPath, k.kubecontext, clusterClientOptions)
	if err != nil {
		return nil, err
	}

	return k.GetDiscoveredPlugins(clusterClient)
}

// GetDiscoveredPlugins returns the list of discovered plugin from a kubernetes cluster
func (k *KubernetesDiscovery) GetDiscoveredPlugins(clusterClient clusterclient.Client) ([]plugin.Discovered, error) {
	plugins := make([]plugin.Discovered, 0)

	exists, err := clusterClient.VerifyCLIPluginCRD()
	if !exists || err != nil {
		logMsg := "Skipping context-aware plugin discovery because CLIPlugin CRD not present on the logged in cluster. "
		if err != nil {
			logMsg += err.Error()
		}
		log.Debug(logMsg)
		return nil, nil
	}

	// get all cliplugins resources available on the cluster
	cliplugins, err := clusterClient.ListCLIPluginResources()
	if err != nil {
		return nil, err
	}

	// Convert all CLIPlugin resources to Discovered object
	for i := range cliplugins {
		dp, err := DiscoveredFromK8sV1alpha1(&cliplugins[i])
		if err != nil {
			return nil, err
		}
		dp.Source = k.name
		dp.DiscoveryType = k.Type()
		plugins = append(plugins, dp)
	}

	return plugins, nil
}

// Type of the repository.
func (k *KubernetesDiscovery) Type() string {
	return common.DiscoveryTypeKubernetes
}
