// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"fmt"
	"strings"
	"time"

	"github.com/pkg/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	capvv1alpha3 "sigs.k8s.io/cluster-api-provider-vsphere/api/v1alpha3"
	capi "sigs.k8s.io/cluster-api/api/v1alpha3"

	capav1alpha3 "sigs.k8s.io/cluster-api-provider-aws/api/v1alpha3"
	capzv1alpha3 "sigs.k8s.io/cluster-api-provider-azure/api/v1alpha3"

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
	var apiServerPort int32
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
	clusterKind := cluster.Spec.InfrastructureRef.Kind
	switch strings.ToLower(clusterKind) {
	case "vspherecluster":
		var vSphereCluster capvv1alpha3.VSphereCluster
		err := regionalClusterClient.GetResource(&vSphereCluster, clusterName, namespace, nil, nil)
		if err != nil {
			if apierrors.IsNotFound(err) {
				errMsg := fmt.Sprintf("VSphereCluster object '%s' is not present in namespace '%s' ", clusterName, namespace)
				log.V(4).Info(errMsg)
				return "", errors.New(errMsg)
			}
			return "", errors.Wrap(err, "failed to get 'VSphereCluster' resource")
		}

		apiServerHost = vSphereCluster.Spec.ControlPlaneEndpoint.Host
		apiServerPort = vSphereCluster.Spec.ControlPlaneEndpoint.Port
	case "awscluster":
		var awsCluster capav1alpha3.AWSCluster
		err := regionalClusterClient.GetResource(&awsCluster, clusterName, namespace, nil, nil)
		if err != nil {
			if apierrors.IsNotFound(err) {
				errMsg := fmt.Sprintf("AWSCluster object '%s' is not present in namespace '%s' ", clusterName, namespace)
				log.V(4).Info(errMsg)
				return "", errors.New(errMsg)
			}
			return "", errors.Wrap(err, "failed to get 'AWSCluster' resource")
		}
		apiServerHost = awsCluster.Spec.ControlPlaneEndpoint.Host
		apiServerPort = awsCluster.Spec.ControlPlaneEndpoint.Port

	case "azurecluster":
		var azureCluster capzv1alpha3.AzureCluster
		err := regionalClusterClient.GetResource(&azureCluster, clusterName, namespace, nil, nil)
		if err != nil {
			if apierrors.IsNotFound(err) {
				errMsg := fmt.Sprintf("AzureCluster object '%s' is not present in namespace '%s' ", clusterName, namespace)
				log.V(4).Info(errMsg)
				return "", errors.New(errMsg)
			}
			return "", errors.Wrap(err, "failed to get 'AzureCluster' resource")
		}
		apiServerHost = azureCluster.Spec.ControlPlaneEndpoint.Host
		apiServerPort = azureCluster.Spec.ControlPlaneEndpoint.Port
	default:
		return "", errors.Errorf("failed to determine the Infra-cluster object type")
	}
	return fmt.Sprintf("https://%s:%d", apiServerHost, apiServerPort), nil
}
