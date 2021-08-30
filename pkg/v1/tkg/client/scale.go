// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package client

import (
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
func (c *TkgClient) DoScaleCluster(clusterClient clusterclient.Client, options *ScaleClusterOptions) error { // nolint:gocyclo
	isPacific, err := clusterClient.IsPacificRegionalCluster()
	if err != nil {
		return errors.Wrap(err, "error determining Tanzu Kubernetes Grid service for vSphere management cluster ")
	}
	if isPacific {
		// If pacific doesn't support the control plane scaling, it will return error
		return c.ScalePacificCluster(*options, clusterClient)
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
		// scale nodes across all machine deployments in the cluster
		workerNodeMachineDeployments, err := clusterClient.GetMDObjectForCluster(options.ClusterName, options.Namespace)
		if err != nil || len(workerNodeMachineDeployments) == 0 {
			errList = append(errList, errors.Wrapf(err, "unable to find worker node machine deployment object for cluster %s", options.ClusterName))
		} else {
			numMachineDeployments := int32(len(workerNodeMachineDeployments))
			if options.WorkerCount < numMachineDeployments {
				errList = append(errList, errors.Errorf("new worker count must be greater than or to the number of machine deployments. worker count: %d, machine deployment count: %d", options.WorkerCount, numMachineDeployments))
			}
			workersPerMD := options.WorkerCount / numMachineDeployments
			leftoverWorkers := options.WorkerCount % numMachineDeployments
			// each machine deployment gets scaled to have an approx equal number of replicas
			for i := int32(0); i < numMachineDeployments; i++ {
				workerCount := workersPerMD
				if leftoverWorkers > 0 {
					workerCount++
					leftoverWorkers--
				}
				err = clusterClient.UpdateReplicas(&workerNodeMachineDeployments[i], workerNodeMachineDeployments[i].Name, workerNodeMachineDeployments[i].Namespace, workerCount)
				if err != nil {
					errList = append(errList, errors.Wrapf(err, "unable to update worker node machine deployment replica count for cluster %s", options.ClusterName))
				}
			}
			if len(errList) == 0 {
				log.Infof("Successfully updated worker node machine deployment replica count for cluster %s", options.ClusterName)
			}
		}
	}

	if len(errList) == 0 {
		return nil
	}
	return kerrors.NewAggregate(errList)
}

// ScalePacificCluster scale TKGS cluster
func (c *TkgClient) ScalePacificCluster(options ScaleClusterOptions, clusterClient clusterclient.Client) error {
	var err error
	errList := []error{}
	// If the option specifying the targetNamespace is empty, try to detect it.
	if options.Namespace == "" {
		if options.Namespace, err = clusterClient.GetCurrentNamespace(); err != nil {
			return errors.Wrap(err, "failed to get current namespace")
		}
	}
	if options.ControlPlaneCount > 0 {
		err := clusterClient.ScalePacificClusterControlPlane(options.ClusterName, options.Namespace, "", options.ControlPlaneCount)
		if err != nil {
			errList = append(errList, errors.Wrapf(err, "unable to scale control plane for workload cluster %s", options.ClusterName))
		} else {
			log.Infof("Successfully scaled control plane for cluster %s", options.ClusterName)
		}
	}
	if options.WorkerCount > 0 {
		err := clusterClient.ScalePacificClusterWorkerNodes(options.ClusterName, options.Namespace, "", options.WorkerCount)
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
