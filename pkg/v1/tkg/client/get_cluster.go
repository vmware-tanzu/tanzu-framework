// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"fmt"
	"time"

	"github.com/pkg/errors"
	"github.com/yalp/jsonpath"
	crtclient "sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/clusterclient"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/log"
)

// ListTKGClustersOptions contains options supported by ListClusters
type ListTKGClustersOptions struct {
	Namespace string
	IncludeMC bool
}

// ClusterInfo defines the fields of get cluster output
type ClusterInfo struct {
	Name              string            `json:"name" yaml:"name"`
	Namespace         string            `json:"namespace" yaml:"namespace"`
	Status            string            `json:"status" yaml:"status"`
	Plan              string            `json:"plan" yaml:"plan"`
	ControlPlaneCount string            `json:"controlplane" yaml:"controlplane"`
	WorkerCount       string            `json:"workers" yaml:"workers"`
	K8sVersion        string            `json:"kubernetes" yaml:"kubernetes"`
	Roles             []string          `json:"roles" yaml:"roles"`
	Labels            map[string]string `json:"labels" yaml:"labels"`
}

// ListTKGClusters lists tkg cluster information
func (c *TkgClient) ListTKGClusters(options ListTKGClustersOptions) ([]ClusterInfo, error) {
	currentRegion, err := c.GetCurrentRegionContext()
	if err != nil {
		return nil, errors.Wrap(err, "cannot get current management cluster context")
	}
	clusterclientOptions := clusterclient.Options{
		GetClientInterval: 1 * time.Second,
		GetClientTimeout:  3 * time.Second,
	}
	regionalClusterClient, err := clusterclient.NewClient(currentRegion.SourceFilePath, currentRegion.ContextName, clusterclientOptions)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get cluster client while listing tkg clusters")
	}

	listOptions := &crtclient.ListOptions{}
	if options.Namespace != "" {
		listOptions.Namespace = options.Namespace
	}

	isPacific, err := regionalClusterClient.IsPacificRegionalCluster()
	if err != nil {
		return nil, errors.Wrap(err, "error determining 'Tanzu Kubernetes Cluster service for vSphere' management cluster")
	}
	if isPacific {
		return c.GetClusterObjectsForPacific(regionalClusterClient, "", listOptions)
	}

	return c.GetClusterObjects(regionalClusterClient, listOptions, currentRegion.ClusterName, options.IncludeMC)
}

// GetClusterObjects gets cluster objects
func (c *TkgClient) GetClusterObjects(clusterClient clusterclient.Client, listOptions *crtclient.ListOptions,
	managementClusterName string, includeMC bool) ([]ClusterInfo, error) {

	clusterInfoMap, err := getClusterObjectsMap(clusterClient, listOptions)
	if err != nil {
		return nil, errors.Wrap(err, "unable to retrieve combined cluster info")
	}

	clusters := []ClusterInfo{}

	for _, clusterInfo := range clusterInfoMap {
		if clusterInfo == nil || (clusterInfo.cluster.Name == managementClusterName && !includeMC) {
			continue
		}

		cluster := ClusterInfo{}
		cluster.Name = clusterInfo.cluster.Name
		cluster.Namespace = clusterInfo.cluster.Namespace
		cluster.Status = string(getClusterStatus(clusterInfo))
		cluster.Plan = getClusterPlan(clusterInfo)
		cluster.ControlPlaneCount = getClusterControlPlaneCount(clusterInfo)
		cluster.WorkerCount = getClusterWorkerCount(clusterInfo)
		cluster.K8sVersion = getClusterK8sVersion(clusterInfo)
		cluster.Roles = getClusterRoles(clusterInfo.cluster.Labels)
		cluster.Labels = clusterInfo.cluster.Labels
		clusters = append(clusters, cluster)
	}

	return clusters, nil
}

// GetClusterObjectsForPacific get cluster objects for TKGS clusters
func (c *TkgClient) GetClusterObjectsForPacific(clusterClient clusterclient.Client, apiVersion string, listOptions *crtclient.ListOptions) ([]ClusterInfo, error) {
	clusterInfoMap, err := getClusterObjectsMapForPacific(clusterClient, apiVersion, listOptions)
	if err != nil {
		return nil, errors.Wrap(err, "unable to retrieve combined cluster info")
	}

	clusters := []ClusterInfo{}

	for _, clusterInfo := range clusterInfoMap {
		if clusterInfo == nil {
			continue
		}

		name, _ := jsonpath.Read(clusterInfo.cluster, "$.metadata.name")
		namespace, _ := jsonpath.Read(clusterInfo.cluster, "$.metadata.namespace")
		clusterLabels, _ := jsonpath.Read(clusterInfo.cluster, "$.metadata.labels")
		clusterStatus, _ := jsonpath.Read(clusterInfo.cluster, "$.status.phase")
		k8sVersion, _ := jsonpath.Read(clusterInfo.cluster, "$.spec.distribution.version")
		desiredCPCount, _ := jsonpath.Read(clusterInfo.cluster, "$.spec.topology.controlPlane.count")
		desiredWorkerCount, _ := jsonpath.Read(clusterInfo.cluster, "$.spec.topology.workers.count")
		plan, _ := jsonpath.Read(clusterInfo.cluster, "$.metadata.labels.tkg-plan")

		getString := func(value interface{}) string {
			if value == nil {
				return ""
			}
			return fmt.Sprintf("%v", value)
		}

		clusterLabelsMap, ok := clusterLabels.(map[string]interface{})
		if !ok {
			log.V(3).Infof("error parsing cluster labels: %v for cluster '%v'", clusterLabels, name)
		}

		clusterLabelsMapString := make(map[string]string)
		for key, value := range clusterLabelsMap {
			clusterLabelsMapString[key] = value.(string)
		}

		cluster := ClusterInfo{}
		cluster.Name = getString(name)
		cluster.Namespace = getString(namespace)
		cluster.Roles = getClusterRoles(clusterLabelsMapString)
		cluster.Status = getString(clusterStatus)
		cluster.Plan = getString(plan)
		cluster.ControlPlaneCount = fmt.Sprintf("%v/%v", getRunningCPMachineCountForPacific(clusterInfo), getString(desiredCPCount))
		cluster.WorkerCount = fmt.Sprintf("%v/%v", clusterInfo.md.Status.ReadyReplicas, getString(desiredWorkerCount))
		cluster.K8sVersion = getString(k8sVersion)
		cluster.Labels = clusterLabelsMapString
		clusters = append(clusters, cluster)
	}

	return clusters, nil
}
