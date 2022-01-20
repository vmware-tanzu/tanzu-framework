// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"fmt"
	"time"

	"github.com/pkg/errors"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"

	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/clusterclient"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/region"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/utils"
)

// GetClusterPinnipedInfoOptions contains options supported by GetClusterPinnipedInfo
type GetClusterPinnipedInfoOptions struct {
	ClusterName         string
	Namespace           string
	IsManagementCluster bool
}

// ClusterPinnipedInfo defines the fields of cluster pinniped info
type ClusterPinnipedInfo struct {
	ClusterName  string
	ClusterInfo  *clientcmdapi.Cluster
	PinnipedInfo *utils.PinnipedConfigMapInfo
}

// GetClusterPinnipedInfo gets pinniped information from cluster
func (c *TkgClient) GetClusterPinnipedInfo(options GetClusterPinnipedInfoOptions) (*ClusterPinnipedInfo, error) {
	clusterclientOptions := clusterclient.Options{
		GetClientInterval: 1 * time.Second,
		GetClientTimeout:  3 * time.Second,
	}

	curRegion, err := c.GetCurrentRegionContext()
	if err != nil {
		return nil, errors.Wrap(err, "unable to get current management cluster configuration")
	}

	regionalClusterClient, err := clusterclient.NewClient(curRegion.SourceFilePath, curRegion.ContextName, clusterclientOptions)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get cluster client while getting cluster pinniped info of tkg clusters")
	}

	isPacific, err := regionalClusterClient.IsPacificRegionalCluster()
	if err != nil {
		return nil, errors.Wrap(err, "error determining 'Tanzu Kubernetes Cluster service for vSphere' management cluster")
	}
	if isPacific {
		return nil, errors.New("getting pinniped information not supported for 'Tanzu Kubernetes Cluster service for vSphere' cluster")
	}

	if options.IsManagementCluster {
		return c.GetMCClusterPinnipedInfo(regionalClusterClient, curRegion, options)
	}

	return c.GetWCClusterPinnipedInfo(regionalClusterClient, curRegion, options)
}

// GetWCClusterPinnipedInfo gets pinniped information for workload cluster
func (c *TkgClient) GetWCClusterPinnipedInfo(regionalClusterClient clusterclient.Client,
	curRegion region.RegionContext, options GetClusterPinnipedInfoOptions) (*ClusterPinnipedInfo, error) {

	wcClusterInfo, err := getClusterInfo(regionalClusterClient, options.ClusterName, options.Namespace)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get workload cluster information")
	}
	// for workload cluster pinniped-info should be available on management cluster
	mcClusterInfo, err := getClusterInfo(regionalClusterClient, curRegion.ClusterName, TKGsystemNamespace)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get management cluster information")
	}
	managementClusterPinnipedInfo, err := utils.GetPinnipedInfoFromCluster(mcClusterInfo)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get pinniped-info from management cluster")
	}
	if managementClusterPinnipedInfo == nil {
		return nil, errors.New("failed to get pinniped-info from management cluster")
	}

	workloadClusterPinnipedInfo, err := utils.GetPinnipedInfoFromCluster(wcClusterInfo)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get pinniped-info from workload cluster")
	}

	pinnipedInfo := managementClusterPinnipedInfo
	if workloadClusterPinnipedInfo != nil {
		// Get ConciergeIsClusterScoped from workload cluster in case it is different from the management cluster
		pinnipedInfo.Data.ConciergeIsClusterScoped = workloadClusterPinnipedInfo.Data.ConciergeIsClusterScoped
	} else {
		// If workloadClusterPinnipedInfo is nil, assume it is an older TKG cluster and set ConciergeIsClusterScoped to defaults
		pinnipedInfo.Data.ConciergeIsClusterScoped = false
	}

	return &ClusterPinnipedInfo{
		ClusterName:  options.ClusterName,
		ClusterInfo:  wcClusterInfo,
		PinnipedInfo: pinnipedInfo,
	}, nil
}

// GetMCClusterPinnipedInfo get pinniped information for management cluster
func (c *TkgClient) GetMCClusterPinnipedInfo(regionalClusterClient clusterclient.Client,
	curRegion region.RegionContext, options GetClusterPinnipedInfoOptions) (*ClusterPinnipedInfo, error) {
	// it is expected that user would call get cluster pinnedInfo of the same management cluster
	clusterInfo, err := getClusterInfo(regionalClusterClient, options.ClusterName, options.Namespace)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get cluster information")
	}
	pinnipedInfo, err := utils.GetPinnipedInfoFromCluster(clusterInfo)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get pinniped-info from cluster")
	}

	if pinnipedInfo == nil {
		return nil, errors.New("failed to get pinniped-info from cluster")
	}

	return &ClusterPinnipedInfo{
		ClusterName:  options.ClusterName,
		ClusterInfo:  clusterInfo,
		PinnipedInfo: pinnipedInfo,
	}, nil
}

func getClusterInfo(
	regionalClusterClient clusterclient.Client,
	clusterName, clusterNamespace string,
) (*clientcmdapi.Cluster, error) {

	kubeconfigData, err := regionalClusterClient.GetKubeConfigForCluster(clusterName, clusterNamespace, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get kubeconfig for cluster %s/%s: %w", clusterNamespace, clusterName, err)
	}

	config, err := clientcmd.Load(kubeconfigData)
	if err != nil {
		return nil, errors.Wrap(err, "failed to load the kubeconfig")
	}

	if len(config.Clusters) == 0 {
		return nil, errors.New("failed to get cluster information")
	}

	// since it is a map with one cluster object, get the first entry
	var cluster *clientcmdapi.Cluster
	for _, cluster = range config.Clusters {
		break
	}

	return cluster, nil
}
