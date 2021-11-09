// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"strings"
	"time"

	"github.com/pkg/errors"
	kerrors "k8s.io/apimachinery/pkg/util/errors"

	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/clusterclient"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/constants"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/log"
)

// ScaleClusterOptions defines options to scale tkg cluster
type ScaleClusterOptions struct {
	Kubeconfig        string
	ClusterName       string
	Namespace         string
	NodePoolName      string
	WorkerCount       int32
	ControlPlaneCount int32
}

// TKGsystemNamespace  and DefaultNamespace are constants for setting the namespaces
const (
	TKGsystemNamespace = "tkg-system"
)

// ScaleCluster scales cluster vertically
func (c *TkgClient) ScaleCluster(options ScaleClusterOptions) error {
	currentRegion, err := c.GetCurrentRegionContext()
	if err != nil {
		return errors.Wrap(err, "not a valid management cluster")
	}
	clusterclientOptions := clusterclient.Options{
		GetClientInterval: 1 * time.Second,
		GetClientTimeout:  3 * time.Second,
	}
	clusterClient, err := clusterclient.NewClient(currentRegion.SourceFilePath, currentRegion.ContextName, clusterclientOptions)
	if err != nil {
		return errors.Wrap(err, "unable to get cluster client while scaling cluster")
	}

	if currentRegion.ClusterName == options.ClusterName {
		options.Namespace = TKGsystemNamespace
	} else if options.Namespace == "" {
		options.Namespace = constants.DefaultNamespace
	}

	return c.DoScaleCluster(clusterClient, &options)
}

// DoScaleCluster performs the scale operation using the given clusterclient.Client
func (c *TkgClient) DoScaleCluster(clusterClient clusterclient.Client, options *ScaleClusterOptions) error {
	isPacific, err := clusterClient.IsPacificRegionalCluster()
	if err != nil {
		return errors.Wrap(err, "error determining Tanzu Kubernetes Grid service for vSphere management cluster ")
	}
	if isPacific {
		err := c.ValidatePacificVersionWithCLI(clusterClient)
		if err != nil {
			return err
		}
		return c.ScalePacificCluster(options, clusterClient)
	}

	errList := []error{}

	controlPlaneNode, err := clusterClient.GetKCPObjectForCluster(options.ClusterName, options.Namespace)
	if err != nil {
		errList = append(errList, errors.Wrapf(err, "unable to find control plane node object for cluster %s", options.ClusterName))
	}

	if options.ControlPlaneCount > 0 && len(errList) == 0 {
		err := clusterClient.UpdateReplicas(controlPlaneNode, controlPlaneNode.Name, controlPlaneNode.Namespace, options.ControlPlaneCount)
		if err != nil {
			errList = append(errList, errors.Wrapf(err, "unable to update control plane replica count for cluster %s", options.ClusterName))
		} else {
			log.Infof("Successfully updated control plane replica count for cluster %s", options.ClusterName)
		}
	}

	if options.WorkerCount > 0 {
		if options.NodePoolName == "" {
			errList = append(errList, c.scaleWorkersDefault(clusterClient, options)...)
		} else {
			errList = append(errList, c.scaleWorkersNodePool(clusterClient, options))
		}
	}

	if len(errList) == 0 {
		return nil
	}
	return kerrors.NewAggregate(errList)
}

// ScalePacificCluster scale TKGS cluster
func (c *TkgClient) ScalePacificCluster(options *ScaleClusterOptions, clusterClient clusterclient.Client) error {
	var err error
	errList := []error{}
	// If the option specifying the targetNamespace is empty, try to detect it.
	if options.Namespace == "" {
		if options.Namespace, err = clusterClient.GetCurrentNamespace(); err != nil {
			return errors.Wrap(err, "failed to get current namespace")
		}
	}
	if options.ControlPlaneCount > 0 {
		err := clusterClient.ScalePacificClusterControlPlane(options.ClusterName, options.Namespace, options.ControlPlaneCount)
		if err != nil {
			errList = append(errList, errors.Wrapf(err, "unable to scale control plane for workload cluster %s", options.ClusterName))
		} else {
			log.Infof("Successfully scaled control plane for cluster %s", options.ClusterName)
		}
	}
	if options.WorkerCount > 0 {
		if options.NodePoolName == "" {
			return errors.Errorf("unable to scale workers nodes for cluster %q in namespace %q , please specify the node pool name", options.ClusterName, options.Namespace)
		}
		err := c.scalePacificClusterNodePool(clusterClient, options)
		if err != nil {
			errList = append(errList, errors.Wrapf(err, "unable to scale workers nodes for workload cluster %s", options.ClusterName))
		} else {
			log.Infof("Successfully scaled workers for cluster %s", options.ClusterName)
		}
	}
	if len(errList) == 0 {
		return nil
	}
	return kerrors.NewAggregate(errList)
}

func (c *TkgClient) scaleWorkersDefault(clusterClient clusterclient.Client, options *ScaleClusterOptions) []error {
	// scale nodes across all machine deployments in the cluster
	workerNodeMachineDeployments, err := clusterClient.GetMDObjectForCluster(options.ClusterName, options.Namespace)
	if err != nil {
		return []error{errors.Wrapf(err, "error retrieving worker node machine deployment objects for cluster %s", options.ClusterName)}
	}
	if len(workerNodeMachineDeployments) == 0 {
		return []error{errors.Errorf("no machine deployments found in cluster %s", options.ClusterName)}
	}

	numMachineDeployments := int32(len(workerNodeMachineDeployments))
	if options.WorkerCount < numMachineDeployments {
		return []error{errors.Errorf("new worker count must be greater than or equal to the number of machine deployments. worker count: %d, machine deployment count: %d", options.WorkerCount, numMachineDeployments)}
	}
	workersPerMD := options.WorkerCount / numMachineDeployments
	leftoverWorkers := options.WorkerCount % numMachineDeployments
	var errList []error
	// each machine deployment gets scaled to have an approx equal number of replicas
	for i := int32(0); i < numMachineDeployments; i++ {
		workerCount := desiredWorkerCount(workersPerMD, leftoverWorkers, i)
		err := clusterClient.UpdateReplicas(&workerNodeMachineDeployments[i], workerNodeMachineDeployments[i].Name, workerNodeMachineDeployments[i].Namespace, workerCount)
		if err != nil {
			errList = append(errList, errors.Wrapf(err, "unable to update worker node replica count for machine deployment %s on cluster %s", workerNodeMachineDeployments[i].Name, options.ClusterName))
		}
	}

	if len(errList) == 0 {
		log.Infof("Successfully updated worker node machine deployment replica count for cluster %s", options.ClusterName)
	}
	return errList
}

func (c *TkgClient) scaleWorkersNodePool(clusterClient clusterclient.Client, options *ScaleClusterOptions) error {
	mdExists, err := c.mdExists(clusterClient, options)
	if err != nil {
		return err
	}
	if !mdExists {
		return errors.Errorf("Could not find node pool with name %s", options.NodePoolName)
	}

	mdOptions := prepareSetMachineDeploymentOptions(options)
	if err := DoSetMachineDeployment(clusterClient, &mdOptions); err != nil {
		return errors.Wrapf(err, "Unable to scale node pool %s", options.NodePoolName)
	}

	return nil
}

func desiredWorkerCount(workersPerMD, leftoverWorkers, i int32) int32 {
	if i < leftoverWorkers {
		return workersPerMD + 1
	}
	return workersPerMD
}

func (c *TkgClient) mdExists(clusterClient clusterclient.Client, options *ScaleClusterOptions) (bool, error) {
	getOptions := GetMachineDeploymentOptions{
		ClusterName: options.ClusterName,
		Namespace:   options.Namespace,
		Name:        options.NodePoolName,
	}
	mds, err := DoGetMachineDeployments(clusterClient, &getOptions)
	if err != nil {
		return false, errors.Wrapf(err, "Failed to get node pools for cluster %s", options.ClusterName)
	}

	npName := strings.Replace(options.NodePoolName, options.ClusterName+"-", "", -1)

	for i := range mds {
		if mds[i].Name == npName {
			return true, nil
		}
	}

	return false, nil
}

func (c *TkgClient) scalePacificClusterNodePool(clusterClient clusterclient.Client, options *ScaleClusterOptions) error {
	nodePoolExists, err := c.tkcNodePoolExists(clusterClient, options)
	if err != nil {
		return err
	}
	if !nodePoolExists {
		return errors.Errorf("could not find node pool with name %s", options.NodePoolName)
	}

	mdOptions := prepareSetMachineDeploymentOptions(options)
	if err = c.SetNodePoolsForPacificCluster(clusterClient, &mdOptions); err != nil {
		return errors.Wrapf(err, "unable to scale node pool %s", options.NodePoolName)
	}

	return nil
}

func (c *TkgClient) tkcNodePoolExists(clusterClient clusterclient.Client, options *ScaleClusterOptions) (bool, error) {
	tkc, err := clusterClient.GetPacificClusterObject(options.ClusterName, options.Namespace)
	if err != nil {
		return false, errors.Wrapf(err, "unable to get TKC object %q in namespace %q", options.ClusterName, options.Namespace)
	}

	nodePools := tkc.Spec.Topology.NodePools
	for idx := range nodePools {
		if nodePools[idx].Name == options.NodePoolName {
			return true, nil
		}
	}
	return false, nil
}

func prepareSetMachineDeploymentOptions(options *ScaleClusterOptions) SetMachineDeploymentOptions {
	return SetMachineDeploymentOptions{
		Namespace:   options.Namespace,
		ClusterName: options.ClusterName,
		NodePool: NodePool{
			Name:     options.NodePoolName,
			Replicas: &options.WorkerCount,
		},
	}
}
