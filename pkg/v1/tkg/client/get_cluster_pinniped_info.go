// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"fmt"
	"net"
	"strconv"
	"time"

	"github.com/pkg/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	capi "sigs.k8s.io/cluster-api/api/v1beta1"

	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/clusterclient"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/log"
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
		return c.GetMCClusterPinnipedInfo(regionalClusterClient, curRegion)
	}

	return c.GetWCClusterPinnipedInfo(regionalClusterClient, curRegion, options)
}

// GetWCClusterPinnipedInfo gets pinniped information for workload cluster
func (c *TkgClient) GetWCClusterPinnipedInfo(regionalClusterClient clusterclient.Client,
	curRegion region.RegionContext, options GetClusterPinnipedInfoOptions) (*ClusterPinnipedInfo, error) {

	clusterAPIServerURL, err := c.getClusterAPIServerURL(regionalClusterClient, options.ClusterName, options.Namespace)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get cluster apiserver url from cluster objects")
	}
	clusterInfo, err := utils.GetClusterInfoFromCluster(clusterAPIServerURL)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get cluster-info from cluster")
	}
	// for workload cluster pinniped-info should be available on management cluster
	mcServerURL, err := utils.GetClusterServerFromKubeconfigAndContext(curRegion.SourceFilePath, curRegion.ContextName)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get cluster apiserver url")
	}
	mcClusterInfo, err := utils.GetClusterInfoFromCluster(mcServerURL)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get management cluster-info")
	}
	pinnipedInfo, err := utils.GetPinnipedInfoFromCluster(mcClusterInfo)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get pinniped-info from management cluster")
	}

	return &ClusterPinnipedInfo{
		ClusterName:  options.ClusterName,
		ClusterInfo:  clusterInfo,
		PinnipedInfo: pinnipedInfo,
	}, nil
}

// GetMCClusterPinnipedInfo get pinniped information for management cluster
func (c *TkgClient) GetMCClusterPinnipedInfo(regionalClusterClient clusterclient.Client,
	curRegion region.RegionContext) (*ClusterPinnipedInfo, error) {
	// it is expected that user would call get cluster pinnedInfo of the same management cluster
	clusterName, err := regionalClusterClient.GetCurrentClusterName(curRegion.ContextName)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get cluster name")
	}
	// management cluster namespace is opinionated and is "tkg-system"
	namespace := TKGsystemNamespace

	clusterAPIServerURL, err := c.getClusterAPIServerURL(regionalClusterClient, clusterName, namespace)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get cluster apiserver url from cluster objects")
	}

	clusterInfo, err := utils.GetClusterInfoFromCluster(clusterAPIServerURL)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get cluster-info from cluster")
	}
	pinnipedInfo, err := utils.GetPinnipedInfoFromCluster(clusterInfo)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get pinniped-info from cluster")
	}

	return &ClusterPinnipedInfo{
		ClusterName:  clusterName,
		ClusterInfo:  clusterInfo,
		PinnipedInfo: pinnipedInfo,
	}, nil
}

func (c *TkgClient) getClusterAPIServerURL(regionalClusterClient clusterclient.Client, clusterName, namespace string) (string, error) {
	var apiServerHost string
	var apiServerPort string
	var cluster capi.Cluster
	err := regionalClusterClient.GetResource(&cluster, clusterName, namespace, nil, nil)
	if err != nil {
		if apierrors.IsNotFound(err) {
			errMsg := fmt.Sprintf("cluster '%s' is not present in namespace '%s' ", clusterName, namespace)
			log.V(4).Info(errMsg)
			return "", errors.New(errMsg)
		}
		return "", errors.Wrap(err, "failed to get 'cluster' resource")
	}

	if cluster.Spec.ControlPlaneEndpoint.Host == "" {
		return "", errors.New("controlplane endpoint 'host' was not set in 'cluster' resource")
	}

	if cluster.Spec.ControlPlaneEndpoint.Port == 0 {
		return "", errors.New("controlplane endpoint 'port' was not set in 'cluster' resource")
	}

	apiServerHost = cluster.Spec.ControlPlaneEndpoint.Host
	apiServerPort = strconv.Itoa(int(cluster.Spec.ControlPlaneEndpoint.Port))
	return net.JoinHostPort(apiServerHost, apiServerPort), nil
}
